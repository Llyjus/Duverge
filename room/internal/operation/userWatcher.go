package operation

import (
	"context"
	"encoding/json"
	"errors"

	"room/internal/entity"

	redis "github.com/redis/go-redis/v9"
)

func ExitRoom(
	client *redis.Client,
	roomID string,
	userId string,
) error {
	ctx := context.Background()

	userStatusKey := "user_status:" + userId
	roomKey := "room:" + roomID

	return client.Watch(ctx, func(tx *redis.Tx) error {

		// Read current user status
		userStatusJSON, err := tx.Get(ctx, userStatusKey).Result()
		if err != nil {
			return err
		}

		// Read current room
		roomJSON, err := tx.Get(ctx, roomKey).Result()
		if err != nil {
			return err
		}

		var userStatus entity.UserStatus
		if err := json.Unmarshal(
			[]byte(userStatusJSON),
			&userStatus,
		); err != nil {
			return err
		}

		var room entity.Room
		if err := json.Unmarshal(
			[]byte(roomJSON),
			&room,
		); err != nil {
			return err
		}

		// Make sure user is still in room
		if userStatus.Status != entity.UserStatusInRoom {
			return errors.New("user is not in room")
		}

		// Delete user
		if err := DeleteUser(&room, userId); err != nil {
			return err
		}

		// Change status
		userStatus.Status = entity.UserStatusInLobby

		newStatusJSON, err := json.Marshal(userStatus)
		if err != nil {
			return err
		}

		newRoomJSON, err := json.Marshal(room)
		if err != nil {
			return err
		}

		// Commit
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, userStatusKey, newStatusJSON, 0)
			pipe.Set(ctx, roomKey, newRoomJSON, 0)

			return nil
		})

		return err

	}, userStatusKey, roomKey)
}

func JoinRoom(
	client *redis.Client,
	roomID string,
	userId string,
) error {
	ctx := context.Background()
	userStatusKey := "user_status:" + userId
	roomKey := "room:" + roomID

	return client.Watch(ctx, func(tx *redis.Tx) error {

		// Read current user status
		userStatusJSON, err := tx.Get(ctx, userStatusKey).Result()
		if err != nil {
			return err
		}

		// Read current room
		roomJSON, err := tx.Get(ctx, roomKey).Result()
		if err != nil {
			return err
		}

		var userStatus entity.UserStatus
		if err := json.Unmarshal(
			[]byte(userStatusJSON),
			&userStatus,
		); err != nil {
			return err
		}

		var room entity.Room
		if err := json.Unmarshal(
			[]byte(roomJSON),
			&room,
		); err != nil {
			return err
		}

		// Make sure user is in lobby
		if userStatus.Status != entity.UserStatusInLobby {
			return errors.New("user is already in a room")
		}

		// Make sure room is still waiting
		if room.Status != entity.RoomStatusWaiting {
			return errors.New("room has already started or deleted")
		}

		// Make sure room is not full
		if int64(len(room.UserIDs)) >= room.MaximumSize {
			return errors.New("room is full")
		}

		// Add user to room
		room.UserIDs = append(room.UserIDs, userId)

		// Change user status
		userStatus.Status = entity.UserStatusInRoom

		// Marshal
		newRoomJSON, err := json.Marshal(room)
		if err != nil {
			return err
		}

		newStatusJSON, err := json.Marshal(userStatus)
		if err != nil {
			return err
		}

		// Commit
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(
				ctx,
				userStatusKey,
				newStatusJSON,
				0,
			)

			pipe.Set(
				ctx,
				roomKey,
				newRoomJSON,
				0,
			)

			return nil
		})

		return err

	}, userStatusKey, roomKey)
}

func SetPreparation(
	client *redis.Client,
	roomID string,
	userId string,
	targetIfPrepared bool,
) error {
	ctx := context.Background()

	userStatusKey := "user_status:" + userId
	roomKey := "room:" + roomID

	return client.Watch(ctx, func(tx *redis.Tx) error {

		// Read user status
		userStatusJSON, err := tx.Get(ctx, userStatusKey).Result()
		if err != nil {
			return err
		}

		// Read room
		roomJSON, err := tx.Get(ctx, roomKey).Result()
		if err != nil {
			return err
		}

		var userStatus entity.UserStatus
		if err := json.Unmarshal([]byte(userStatusJSON), &userStatus); err != nil {
			return err
		}

		var room entity.Room
		if err := json.Unmarshal([]byte(roomJSON), &room); err != nil {
			return err
		}

		// User must be in room
		if userStatus.Status != entity.UserStatusInRoom {
			return errors.New("user is not in room")
		}

		// Room must still be waiting
		if room.Status != entity.RoomStatusWaiting {
			return errors.New("game has already started")
		}

		// Set preparation status
		if targetIfPrepared {
			// Already prepared
			for _, id := range room.Prepared {
				if id == userId {
					return nil
				}
			}

			room.Prepared = append(room.Prepared, userId)

		} else {
			// Cancel preparation
			for i, id := range room.Prepared {
				if id == userId {
					room.Prepared = append(
						room.Prepared[:i],
						room.Prepared[i+1:]...,
					)
					break
				}
			}
		}

		// Marshal updated room
		newRoomJSON, err := json.Marshal(room)
		if err != nil {
			return err
		}

		// Commit
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
			pipe.Set(ctx, roomKey, newRoomJSON, 0)
			return nil
		})

		return err

	}, userStatusKey, roomKey)
}
