package userdata

import (
	"context"
	gen "gateway/proto/userdata/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client gen.UserDataServiceClient
	conn   *grpc.ClientConn
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := gen.NewUserDataServiceClient(conn)

	return &Client{
		client: client,
		conn:   conn,
	}, nil
}

func (c *Client) UserLogin(ctx context.Context, request *gen.UserLoginRequest) (*gen.UserLoginResponse, error) {
	return c.client.UserLogin(ctx, request)
}

func (c *Client) UserRegistration(ctx context.Context, request *gen.UserRegistrationRequest) (*gen.UserRegistrationResponse, error) {
	return c.client.UserRegistration(ctx, request)
}

func (c *Client) Close() {
	c.conn.Close()
}
