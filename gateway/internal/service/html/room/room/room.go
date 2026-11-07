package room

import (
	"context"
	"gateway/internal/client/room"
	"gateway/internal/function"
	gen "gateway/proto/room/gen"
	redis "github.com/redis/go-redis/v9"
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

	res, err := roomClient.GetRoomInfo(
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
	redisClient *redis.Client,
) (
	bool,
	error) {
	messageID, err := function.GenerateMessageID(ctx, redisClient)
	if err != nil {
		return false, err
	}
	req := &gen.JoinRoomRequest{
		RoomId:    roomID,
		UserId:    userId,
		MessageId: messageID,
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
	redisClient *redis.Client,
) (
	bool,
	error) {
	messageID, err := function.GenerateMessageID(ctx, redisClient)
	if err != nil {
		return false, err
	}
	req := &gen.LeaveRoomRequest{
		RoomId:    roomID,
		UserId:    userId,
		MessageId: messageID,
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
	userID string,
	redisClient *redis.Client) (
	bool,
	error) {
	messageID, err := function.GenerateMessageID(ctx, redisClient)
	if err != nil {
		return false, err
	}
	req := &gen.StartGameRequest{
		RoomId:    roomID,
		UserId:    userID,
		MessageId: messageID,
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
