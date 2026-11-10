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
	roomSnapshot, err := GetRoom(client, roomID)
	if err != nil {
		return err
	}

	if roomSnapshot == nil {
		return errors.New("room not found")
	}

	roomKey := "room:" + roomID
	watchKeys := []string{roomKey}
	for _, userID := range append(append([]string{}, roomSnapshot.UserIDs...), roomSnapshot.DisconnectedUserIDs...) {
		watchKeys = append(watchKeys, "user_status:"+userID)
	}

	return client.Watch(ctx, func(tx *redis.Tx) error {
		roomJSON, err := tx.Get(ctx, roomKey).Result()
		if err != nil {
			return err
		}
		var room entity.Room
		if err := json.Unmarshal([]byte(roomJSON), &room); err != nil {
			return err
		}
		if !sameIDs(roomSnapshot.UserIDs, room.UserIDs) || !sameIDs(roomSnapshot.DisconnectedUserIDs, room.DisconnectedUserIDs) {
			return redis.TxFailedErr
		}

		remainingUsers := make([]string, 0, len(room.UserIDs))
		disconnectedUsers := make(map[string]struct{})
		userStatuses := make(map[string][]byte, len(room.UserIDs))
		for _, userID := range room.UserIDs {
			status := entity.UserStatus{UserID: userID, Status: entity.UserStatusInRoom, RoomID: roomID}
			statusJSON, err := tx.Get(ctx, "user_status:"+userID).Result()
			if err == nil {
				if err := json.Unmarshal([]byte(statusJSON), &status); err != nil {
					return err
				}
			} else if err != redis.Nil {
				return err
			}
			if containsID(room.DisconnectedUserIDs, userID) || status.ConnectionStatus == entity.UserConnectionStatusDisconnected {
				disconnectedUsers[userID] = struct{}{}
				continue
			}
			status.Status = entity.UserStatusInRoom
			status.RoomID = roomID
			encodedStatus, err := json.Marshal(status)
			if err != nil {
				return err
			}
			userStatuses[userID] = encodedStatus
			remainingUsers = append(remainingUsers, userID)
		}
		for _, userID := range room.DisconnectedUserIDs {
			disconnectedUsers[userID] = struct{}{}
		}

		room.UserIDs = remainingUsers
		room.DisconnectedUserIDs = nil
		room.Status = entity.RoomStatusWaiting
		room.Prepared = []string{}
		if len(room.UserIDs) > 0 && !containsID(room.UserIDs, room.HosterID) {
			var updatedRoom *entity.Room
			updatedRoom, err = RandomHoster(&room)
			if err != nil {
				return err
			}
			room = *updatedRoom
		}
		updatedRoomJSON, err := json.Marshal(room)
		if err != nil {
			return err
		}

		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			if len(room.UserIDs) == 0 {
				pipe.Del(ctx, roomKey)
			} else {
				pipe.Set(ctx, roomKey, updatedRoomJSON, 0)
			}
			for userID, statusJSON := range userStatuses {
				pipe.Set(ctx, "user_status:"+userID, statusJSON, 0)
			}
			for userID := range disconnectedUsers {
				pipe.Del(ctx, "user_status:"+userID)
			}
			return nil
		})
		return err
	}, watchKeys...)
}

func sameIDs(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func connectionStatusFor(disconnectedUserIDs []string, userID string) string {
	if containsID(disconnectedUserIDs, userID) {
		return entity.UserConnectionStatusDisconnected
	}
	return ""
}
