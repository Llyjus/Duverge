package auth

import (
	"context"
	"strings"
	"time"

	redis "github.com/redis/go-redis/v9"
)

func SessionCreation(
	redisClient *redis.Client,
	sessionID string,
	accountID string,
) (string, error) {

	// Check if the session already exists
	session, err := redisClient.Get(context.Background(), accountID).Result()
	if err != redis.Nil && err != nil {
		return "", err
	}

	if session != "" {
		// Delete the existing session and account ID from Redis
		err = redisClient.Del(context.Background(), session).Err()
		err = redisClient.Del(context.Background(), accountID).Err()
		if err != nil {
			return "", err
		}
	}

	// add the pre to sessionID
	sessionID = "session=" + sessionID

	// set the session ID and account ID in Redis with an expiration time of 7 days
	err = redisClient.Set(
		context.Background(),
		sessionID,
		accountID,
		7*24*time.Hour,
	).Err()

	if err != nil {
		return "", err
	}

	err = redisClient.Set(
		context.Background(),
		accountID,
		sessionID,
		7*24*time.Hour,
	).Err()

	if err != nil {
		return "", err
	}
	sessionID = strings.TrimPrefix(sessionID, "session=")

	return sessionID, nil
}

func SessionAuthentication(
	redisClient *redis.Client,
	session string,
) (string, error) {

	session = "session=" + session

	accountID, err := redisClient.Get(context.Background(), session).Result()
	if err != nil && err != redis.Nil {
		return "", err
	}

	return accountID, nil
}
