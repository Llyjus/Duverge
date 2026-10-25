package chat

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	connection "gateway/internal/service/tcp/connection"
	"gateway/internal/service/tcp/protocol"
)

func HandleChat(
	userID string,
	manager *connection.ConnectionManager,
	redisClient *redis.Client,
	ctx context.Context,
	data []byte,
) error {
	var msg protocol.ChatMessage

	// Decode the chat message
	if err := json.Unmarshal(data, &msg); err != nil {
		return fmt.Errorf(
			"%w: %v",
			protocol.ErrInvalidMessage,
			err)

	}

	// Validate the chat message
	if msg.Type == "" || msg.Data == nil {
		return fmt.Errorf(
			"%w: missing required fields",
			protocol.ErrInvalidMessage)
	}

	switch msg.Type {
	case "room":
		var roomMsg protocol.RoomChatMessage
		if err := json.Unmarshal(msg.Data, &roomMsg); err != nil {
			return fmt.Errorf(
				"%w: %v",
				protocol.ErrInvalidMessage,
				err)
		}

		if roomMsg.RoomID == "" || roomMsg.Content == "" {
			return fmt.Errorf(
				"%w: missing required fields",
				protocol.ErrInvalidMessage)
		}

		return roomChat(
			redisClient,
			manager,
			ctx,
			roomMsg.RoomID,
			userID,
			roomMsg.Type,
			roomMsg.Content)

	case "private":
		var privateMsg protocol.PrivateChatMessage
		if err := json.Unmarshal(msg.Data, &privateMsg); err != nil {
			return fmt.Errorf(
				"%w: %v",
				protocol.ErrInvalidMessage,
				err)
		}

		if privateMsg.To == "" || privateMsg.Content == "" {
			return fmt.Errorf(
				"%w: missing required fields",
				protocol.ErrInvalidMessage)
		}

		return privateChat(
			manager,
			userID,
			privateMsg.To,
			privateMsg.Type,
			privateMsg.Content)

	default:
		return fmt.Errorf(
			"%w: unsupported chat type: %s",
			protocol.ErrInvalidMessage,
			msg.Type)
	}

}
