package user

import (
	"context"
	"gateway/internal/client/room"
	"gateway/internal/function"
	gen "gateway/proto/room/gen"
	redis "github.com/redis/go-redis/v9"
)

func createRoom(
	roomClient *room.Client,
	ctx context.Context,
	userId string,
	redisClient *redis.Client) (
	string,
	error) {
	messageID, err := function.GenerateMessageID(ctx, redisClient)
	if err != nil {
		return "", err
	}
	req := &gen.CreateRoomRequest{
		UserId:    userId,
		MessageId: messageID,
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
