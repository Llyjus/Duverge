package userdata

import (
	gen "gateway/proto/userdata/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client gen.UserDataServiceClient
	conn   *grpc.ClientConn
}

func (c *Client) GetClient() gen.UserDataServiceClient {
	return c.client
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

func (c *Client) Close() {
	c.conn.Close()
}
