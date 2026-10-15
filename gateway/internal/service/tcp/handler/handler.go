package handler

import (
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	connection "gateway/internal/service/tcp/connection"
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
			return
		}

		// decode the message
		var msg Message
		if err := json.Unmarshal(body, &msg); err != nil {
			fmt.Printf("[TCP] invalid message: %v\n", err)
			return
		}

		if msg.Version != "1.0" {
			fmt.Printf("[TCP] unsupported message version: %v\n", msg.Version)
			return
		}

		// check authentication status
		if !session.AuthStatus {
			if msg.Type != "auth" {
				fmt.Println("[TCP] authentication required")
				return
			}

			if err := authenticate(
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
			fmt.Printf("[TCP] received chat message from %s\n", session.UserID)

		case "game":
			// TODO: handle the game message
			fmt.Printf("[TCP] received game message from %s\n", session.UserID)

		default:
			fmt.Printf("[TCP] unsupported message type: %v\n", msg.Type)

		}
	}
}
