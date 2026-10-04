package operation

import (
	"context"
	"encoding/json"

	"room/internal/entity"

	redis "github.com/redis/go-redis/v9"
)

func CheckUserStatus(
	client *redis.Client,
	userId string,
) (string, error) {
	// Check user status
	userStatusKey := "user_status:" + userId
	userStatusJSON, err := client.Get(context.Background(), userStatusKey).Result()
	if err == redis.Nil {
		// User status not found(should not be triggered)
		return "", nil
	} else if err != nil {
		return "", err
	}

	// Unmarshal user status
	var userStatus entity.UserStatus
	err = json.Unmarshal([]byte(userStatusJSON), &userStatus)
	if err != nil {
		return "", err
	}

	return userStatus.Status, nil
}

func UpdateUserStatus(
	client *redis.Client,
	userId string,
	status string,
) error {
	// Update user status
	userStatusKey := "user_status:" + userId
	userStatus := entity.UserStatus{
		UserID: userId,
		Status: status,
	}
	userStatusJSON, err := json.Marshal(userStatus)
	if err != nil {
		return err
	}

	return client.Set(context.Background(),
		userStatusKey,
		userStatusJSON,
		0).Err()
}
