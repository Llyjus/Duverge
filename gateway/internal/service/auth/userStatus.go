package auth

import (
	"context"
	"encoding/json"
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
	status := struct {
		UserID string `json:"user_id"`
		Status string `json:"status"`
	}{
		UserID: userID,
		Status: "in_lobby",
	}
	statusJSON, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return redisClient.Set(ctx, statusKey, statusJSON, 30*time.Second).Err()
}

func RenewUserStatus(
	redisClient *redis.Client,
	ctx context.Context,
	userID string,
) error {
	statusKey := "user_status:" + userID
	return redisClient.Expire(ctx, statusKey, 30*time.Second).Err()
}
