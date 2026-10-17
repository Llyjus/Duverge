package auth

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"
)

// session creation and authentication functions
func SessionCreation(
	redisClient *redis.Client,
	sessionID string,
	accountID string,
) (string, error) {

	// Check if the session already exists
	accountID = "session_account:" + accountID
	session, err := redisClient.Get(context.Background(), accountID).Result()
	if err != redis.Nil && err != nil {
		return "", err
	}

	if session != "" {
		// Delete the existing session and account ID from Redis
		err = redisClient.Del(context.Background(), session).Err()
		err = redisClient.Del(context.Background(), accountID).Err()
		if err != nil {
			return "", err
		}
	}

	// add the pre to sessionID
	sessionID = "session:" + sessionID

	// set the session ID and account ID in Redis with an expiration time of 7 days
	err = redisClient.Set(
		context.Background(),
		sessionID,
		accountID,
		7*24*time.Hour,
	).Err()

	if err != nil {
		return "", err
	}

	err = redisClient.Set(
		context.Background(),
		accountID,
		sessionID,
		7*24*time.Hour,
	).Err()

	if err != nil {
		return "", err
	}
	sessionID = strings.TrimPrefix(sessionID, "session:")

	return sessionID, nil
}

func SessionAuthentication(
	redisClient *redis.Client,
	session string,
) (string, error) {

	session = "session:" + session

	accountID, err := redisClient.Get(context.Background(), session).Result()
	if err != nil && err != redis.Nil {
		return "", err
	}
	accountID = strings.TrimPrefix(accountID, "session_account:")
	return accountID, nil
}

func GetUserIDFromSessionID(
	redisClient *redis.Client,
	ctx context.Context,
	sessionID string,
) (string, error) {
	sessionKey := "session:" + sessionID
	accountID, err := redisClient.Get(ctx, sessionKey).Result()
	if err != nil {
		return "", err
	}
	accountID = strings.TrimPrefix(accountID, "session_account:")
	return accountID, nil
}

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
