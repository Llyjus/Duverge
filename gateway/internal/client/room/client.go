package room

import (
	"context"
	"gateway/internal/client/retry"
	gen "gateway/proto/room/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client      gen.RoomServiceClient
	conn        *grpc.ClientConn
	retryPolicy retry.Policy
}

func (c *Client) GetClient() gen.RoomServiceClient {
	return c.client
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

	client := gen.NewRoomServiceClient(conn)

	return &Client{
		client:      client,
		conn:        conn,
		retryPolicy: policy,
	}, nil
}

// GetRoomInfo is safe to retry because it does not mutate Room state.
func (c *Client) GetRoomInfo(ctx context.Context, request *gen.GetRoomInfoRequest) (*gen.GetRoomInfoResponse, error) {
	return retry.Do(ctx, c.retryPolicy, func(ctx context.Context) (*gen.GetRoomInfoResponse, error) {
		return c.client.GetRoomInfo(ctx, request)
	})
}

func (c *Client) Close() {
	c.conn.Close()
}
