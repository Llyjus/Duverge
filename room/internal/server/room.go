package server

import (
	"room/internal/service"
	"room/proto/gen"

	"github.com/redis/go-redis/v9"
)

type RoomServer struct {
	gen.UnimplementedRoomServiceServer
	client             *redis.Client
	idempotencyService *service.IdempotencyService
}

func NewRoomServer(client *redis.Client) *RoomServer {
	return &RoomServer{
		client:             client,
		idempotencyService: service.NewIdempotencyService(client),
	}
}
