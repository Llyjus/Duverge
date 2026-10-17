package auth

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// Create user status and manage it after successful authentication
func UserStatusCreation(
	redisClient *redis.Client,
	ctx context.Context,
	userID string,
) error {
	statusKey := "user_status:" + userID
	return redisClient.Set(ctx, statusKey, "in_lobby", 30*time.Second).Err()
}

func RenewUserStatus(
	redisClient *redis.Client,
	ctx context.Context,
	userID string,
) error {
	statusKey := "user_status:" + userID
	return redisClient.Expire(ctx, statusKey, 30*time.Second).Err()
}
