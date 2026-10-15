package main

import (
	"context"
	"fmt"
	"gateway/internal/client/redis_client"
	"gateway/internal/client/room"
	"gateway/internal/client/userdata"
	"gateway/internal/server"
	"gateway/internal/service/tcp/connection"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Gateway is running on port 26526 and 26527...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Create a new redis client
	redisClient := redis_client.NewClient()
	defer redisClient.Close()

	// Test the Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		fmt.Println("Error connecting to Redis:", err)
		return
	}

	// Create a new userdata client
	userDataClient, err := userdata.NewClient("userdata:9090")
	if err != nil {
		fmt.Println("Error creating userdata client:", err)
		return
	}
	defer userDataClient.Close()
	// create a new room client
	roomClient, err := room.NewClient("room:26528")
	if err != nil {
		fmt.Println("Error creating room client:", err)
		return
	}
	defer roomClient.Close()

	// Start the HTTP server
	server.HttpServer(userDataClient, redisClient, roomClient)
	manager := connection.NewConnectionManager()

	// Start the TCP server
	server.TcpServer(manager, redisClient)

	<-quit
	fmt.Println("Shutting down server...")
}
