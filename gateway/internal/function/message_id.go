package function

import "github.com/google/uuid"

// GenerateMessageID returns a new identifier for one logical message.
func GenerateMessageID() string {
	return uuid.NewString()
}
