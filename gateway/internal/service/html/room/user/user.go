package user

import (
	"context"
	"gateway/internal/client/room"
	gen "gateway/proto/room/gen"
)

func createRoom(
	roomClient *room.Client,
	ctx context.Context,
	userId string) (
	string,
	error) {
	req := &gen.CreateRoomRequest{
		UserId: userId,
	}

	res, err := roomClient.GetClient().CreateRoom(
		ctx,
		req,
	)

	if err != nil {
		return "", err
	}

	return res.RoomId, nil
}
