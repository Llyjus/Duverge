package userdata

import (
	"context"
	"fmt"
	"gateway/internal/client/retry"
	"gateway/internal/errors"
	gen "gateway/proto/userdata/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	// Validate the request
	if request == nil || request.GetMessageId() == "" {
		return nil, fmt.Errorf("UserData login request message_id is required")
	}

	// Use the retry mechanism to call the UserLogin RPC
	response, err := retry.Do(ctx, c.retryPolicy, func(ctx context.Context) (*gen.UserLoginResponse, error) {
		return c.client.UserLogin(ctx, request)
	})

	// Map the error to a more specific error if it's a dependency error
	if err = errors.MapDependencyError(err); err != nil {
		return response, err
	}

	// Check if the response message_id matches the request message_id
	// shouldn't happen or there's a bug in the UserData service, but we should handle it gracefully
	if response != nil && response.GetMessageId() != request.GetMessageId() {
		return nil, fmt.Errorf("UserData login response message_id does not match request")
	}
	return response, nil
}

func (c *Client) UserRegistration(ctx context.Context, request *gen.UserRegistrationRequest) (*gen.UserRegistrationResponse, error) {

	// Validate the request
	if request == nil || request.GetMessageId() == "" {
		return nil, fmt.Errorf("UserData registration request message_id is required")
	}
	// Use the retry mechanism to call the UserRegistration RPC
	response, err := c.client.UserRegistration(ctx, request)
	if err = errors.MapDependencyError(err); err != nil {
		return response, err
	}

	// Check if the response message_id matches the request message_id
	// shouldn't happen or there's a bug in the UserData service, but we should handle it gracefully
	if response != nil && response.GetMessageId() != request.GetMessageId() {
		return nil, fmt.Errorf("UserData registration response message_id does not match request")
	}
	return response, nil
}

func (c *Client) Close() {
	c.conn.Close()
}
