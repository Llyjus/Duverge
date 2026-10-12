package service

import (
	"context"
	"encoding/json"

	"room/internal/entity"
	"room/internal/function"
	"room/internal/operation"

	redis "github.com/redis/go-redis/v9"
)

func GetRoomUsers(
	client *redis.Client,
	roomId string,
) ([]string, error) {
	room, err := operation.GetRoom(client, roomId)
	if err != nil {
		return nil, err
	}

	if room == nil {
		return nil, nil
	}

	return room.UserIDs, nil
}

func CreateRoom(
	client *redis.Client,
	HosterID string,
) (roomId string,
	err error) {

	var roomCode string
	// Create a new room entity
	room := &entity.Room{
		RoomID:      roomCode,
		MaximumSize: 4,
		UserIDs:     []string{HosterID},
		HosterID:    HosterID,
		Status:      entity.RoomStatusWaiting,
	}

	for {
		// Generate a room code
		roomCode = function.GenerateRoomCode()
		roomKey := "room:" + roomCode
		room.RoomID = roomCode

		// Store the room in Redis;
		// Use SetNX to ensure that the room is only created if it doesn't already exist
		roomJSON, err := json.Marshal(room)
		if err != nil {
			return "", err
		}
		res, err := client.SetNX(context.Background(), roomKey, roomJSON, 0).Result()
		if err != nil {
			return "", err
		}
		if res {
			return roomCode, nil
		}
	}

}

func StartRoomGame(
	client *redis.Client,
	roomID string,
	userId string,
) error {
	var err error

	for i := 0; i < 3; i++ {
		err = operation.StartGame(client, roomID, userId)

		if err == nil {
			return nil
		}

		if err != redis.TxFailedErr {
			return err
		}
	}

	return err
}

func EndRoomGame(
	client *redis.Client,
	roomID string,
) error {
	var err error

	for i := 0; i < 3; i++ {
		err = operation.EndGame(client, roomID)

		if err == nil {
			return nil
		}
	}

	return err
}
