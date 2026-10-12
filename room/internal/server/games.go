package server

import (
	"context"

	"room/internal/service"
	pb "room/proto/gen"
)

func (s *RoomServer) StartGame(
	ctx context.Context,
	req *pb.StartGameRequest,
) (*pb.StartGameResponse, error) {

	err := service.StartRoomGame(
		s.client,
		req.GetRoomId(),
		req.GetUserId(),
	)
	if err != nil {
		return nil, err
	}

	return &pb.StartGameResponse{
		Result: true,
	}, nil
}

func (s *RoomServer) EndGame(
	ctx context.Context,
	req *pb.EndGameRequest,
) (*pb.EndGameResponse, error) {

	err := service.EndRoomGame(
		s.client,
		req.GetRoomId(),
	)
	if err != nil {
		return nil, err
	}

	return &pb.EndGameResponse{
		Result: true,
	}, nil
}
