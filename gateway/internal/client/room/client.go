package room

import (
	gen "gateway/proto/room/gen"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	client gen.RoomServiceClient
	conn   *grpc.ClientConn
}

func (c *Client) GetClient() gen.RoomServiceClient {
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

	client := gen.NewRoomServiceClient(conn)

	return &Client{
		client: client,
		conn:   conn,
	}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}
