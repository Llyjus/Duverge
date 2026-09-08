package game

import (
	"fmt"
	"net"
)

func HandleGame(conn net.Conn) {
	// close the connection when the function returns
	defer func() {
		fmt.Printf("[TCP] player disconnected: %s\n", conn.RemoteAddr().String())
		conn.Close()
	}()
	fmt.Printf("[TCP] player connected: %s\n", conn.RemoteAddr().String())

	buf := make([]byte, 1024)
	for {
		// Read data from the connection
		n, err := conn.Read(buf)
		if err != nil {
			break
		}

		fmt.Printf("[Data] Received message: %s\n", string(buf[:n]))

		conn.Write([]byte("Server Received: " + string(buf[:n])))
	}
}
