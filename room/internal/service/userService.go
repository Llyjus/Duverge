package service

import (
	"room/internal/entity"
	"room/internal/operation"

	redis "github.com/redis/go-redis/v9"
)

func AddUserToRoom(
	client *redis.Client,
	roomId string,
	userId string,
) error {
	room, err := operation.GetRoom(client, roomId)
	if err != nil {
		return err
	}
	if room == nil {
		return nil // Room not found
	}

	// Check room status
	if room.Status != entity.RoomStatusWaiting {
		return nil // Room is not in a state to accept new users
	}
	// // Check user status
	UserStatus, err := operation.CheckUserStatus(client, userId)
	if err != nil {
		return err
	}
	// User is not in the lobby
	if UserStatus != entity.UserStatusInLobby {
		return nil // User is not in the lobby
	}

	// Add the user to the room
	room.UserIDs = append(room.UserIDs, userId)

	// Update the user's status
	err = operation.UpdateUserStatus(client,
		userId,
		entity.UserStatusInRoom)
	if err != nil {
		return err
	}

	// Update the room in Redis
	err = operation.UpdateRoom(client, room)
	if err != nil {
		return err
	}

	return nil
}

func DeleteUserFromRoom(
	client *redis.Client,
	roomId string,
	userId string,
) error {
	room, err := operation.GetRoom(client, roomId)
	if err != nil {
		return err
	}
	if room == nil {
		return nil // Room not found
	}

	// Remove the user from the room
	err = operation.DeleteUser(room, userId)
	if err != nil {
		return err
	}

	// Update the user's status to "in_lobby"
	err = operation.UpdateUserStatus(client, userId, entity.UserStatusInLobby)
	if err != nil {
		return err
	}

	// Check if the room is empty after removing the user
	isEmpty, err := operation.CheckRoomIsEmpty(room)
	if err != nil {
		return err
	}
	if isEmpty {
		// Delete the room if it's empty
		err = operation.DeleteRoom(client, roomId)
		if err != nil {
			return err
		} else {
			return nil
		}
	}

	// Check if the user is hoster
	isHoster, err := operation.CheckHoster(room, userId)
	if err != nil {
		return err
	}
	if isHoster {
		// Randomly select a new hoster from the remaining users
		room, err = operation.RandomHoster(room)
		if err != nil {
			return err
		}
	}
	err = operation.UpdateRoom(client, room)
	if err != nil {
		return err
	}

	return nil
}

//TODO: add finite status machine to user status and room status, to avoid invalid state transition
