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

	err2 := redisClient.Set(
		context.Background(),
		accountID,
		sessionID,
		7*24*time.Hour,
	).Err()

	if err != nil || err2 != nil {
		return "", err
	}

	return sessionID, nil
}

func SessionDeletion(
	redisClient *redis.Client,
	accountID string,
) (string, error) {
	// Find the sessionID associated with the accountID
	sessionID, err := redisClient.Get(context.Background(), accountID).Result()
	if err != nil {
		return "", err
	}

	err2 := redisClient.Del(context.Background(), sessionID).Err()
	err3 := redisClient.Del(context.Background(), accountID).Err()
	if err2 != nil || err3 != nil {
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
