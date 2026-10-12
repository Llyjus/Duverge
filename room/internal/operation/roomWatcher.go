package operation

import (
	"context"
	"encoding/json"
	"errors"

	"room/internal/entity"

	redis "github.com/redis/go-redis/v9"
)

func StartGame(
	client *redis.Client,
	roomID string,
	userId string,
) error {
	ctx := context.Background()
	roomKey := "room:" + roomID

	return client.Watch(ctx, func(tx *redis.Tx) error {

		// Read current room
		roomJSON, err := tx.Get(ctx, roomKey).Result()
		if err != nil {
			return err
		}

		var room entity.Room
		if err := json.Unmarshal(
			[]byte(roomJSON),
			&room,
		); err != nil {
			return err
		}

		// Make sure user is the hoster
		if room.HosterID != userId {
			return errors.New("only the hoster can start the game")
		}

		// Make sure room is still waiting
		if room.Status != entity.RoomStatusWaiting {
			return errors.New("game has already started")
		}

		// Make sure all users are prepared
		if len(room.Prepared) != len(room.UserIDs) {
			return errors.New("not all users are prepared")
		}

		// Update room status
		room.Status = entity.RoomStatusInGame

		newRoomJSON, err := json.Marshal(room)
		if err != nil {
			return err
		}

		// Commit
		_, err = tx.TxPipelined(ctx, func(pipe redis.Pipeliner) error {

			// Update room
			pipe.Set(
				ctx,
				roomKey,
				newRoomJSON,
				0,
			)

			// Update all users
			for _, userID := range room.UserIDs {
				userStatusKey := "user_status:" + userID

				userStatus := entity.UserStatus{
					UserID: userID,
					Status: entity.UserStatusInGame,
				}

				userStatusJSON, err := json.Marshal(userStatus)
				if err != nil {
					return err
				}

				pipe.Set(
					ctx,
					userStatusKey,
					userStatusJSON,
					0,
				)
			}

			return nil
		})

		return err
	}, roomKey)
}
