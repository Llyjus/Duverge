package login

import (
	"context"
	"gateway/internal/client/userdata"
	"gateway/internal/function"
	gen "gateway/proto/userdata/gen"

	redis "github.com/redis/go-redis/v9"
)

func register(
	userDataClient *userdata.Client,
	redisClient *redis.Client,
	ctx context.Context,
	username,
	password string) (bool, error) {
	messageID, err := function.GenerateMessageID(ctx, redisClient)
	if err != nil {
		return false, err
	}

	req := &gen.UserRegistrationRequest{
		MessageId: messageID,
		AccountId: username,
		Password:  password,
	}

	res, err := userDataClient.UserRegistration(
		ctx,
		req,
	)

	if err != nil || res == nil {
		return false, err
	}

	return res.Result, nil
}
