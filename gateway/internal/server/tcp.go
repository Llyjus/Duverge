package server

import (
	"fmt"
	"gateway/internal/service/tcp/game"
	"net"
)

func TcpServer() {
	// TCP server
	tcpListener, err := net.Listen("tcp", ":26527")
	if err != nil {
		panic(err)
	}

	fmt.Println("[TCP] game server is running...")

	go func() {
		for {
			conn, err := tcpListener.Accept()
			if err != nil {
				fmt.Printf("[TCP] error accepting connection: %v\n", err)
				return
			}
			// for every new connection, handle it in a separate goroutine
			go game.HandleGame(conn)
		}
	}()
}
