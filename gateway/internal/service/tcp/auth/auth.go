package auth

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"gateway/internal/service/auth"
	connection "gateway/internal/service/tcp/connection"
	protocol "gateway/internal/service/tcp/protocol"
)

func Authenticate(
	session *connection.Session,
	manager *connection.ConnectionManager,
	redisClient *redis.Client,
	ctx context.Context,
	data json.RawMessage,
) error {
	// Decode authentication data
	var authMsg AuthMessage
	if err := json.Unmarshal(data, &authMsg); err != nil {
		return fmt.Errorf("%w: %v", protocol.ErrInvalidMessage, err)
	}

	// Check login session
	if authMsg.LoginSession == "" {
		return fmt.Errorf("%w: missing login session", protocol.ErrInvalidMessage)
	}

	// Authenticate the login session
	userID, err := auth.SessionAuthentication(
		ctx,
		redisClient,
		authMsg.LoginSession,
	)
	if err != nil {
		return fmt.Errorf("%w: %v", protocol.ErrAuthenticationFailed, err)
	}

	if userID == "" {
		return fmt.Errorf("%w: empty user ID", protocol.ErrAuthenticationFailed)
	}

	// Mark the connection as authenticated
	session.UserID = userID
	session.AuthStatus = true

	// Register the authenticated session
	manager.Add(session)

	// Create user status in Redis
	if err := auth.UserStatusCreation(
		redisClient,
		ctx,
		userID,
	); err != nil {
		return fmt.Errorf("%w: failed to create user status: %v", protocol.ErrInternal, err)
	}

	return nil
}
