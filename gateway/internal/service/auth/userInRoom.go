package auth

import (
	"context"
	"encoding/json"

	redis "github.com/redis/go-redis/v9"
)

// get user IDs from room ID for chat message broadcasting
func GetUserIDsFromRoomID(
	redisClient *redis.Client,
	ctx context.Context,
	roomID string,
) ([]string, error) {
	roomKey := "room:" + roomID
	roomJSON, err := redisClient.Get(ctx, roomKey).Result()
	if err != nil {
		return nil, err
	}

	var room struct {
		UserIDs []string `json:"userIds"`
	}
	if err := json.Unmarshal([]byte(roomJSON), &room); err != nil {
		return nil, err
	}

	return room.UserIDs, nil
}
