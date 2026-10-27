package userdata

import (
	"context"
	"fmt"
	"gateway/internal/client/retry"
	internalerrors "gateway/internal/errors"
	gen "gateway/proto/userdata/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	client      gen.UserDataServiceClient
	conn        *grpc.ClientConn
	retryPolicy retry.Policy
}

func NewClient(addr string) (*Client, error) {
	return NewClientWithRetryPolicy(addr, retry.DefaultPolicy())
}

func NewClientWithRetryPolicy(addr string, policy retry.Policy) (*Client, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := gen.NewUserDataServiceClient(conn)

	return &Client{
		client:      client,
		conn:        conn,
		retryPolicy: policy,
	}, nil
}

func (c *Client) UserLogin(ctx context.Context, request *gen.UserLoginRequest) (*gen.UserLoginResponse, error) {
	response, err := retry.Do(ctx, c.retryPolicy, func(ctx context.Context) (*gen.UserLoginResponse, error) {
		return c.client.UserLogin(ctx, request)
	})
	return response, mapDependencyError(err)
}

func (c *Client) UserRegistration(ctx context.Context, request *gen.UserRegistrationRequest) (*gen.UserRegistrationResponse, error) {
	response, err := c.client.UserRegistration(ctx, request)
	return response, mapDependencyError(err)
}

func mapDependencyError(err error) error {
	if err != nil && status.Code(err) == codes.Unavailable {
		return fmt.Errorf("%w: %w", internalerrors.ErrUserDataUnavailable, err)
	}
	return err
}

func (c *Client) Close() {
	c.conn.Close()
}
