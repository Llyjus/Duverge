package main

import (
	"fmt"
	"gateway/internal/client/userdata"
	"gateway/internal/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Gateway is running on port 26526 and 26527...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	userdata.NewClient("userdata:9090")
	// Start the HTTP server
	server.HttpServer()

	// Start the TCP server
	server.TcpServer()

	<-quit
	fmt.Println("Shutting down server...")
}
