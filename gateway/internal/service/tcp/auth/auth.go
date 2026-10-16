package auth

import (
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"

	"gateway/internal/service/auth"
	connection "gateway/internal/service/tcp/connection"
)

func Authenticate(
	session *connection.Session,
	manager *connection.ConnectionManager,
	redisClient *redis.Client,
	data json.RawMessage,
) error {
	// Decode authentication data.
	var authMsg AuthMessage
	if err := json.Unmarshal(data, &authMsg); err != nil {
		return fmt.Errorf("invalid auth message: %w", err)
	}

	// Check login session.
	if authMsg.LoginSession == "" {
		return fmt.Errorf("missing login session")
	}

	// Authenticate the login session.
	userID, err := auth.SessionAuthentication(
		redisClient,
		authMsg.LoginSession,
	)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	if userID == "" {
		return fmt.Errorf("authentication returned empty user ID")
	}

	// Mark the connection as authenticated.
	session.UserID = userID
	session.AuthStatus = true

	// Register the authenticated session.
	manager.Add(session)

	return nil
}
