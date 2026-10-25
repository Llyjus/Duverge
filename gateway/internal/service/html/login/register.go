package login

import (
	"context"
	"gateway/internal/client/userdata"
	gen "gateway/proto/userdata/gen"
)

func register(
	userDataClient *userdata.Client,
	ctx context.Context,
	username,
	password string) (bool, error) {

	req := &gen.UserRegistrationRequest{
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
