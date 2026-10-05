package operation

import (
	"context"
	"encoding/json"

	"room/internal/entity"

	redis "github.com/redis/go-redis/v9"
)

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

func CheckRoomIsEmpty(
	room *entity.Room,
) (bool, error) {
	return len(room.UserIDs) == 0, nil
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
