package service

import (
	"context"
	"errors"

	"room/internal/entity"
	"room/internal/operation"

	redis "github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func HandleDisconnect(ctx context.Context, client *redis.Client, userID string) error {
	if userID == "" {
		return status.Error(codes.InvalidArgument, "user_id is required")
	}

	for attempt := 0; attempt < 3; attempt++ {
		userStatus, err := operation.GetUserStatus(ctx, client, userID)
		if err != nil {
			return err
		}
		if userStatus == nil || userStatus.Status == entity.UserStatusInLobby {
			return nil
		}

		roomID := userStatus.RoomID
		if roomID == "" {
			roomID, err = operation.FindRoomIDByUser(ctx, client, userID)
			if err != nil {
				return err
			}
		}

		if userStatus.Status == entity.UserStatusInGame {
			err = operation.MarkUserDisconnected(ctx, client, roomID, userID)
			if errors.Is(err, operation.ErrUserNotInGame) {
				if roomID == "" {
					err = operation.DeleteUserStatus(ctx, client, userID)
				} else {
					err = operation.RemoveDisconnectedUser(ctx, client, roomID, userID)
					if errors.Is(err, operation.ErrUserIsInGame) {
						err = operation.MarkUserDisconnected(ctx, client, roomID, userID)
					}
				}
			}
			if errors.Is(err, operation.ErrUserNotInGame) || err == redis.TxFailedErr {
				continue
			}
			return err
		}

		if userStatus.Status != entity.UserStatusInRoom {
			return nil
		}
		if roomID == "" {
			return operation.DeleteUserStatus(ctx, client, userID)
		}

		err = operation.RemoveDisconnectedUser(ctx, client, roomID, userID)
		if errors.Is(err, operation.ErrUserIsInGame) {
			err = operation.MarkUserDisconnected(ctx, client, roomID, userID)
			if errors.Is(err, operation.ErrUserNotInGame) {
				continue
			}
		}
		if err == redis.TxFailedErr {
			continue
		}
		return err
	}
	return redis.TxFailedErr
}

func AddUserToRoom(
	client *redis.Client,
	roomId string,
	userId string,
) error {
	var err error

	for i := 0; i < 3; i++ {
		err = operation.JoinRoom(client, roomId, userId)

		if err == nil {
			return nil
		}

		if err != redis.TxFailedErr {
			return err
		}
	}

	return err
}

func DeleteUserFromRoom(
	client *redis.Client,
	roomId string,
	userId string,
) error {
	var err error

	for i := 0; i < 3; i++ {
		err = operation.ExitRoom(client, roomId, userId)

		if err == nil {
			return nil
		}

		if err != redis.TxFailedErr {
			return err
		}
	}

	// check if the room is empty and delete it if necessary
	room, err := operation.GetRoom(client, roomId)
	if err != nil {
		return err
	}
	if room == nil {
		return nil
	}
	res, err := operation.CheckRoomIsEmpty(room)
	if err != nil {
		return err
	}
	if res {
		return operation.DeleteRoom(client, roomId)
	}

	return err
}

func SetUserPreparation(
	client *redis.Client,
	roomID string,
	userId string,
	targetIfPrepared bool,
) error {
	var err error

	for i := 0; i < 3; i++ {
		err = operation.SetPreparation(
			client,
			roomID,
			userId,
			targetIfPrepared,
		)

		if err == nil {
			return nil
		}

		if err != redis.TxFailedErr {
			return err
		}
	}

	return err
}
