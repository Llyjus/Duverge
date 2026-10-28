package function

import (
	"context"
	"fmt"
	internalerrors "gateway/internal/errors"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// GenerateMessageID returns a UUID after atomically reserving its Redis key.
// If the UUID already exists, it generates another and checks again.
func GenerateMessageID(ctx context.Context, redisClient *redis.Client) (string, error) {
	for {
		messageID := uuid.NewString()
		created, err := redisClient.SetNX(ctx, "message_id:"+messageID, "uncompleted", 0).Result()
		if err != nil {
			return "", fmt.Errorf("%w: reserve message ID: %w", internalerrors.ErrRedisUnavailable, err)
		}
		if created {
			return messageID, nil
		}
	}
}
