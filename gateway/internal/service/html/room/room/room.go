package room

import (
	"context"
	"gateway/internal/client/room"
	gen "gateway/proto/room/gen"
)

func getRoomInfo(
	roomClient *room.Client,
	ctx context.Context,
	roomID string) (
	[]string,
	error) {
	req := &gen.GetRoomInfoRequest{
		RoomId: roomID,
	}

	res, err := roomClient.GetClient().GetRoomInfo(
		ctx,
		req,
	)

	if err != nil {
		return nil, err
	}

	return res.UserIds, nil
}

func joinRoom(
	roomClient *room.Client,
	ctx context.Context,
	roomID string,
	userId string,
) (
	bool,
	error) {
	req := &gen.JoinRoomRequest{
		RoomId: roomID,
		UserId: userId,
	}

	res, err := roomClient.GetClient().JoinRoom(
		ctx,
		req,
	)

	if err != nil {
		return false, err
	}

	return res.Result, nil
}

func leaveRoom(
	roomClient *room.Client,
	ctx context.Context,
	roomID string,
	userId string,
) (
	bool,
	error) {
	req := &gen.LeaveRoomRequest{
		RoomId: roomID,
		UserId: userId,
	}

	res, err := roomClient.GetClient().LeaveRoom(
		ctx,
		req,
	)

	if err != nil {
		return false, err
	}

	return res.Result, nil
}

func startGame(
	roomClient *room.Client,
	ctx context.Context,
	roomID string,
	userID string) (
	bool,
	error) {
	req := &gen.StartGameRequest{
		RoomId: roomID,
		UserId: userID,
	}

	res, err := roomClient.GetClient().StartGame(
		ctx,
		req,
	)

	if err != nil {
		return false, err
	}

	return res.Result, nil
}
