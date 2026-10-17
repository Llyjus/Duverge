package chat

import (
	"context"

	"github.com/redis/go-redis/v9"

	auth "gateway/internal/service/auth"
	connection "gateway/internal/service/tcp/connection"
	protocol "gateway/internal/service/tcp/protocol"

	"github.com/google/uuid"
)

// roomChat sends a chat message to all members of a room.
func roomChat(
	redisClient *redis.Client,
	manager *connection.ConnectionManager,
	roomID string,
	userID string,
	message_type string,
	content string,
) error {
	ctx := context.Background()

	// get the room members
	getRoomMembers, err := getRoomMembers(
		redisClient,
		roomID,
		ctx)
	if err != nil {
		return err
	}

	// send the message to each member
	for _, memberID := range getRoomMembers {
		session, ok := manager.Get(memberID)
		if !ok {
			continue
		}

		messageID := uuid.NewString()

		chatPostBytes, err := protocol.CapsulateChatMessage(
			messageID,
			userID,
			memberID,
			message_type,
			content,
		)
		if err != nil {
			return err
		}

		err = connection.WriteMessage(session.Conn, chatPostBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

func getRoomMembers(
	redisClient *redis.Client,
	roomID string,
	ctx context.Context,
) ([]string, error) {

	if userIDs, err := auth.GetUserIDsFromRoomID(
		redisClient,
		ctx,
		roomID,
	); err != nil {
		return nil, err
	} else {
		return userIDs, nil
	}
}
