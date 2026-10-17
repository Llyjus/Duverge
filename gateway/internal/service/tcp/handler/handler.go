package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"

	auth "gateway/internal/service/tcp/auth"
	chat "gateway/internal/service/tcp/chat"
	connection "gateway/internal/service/tcp/connection"
	heartbeat "gateway/internal/service/tcp/heartbeat"
	protocol "gateway/internal/service/tcp/protocol"
)

func HandleConnection(
	session *connection.Session,
	manager *connection.ConnectionManager,
	redisClient *redis.Client,
) {
	ctx := context.Background()
	defer session.Close()

	for {
		// read the first message from the connection
		body, err := connection.ReadMessage(session.Conn)
		if err != nil {
			log.Printf("[TCP] failed to read message: %v\n", err)

			// Send a message to the client
			closeMessage, err := protocol.CapsulateErrorResponse(
				protocol.ErrCodeInternal,
				"服务器连接错误，请稍后重试",
			)
			if err != nil {
				log.Printf("[TCP] failed to capsulate close message: %v\n", err)
				return
			}

			if writeErr := connection.WriteMessage(session.Conn, closeMessage); writeErr != nil {
				log.Printf("[TCP] failed to send close message: %v\n", writeErr)
			}
			return
		}

		// decode the message
		var msg protocol.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			log.Printf("[TCP] invalid message: %v\n", err)

			responseData, err := protocol.CapsulateErrorResponse(
				protocol.ErrCodeInvalidMessage,
				"非法消息格式",
			)
			if err != nil {
				log.Printf("[TCP] failed to capsulate invalid message response: %v\n", err)
				return
			}
			if writeErr := connection.WriteMessage(session.Conn, responseData); writeErr != nil {
				log.Printf("[TCP] failed to send invalid message response: %v\n", writeErr)
				return
			}
			continue
		}

		if msg.Version != protocol.Version {
			log.Printf("[TCP] unsupported message version: %v\n", msg.Version)

			responseData, err := protocol.CapsulateErrorResponse(
				protocol.ErrCodeVersionMismatch,
				"不支持的消息版本",
			)
			if err != nil {
				log.Printf("[TCP] failed to capsulate version mismatch response: %v\n", err)
				return
			}
			if writeErr := connection.WriteMessage(session.Conn, responseData); writeErr != nil {
				log.Printf("[TCP] failed to send version mismatch response: %v\n", writeErr)
			}
			return
		}

		// check authentication status
		if !session.AuthStatus {
			if msg.Type != "auth" {
				ErrorResponseData, err := protocol.CapsulateErrorResponse(
					protocol.ErrCodeUnauthenticated,
					"请先进行身份验证",
				)
				if err != nil {
					log.Printf("[TCP] failed to capsulate unauthenticated error response: %v\n", err)
					return
				}
				if writeErr := connection.WriteMessage(session.Conn, ErrorResponseData); writeErr != nil {
					log.Printf("[TCP] failed to send unauthenticated error response: %v\n", writeErr)
					return
				}
				continue
			}

			if err := auth.Authenticate(
				session,
				manager,
				redisClient,
				ctx,
				msg.Data,
			); err != nil {
				log.Printf("[TCP] authentication failed: %v\n", err)
				return
			}

			// send a success response back to the sender
			responseData, err := protocol.CapsulateSuccessResponse(msg.RequestID)
			if err != nil {
				log.Printf("[TCP] failed to send chat response: %v\n", err)
				return
			}

			if writeErr := connection.WriteMessage(session.Conn, responseData); writeErr != nil {
				log.Printf("[TCP] failed to send chat response: %v\n", writeErr)
				return
			}

			defer manager.Remove(session.UserID, session)

			continue
		}

		// classify the message
		switch msg.Type {
		case "auth":
			log.Println("[TCP] connection is already authenticated")
			errorResponseData, err := protocol.CapsulateErrorResponse(
				protocol.ErrCodeForbidden,
				"连接已认证，禁止重复认证",
			)
			if err != nil {
				log.Printf("[TCP] failed to capsulate auth error response: %v\n", err)
				return
			}
			if writeErr := connection.WriteMessage(session.Conn, errorResponseData); writeErr != nil {
				log.Printf("[TCP] failed to send auth error response: %v\n", writeErr)
				return
			}

		case "chat":

			log.Printf("[TCP] received chat message from %s\n", session.UserID)

			err := chat.HandleChat(
				session.UserID,
				manager,
				redisClient,
				msg.Data,
			)

			// error handling for chat message
			if err != nil {
				log.Printf("[TCP] failed to handle chat message: %v\n", err)

				// classify the error and capsulate the error response
				errorCode, detail := mapChatError(err)

				errorResponseData, err := protocol.CapsulateErrorResponse(
					errorCode,
					detail,
				)
				if err != nil {
					fmt.Printf("[TCP] failed to capsulate error response: %v\n", err)
					return
				}

				if writeErr := connection.WriteMessage(session.Conn, errorResponseData); writeErr != nil {
					fmt.Printf("[TCP] failed to send error response: %v\n", writeErr)
					return
				}

				//continue to next message
				continue
			}

			// send a success response back to the sender
			responseData, err := protocol.CapsulateSuccessResponse(msg.RequestID)
			if err != nil {
				log.Printf("[TCP] failed to send chat response: %v\n", err)
				return
			}

			if writeErr := connection.WriteMessage(session.Conn, responseData); writeErr != nil {
				log.Printf("[TCP] failed to send chat response: %v\n", writeErr)
				return
			}

		case "heartbeat":

			log.Printf("[TCP] received heartbeat from %s\n", session.UserID)

			err := heartbeat.HandleHeartbeat(
				session,
				redisClient,
				ctx,
			)
			// error handling for heartbeat message
			if err != nil {
				log.Printf("[TCP] failed to handle heartbeat message: %v\n", err)

				errorResponseData, err := protocol.CapsulateErrorResponse(
					protocol.ErrCodeInternal,
					"内部服务器错误",
				)
				if err != nil {
					log.Printf("[TCP] failed to capsulate error response: %v\n", err)
					return
				}

				if writeErr := connection.WriteMessage(session.Conn, errorResponseData); writeErr != nil {
					log.Printf("[TCP] failed to send error response: %v\n", writeErr)
					return
				}

				//continue to next message
				continue
			}

			// send a success response back to the sender
			responseData, err := protocol.CapsulateSuccessResponse(msg.RequestID)
			if err != nil {
				log.Printf("[TCP] failed to send heartbeat response: %v\n", err)
				return
			}

			if writeErr := connection.WriteMessage(session.Conn, responseData); writeErr != nil {
				log.Printf("[TCP] failed to send heartbeat response: %v\n", writeErr)
				return
			}

		case "game":
			// TODO: handle the game message
			log.Printf("[TCP] received game message from %s\n", session.UserID)

		default:
			log.Printf("[TCP] unsupported message type: %v\n", msg.Type)
			errorResponseData, err := protocol.CapsulateErrorResponse(
				protocol.ErrCodeInvalidMessage,
				"不支持的消息类型",
			)
			if err != nil {
				log.Printf("[TCP] failed to capsulate type error response: %v\n", err)
				return
			}
			if writeErr := connection.WriteMessage(session.Conn, errorResponseData); writeErr != nil {
				log.Printf("[TCP] failed to send type error response: %v\n", writeErr)
				return
			}

		}
	}
}
