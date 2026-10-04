package service

import (
	"context"
	"encoding/json"

	"room/internal/entity"
	"room/internal/function"

	redis "github.com/redis/go-redis/v9"
)

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
		Status:      "waiting",
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

func GetRoom(
	client *redis.Client,
	roomId string,
) (
	*entity.Room,
	error) {
	roomKey := "room:" + roomId

	// Get the room from Redis
	roomJSON, err := client.Get(context.Background(), roomKey).Result()
	if err == redis.Nil {
		return nil, nil // Room not found
	} else if err != nil {
		return nil, err
	}

	// Unmarshal the JSON into a Room struct
	var room entity.Room
	err = json.Unmarshal([]byte(roomJSON), &room)
	if err != nil {
		return nil, err
	}

	return &room, nil
}

func UpdateRoom(
	client *redis.Client,
	room *entity.Room,
) error {
	roomKey := "room:" + room.RoomID

	roomJSON, err := json.Marshal(room)
	if err != nil {
		return err
	}

	return client.Set(
		context.Background(),
		roomKey,
		roomJSON,
		0,
	).Err()
}

func DeleteRoom(
	client *redis.Client,
	roomId string,
) error {
	roomKey := "room:" + roomId

	return client.Del(context.Background(), roomKey).Err()
}
