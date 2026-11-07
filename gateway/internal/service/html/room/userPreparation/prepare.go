package userpreparation

import (
	"context"
	"gateway/internal/client/room"
	"gateway/internal/function"
	gen "gateway/proto/room/gen"
	redis "github.com/redis/go-redis/v9"
)

func setPreparation(
	roomClient *room.Client,
	ctx context.Context,
	roomID string,
	userId string,
	targetStatus bool,
	redisClient *redis.Client) (
	bool,
	error) {
	messageID, err := function.GenerateMessageID(ctx, redisClient)
	if err != nil {
		return false, err
	}
	req := &gen.SetPreparationRequest{
		RoomId:       roomID,
		UserId:       userId,
		TargetStatus: targetStatus,
		MessageId:    messageID,
	}

	res, err := roomClient.GetClient().SetPreparation(
		ctx,
		req,
	)

	if err != nil {
		return false, err
	}

	return res.Status, nil
}
