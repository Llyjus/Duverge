package redis_client

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func NewClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")

	return redis.NewClient(&redis.Options{
		Addr: addr,
	})
}
