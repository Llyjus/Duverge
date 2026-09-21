package login

import (
	"context"
	"gateway/internal/client/userdata"
	"gateway/internal/service/auth"
	gen "gateway/proto/userdata/gen"

	redis "github.com/redis/go-redis/v9"
)

func login(
	userDataClient *userdata.Client,
	redisClient *redis.Client,
	context context.Context,
	accountId string,
	password string) (string, error) {

	req := &gen.UserLoginRequest{
		AccountId: accountId,
		Password:  password,
	}

	res, err := userDataClient.GetClient().UserLogin(
		context,
		req,
	)

	if err != nil || res == nil {
		return "", err
	}
	// Store the session ID in Redis
	sessionId, err := auth.SessionCreation(redisClient,
		res.SessionId,
		accountId)
	if err != nil || sessionId == "" {
		return "", err
	}

	return sessionId, nil
}
