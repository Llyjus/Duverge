package server

import (
	"room/proto/gen"

	"github.com/redis/go-redis/v9"
)

type RoomServer struct {
	gen.UnimplementedRoomServiceServer
	client *redis.Client
}

func NewRoomServer(client *redis.Client) *RoomServer {
	return &RoomServer{
		client: client,
	}
}
