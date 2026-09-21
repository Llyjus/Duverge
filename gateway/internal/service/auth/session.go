package auth

import (
	"context"
	"time"

	redis "github.com/redis/go-redis/v9"
)

func SessionCreation(
	redisClient *redis.Client,
	sessionID string,
	accountID string,
) (string, error) {

	err := redisClient.Set(
		context.Background(),
		sessionID,
		accountID,
		7*24*time.Hour,
	).Err()

	if err != nil {
		return "", err
	}

	return sessionID, nil
}

func SessionAuthentication(
	redisClient *redis.Client,
	session string,
) (string, error) {

	accountID, err := redisClient.Get(context.Background(), session).Result()
	if err != nil {
		return "", err
	}

	return accountID, nil
}
