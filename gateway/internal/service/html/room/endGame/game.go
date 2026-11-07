package endGame

import (
	"context"
	"gateway/internal/client/room"
	"gateway/internal/function"
	gen "gateway/proto/room/gen"
	redis "github.com/redis/go-redis/v9"
)

// not called by user, called by game service when game ends
func EndGame(
	roomClient *room.Client,
	ctx context.Context,
	roomID string,
	redisClient *redis.Client) (
	bool,
	error) {
	messageID, err := function.GenerateMessageID(ctx, redisClient)
	if err != nil {
		return false, err
	}
	req := &gen.EndGameRequest{
		RoomId:    roomID,
		MessageId: messageID,
	}

	res, err := roomClient.GetClient().EndGame(
		ctx,
		req,
	)

	if err != nil {
		return false, err
	}

	return res.Result, nil
}
