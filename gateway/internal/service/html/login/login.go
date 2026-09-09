package login

import (
	"context"
	"gateway/internal/client/userdata"
	gen "gateway/proto/userdata/gen"
)

func login(
	userDataClient *userdata.Client,
	context context.Context,
	username,
	password string) (bool, error) {

	req := &gen.UserLoginRequest{
		AccountId: username,
		Password:  password,
	}

	res, err := userDataClient.GetClient().UserLogin(
		context,
		req,
	)

	if err != nil || res == nil {
		return false, err
	}

	return res.Result, nil
}
