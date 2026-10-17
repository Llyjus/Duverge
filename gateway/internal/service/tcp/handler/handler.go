package handler

import (
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	auth "gateway/internal/service/tcp/auth"
	chat "gateway/internal/service/tcp/chat"
	connection "gateway/internal/service/tcp/connection"
	protocol "gateway/internal/service/tcp/protocol"
)

func HandleConnection(
	session *connection.Session,
	manager *connection.ConnectionManager,
	redisClient *redis.Client,
) {
	defer session.Close()

	for {
		// read the first message from the connection
		body, err := connection.ReadMessage(session.Conn)
		if err != nil {
			fmt.Printf("[TCP] failed to read message: %v\n", err)

			// Send a message to the client
			closeMessage := []byte("Connection closed by server")
			if writeErr := connection.WriteMessage(session.Conn, closeMessage); writeErr != nil {
				fmt.Printf("[TCP] failed to send close message: %v\n", writeErr)
			}
			return
		}

		// decode the message
		var msg protocol.Message
		if err := json.Unmarshal(body, &msg); err != nil {
			fmt.Printf("[TCP] invalid message: %v\n", err)
			return
		}

		if msg.Version != protocol.Version {
			fmt.Printf("[TCP] unsupported message version: %v\n", msg.Version)
			return
		}

		// check authentication status
		if !session.AuthStatus {
			if msg.Type != "auth" {
				fmt.Println("[TCP] authentication required")
				return
			}

			if err := auth.Authenticate(
				session,
				manager,
				redisClient,
				msg.Data,
			); err != nil {
				fmt.Printf("[TCP] authentication failed: %v\n", err)
				return
			}

			defer manager.Remove(session.UserID, session)

			continue
		}

		// classify the message
		switch msg.Type {
		case "auth":
			fmt.Println("[TCP] connection is already authenticated")

		case "chat":
			err := chat.HandleChat(
				session.UserID,
				manager,
				redisClient,
				msg.Data,
			)

			// error handling for chat message
			if err != nil {
				fmt.Printf("[TCP] failed to handle chat message: %v\n", err)

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
			responseData, err := protocol.CapsulateSuccessResponse()
			if err != nil {
				fmt.Printf("[TCP] failed to send chat response: %v\n", err)
				return
			}

			if writeErr := connection.WriteMessage(session.Conn, responseData); writeErr != nil {
				fmt.Printf("[TCP] failed to send chat response: %v\n", writeErr)
				return
			}

			fmt.Printf("[TCP] received chat message from %s\n", session.UserID)

		case "game":
			// TODO: handle the game message
			fmt.Printf("[TCP] received game message from %s\n", session.UserID)

		default:
			fmt.Printf("[TCP] unsupported message type: %v\n", msg.Type)

		}
	}
}
