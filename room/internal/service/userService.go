package service

import (
	"room/internal/operation"

	redis "github.com/redis/go-redis/v9"
)

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
