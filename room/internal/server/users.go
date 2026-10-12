package server

import (
	"context"

	"room/internal/service"
	pb "room/proto/gen"
)

func (s *RoomServer) CreateRoom(
	ctx context.Context,
	req *pb.CreateRoomRequest,
) (*pb.CreateRoomResponse, error) {

	roomID, err := service.CreateRoom(
		s.client,
		req.GetUserId(),
	)
	if err != nil {
		return nil, err
	}

	return &pb.CreateRoomResponse{
		RoomId: roomID,
	}, nil
}

func (s *RoomServer) JoinRoom(
	ctx context.Context,
	req *pb.JoinRoomRequest,
) (*pb.JoinRoomResponse, error) {

	err := service.AddUserToRoom(
		s.client,
		req.GetRoomId(),
		req.GetUserId(),
	)
	if err != nil {
		return nil, err
	}

	return &pb.JoinRoomResponse{
		Result: true,
	}, nil
}

func (s *RoomServer) LeaveRoom(
	ctx context.Context,
	req *pb.LeaveRoomRequest,
) (*pb.LeaveRoomResponse, error) {

	err := service.DeleteUserFromRoom(
		s.client,
		req.GetRoomId(),
		req.GetUserId(),
	)
	if err != nil {
		return nil, err
	}

	return &pb.LeaveRoomResponse{
		Result: true,
	}, nil
}

func (s *RoomServer) GetRoomInfo(
	ctx context.Context,
	req *pb.GetRoomInfoRequest,
) (*pb.GetRoomInfoResponse, error) {

	userIDs, err := service.GetRoomUsers(
		s.client,
		req.GetRoomId(),
	)
	if err != nil {
		return nil, err
	}

	return &pb.GetRoomInfoResponse{
		RoomId:  req.GetRoomId(),
		UserIds: userIDs,
	}, nil
}
