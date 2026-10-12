package operation

import (
	"context"
	"encoding/json"
	"errors"

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

func EndGame(
	client *redis.Client,
	roomID string,
) error {
	ctx := context.Background()

	room, err := GetRoom(client, roomID)
	if err != nil {
		return err
	}

	if room == nil {
		return errors.New("room not found")
	}

	room.Status = entity.RoomStatusWaiting
	room.Prepared = []string{}

	roomJSON, err := json.Marshal(room)
	if err != nil {
		return err
	}

	_, err = client.TxPipelined(ctx, func(pipe redis.Pipeliner) error {

		pipe.Set(
			ctx,
			"room:"+roomID,
			roomJSON,
			0,
		)

		for _, userID := range room.UserIDs {
			userStatus := entity.UserStatus{
				UserID: userID,
				Status: entity.UserStatusInRoom,
			}

			userStatusJSON, err := json.Marshal(userStatus)
			if err != nil {
				return err
			}

			pipe.Set(
				ctx,
				"user_status:"+userID,
				userStatusJSON,
				0,
			)
		}

		return nil
	})

	return err
}
