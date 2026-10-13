package userpreparation

import (
	"context"
	"gateway/internal/client/room"
	gen "gateway/proto/room/gen"
)

func setPreparation(
	roomClient *room.Client,
	ctx context.Context,
	roomID string,
	userId string,
	targetStatus bool) (
	bool,
	error) {
	req := &gen.SetPreparationRequest{
		RoomId:       roomID,
		UserId:       userId,
		TargetStatus: targetStatus,
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
