package operation

import (
	"context"
	"encoding/json"

	"room/internal/entity"
	"room/internal/function"

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

func GetUserStatus(ctx context.Context, client *redis.Client, userId string) (*entity.UserStatus, error) {
	userStatusJSON, err := client.Get(ctx, "user_status:"+userId).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var userStatus entity.UserStatus
	if err := json.Unmarshal([]byte(userStatusJSON), &userStatus); err != nil {
		return nil, err
	}
	return &userStatus, nil
}

func IsCurrentSession(ctx context.Context, client *redis.Client, userID, sessionID string) (bool, error) {
	currentSession, err := client.Get(ctx, "session_account:"+userID).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return currentSession == "session:"+sessionID, nil
}

func DeleteUserStatusIfCurrentSession(ctx context.Context, client *redis.Client, userID, sessionID string) error {
	userStatusKey := "user_status:" + userID
	sessionKey := "session_account:" + userID
	return client.Watch(ctx, func(tx *redis.Tx) error {
		if err := ensureCurrentSession(ctx, tx, userID, sessionID); err != nil {
			return err
		}
		_, err := tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Del(ctx, userStatusKey)
			return nil
		})
		return err
	}, userStatusKey, sessionKey)
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

func CheckHoster(
	room *entity.Room,
	userId string,
) (bool, error) {
	return room.HosterID == userId, nil
}

func RandomHoster(
	room *entity.Room,
) (*entity.Room, error) {

	// Randomly select a new hoster from the remaining users
	newHosterID := function.RandomChoice(room.UserIDs)
	room.HosterID = newHosterID

	return room, nil
}

func DeleteUser(
	room *entity.Room,
	userId string,
) error {

	// Remove the user from the room
	for i, id := range room.UserIDs {
		if id == userId {
			room.UserIDs = append(room.UserIDs[:i], room.UserIDs[i+1:]...)
			break
		}
	}

	return nil
}
