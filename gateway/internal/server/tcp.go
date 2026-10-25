package server

import (
	"fmt"
	"net"

	"github.com/redis/go-redis/v9"

	connection "gateway/internal/service/tcp/connection"
	"gateway/internal/service/tcp/handler"
)

func TcpServer(
	manager *connection.ConnectionManager,
	redisClient *redis.Client,
	timeouts connection.Timeouts,
) {
	tcpListener, err := net.Listen("tcp", ":26527")
	if err != nil {
		panic(err)
	}
	defer tcpListener.Close()

	fmt.Println("[TCP] tcp server is running...")

	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			fmt.Printf("[TCP] error accepting connection: %v\n", err)
			continue
		}

		session := connection.NewSession(conn, timeouts.Write)

		go handler.HandleConnection(
			session,
			manager,
			redisClient,
			timeouts,
		)
	}
}
