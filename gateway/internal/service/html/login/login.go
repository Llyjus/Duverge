package login

import (
	"context"
	"gateway/internal/client/userdata"
	"gateway/internal/function"
	"gateway/internal/service/auth"
	gen "gateway/proto/userdata/gen"

	redis "github.com/redis/go-redis/v9"
)

func login(
	userDataClient *userdata.Client,
	redisClient *redis.Client,
	ctx context.Context,
	accountId string,
	password string) (string, error) {
	messageID, err := function.GenerateMessageID(ctx, redisClient)
	if err != nil {
		return "", err
	}

	req := &gen.UserLoginRequest{
		MessageId: messageID,
		AccountId: accountId,
		Password:  password,
	}

	res, err := userDataClient.UserLogin(
		ctx,
		req,
	)

	if err != nil || res == nil {
		return "", err
	}
	// Store the session ID in Redis
	if res.SessionId == "" {
		return "", err
	}
	sessionId, err := auth.SessionCreation(ctx,
		redisClient,
		res.SessionId,
		accountId)
	if err != nil || sessionId == "" {
		return "", err
	}

	return sessionId, nil
}
