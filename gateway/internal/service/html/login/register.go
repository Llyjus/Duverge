package login

import (
	"context"
	"gateway/internal/client/userdata"
	gen "gateway/proto/userdata/gen"
)

func register(
	userDataClient *userdata.Client,
	context context.Context,
	username,
	password string) (bool, error) {

	req := &gen.UserRegistrationRequest{
		AccountId: username,
		Password:  password,
	}

	res, err := userDataClient.GetClient().UserRegistration(
		context,
		req,
	)

	if err != nil || res == nil {
		return false, err
	}

	return res.Result, nil
}
