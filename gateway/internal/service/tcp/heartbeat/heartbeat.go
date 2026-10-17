package heartbeat

import (
	"context"
	"fmt"
	auth "gateway/internal/service/auth"
	connection "gateway/internal/service/tcp/connection"
	protocol "gateway/internal/service/tcp/protocol"
	"log"

	"github.com/redis/go-redis/v9"
)

// HandleHeartbeat processes heartbeat messages from clients
func HandleHeartbeat(
	session *connection.Session,
	redisClient *redis.Client,
	ctx context.Context,
) error {
	err := auth.RenewUserStatus(
		redisClient,
		ctx,
		session.UserID,
	)
	if err != nil {
		log.Printf("failed to renew user status for user %s: %v",
			session.UserID, err)

		return fmt.Errorf(
			"failed to renew user status for user %s: %w",
			session.UserID,
			protocol.ErrInternal,
		)
	}
	return nil
}
