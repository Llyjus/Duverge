package service

import (
	"context"
	"encoding/json"
	"room/internal/operation"

	redis "github.com/redis/go-redis/v9"
)

func AddUserToRoom(
	client *redis.Client,
	roomId string,
	userId string,
) error {
	roomKey := "room:" + roomId
	room, err := GetRoom(client, roomKey)
	if err != nil {
		return err
	}
	if room == nil {
		return nil // Room not found
	}

	// Check room status
	if room.Status != "waiting" {
		return nil // Room is not in a state to accept new users
	}
	// // Check user status
	UserStatus, err := operation.CheckUserStatus(client, userId)
	if err != nil {
		return err
	}
	// User is not in the lobby
	if UserStatus != "in_lobby" {
		return nil // User is not in the lobby
	}

	// Add the user to the room
	room.UserIDs = append(room.UserIDs, userId)

	// Update the user's status
	err = operation.UpdateUserStatus(client, userId, "in_room")
	if err != nil {
		return err
	}

	// Update the room in Redis
	roomJSON, err := json.Marshal(room)
	if err != nil {
		return err
	}
	return client.Set(context.Background(),
		roomKey,
		roomJSON,
		0).Err()
}
