package endGame

import (
	"context"
	"gateway/internal/client/room"
	gen "gateway/proto/room/gen"
)

// not called by user, called by game service when game ends
func EndGame(
	roomClient *room.Client,
	ctx context.Context,
	roomID string) (
	bool,
	error) {
	req := &gen.EndGameRequest{
		RoomId: roomID,
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
