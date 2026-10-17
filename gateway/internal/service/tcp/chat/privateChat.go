// TODO: insert the message into the database for persistence

package chat

import (
	"fmt"

	connection "gateway/internal/service/tcp/connection"
	protocol "gateway/internal/service/tcp/protocol"

	"github.com/google/uuid"
)

// roomChat sends a chat message to all members of a room.
func privateChat(
	manager *connection.ConnectionManager,
	userID string,
	targetUserID string,
	message_type string,
	content string,
) error {
	session, ok := manager.Get(targetUserID)
	if !ok {
		return fmt.Errorf("session not found for user: %s", protocol.ErrNotFound)
	}

	messageID := uuid.NewString()

	// capsulate the message to send
	chatPostBytes, err := protocol.CapsulateChatMessage(
		messageID,
		userID,
		targetUserID,
		message_type,
		content,
	)
	if err != nil {
		return err
	}

	// send the message to the target user
	err = connection.WriteMessage(session.Conn, chatPostBytes)
	if err != nil {
		return err
	}

	// TODO: insert the message into the database for persistence

	return nil
}
