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

	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	fmt.Println("Gateway is running on port 26526 and 26527...")

	// Set the request timeout for HTTP requests, defaulting to 10 seconds if not specified in the environment variable
	requestTimeout := 10 * time.Second
	if configured := os.Getenv("GATEWAY_REQUEST_TIMEOUT"); configured != "" {
		parsed, err := time.ParseDuration(configured)
		if err != nil || parsed <= 0 {
			fmt.Printf("Invalid GATEWAY_REQUEST_TIMEOUT %q: expected a positive duration\n", configured)
			return
		}
		requestTimeout = parsed
	}

	// Load TCP timeouts from environment variables
	tcpTimeouts, err := loadTCPTimeouts()
	if err != nil {
		fmt.Println(err)
		return
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Register metrics endpoint.
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	go func() {
		fmt.Println("Room metrics server running on :9091")

		if err := http.ListenAndServe(":9091", mux); err != nil {
			fmt.Println("Metrics server failed:", err)
		}
	}()

	// Create a new redis client
	redisClient := redis_client.NewClient()
	defer redisClient.Close()

	// Test the Redis connection
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		fmt.Println("Error connecting to Redis:", err)
		return
	}

	// Create a new userdata client
	userDataClient, err := userdata.NewClient("userdata:26529")
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
	server.HttpServer(userDataClient, redisClient, roomClient, requestTimeout)
	manager := connection.NewConnectionManager()

	// Start the TCP server
	server.TcpServer(manager, redisClient, tcpTimeouts)

	<-quit
	fmt.Println("Shutting down server...")
}

func loadTCPTimeouts() (connection.Timeouts, error) {
	unauthenticated, err := durationFromEnv("GATEWAY_TCP_UNAUTHENTICATED_TIMEOUT", 10*time.Second)
	if err != nil {
		return connection.Timeouts{}, err
	}
	idle, err := durationFromEnv("GATEWAY_TCP_IDLE_TIMEOUT", 45*time.Second)
	if err != nil {
		return connection.Timeouts{}, err
	}
	operation, err := durationFromEnv("GATEWAY_TCP_OPERATION_TIMEOUT", 5*time.Second)
	if err != nil {
		return connection.Timeouts{}, err
	}
	write, err := durationFromEnv("GATEWAY_TCP_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return connection.Timeouts{}, err
	}

	return connection.Timeouts{
		Unauthenticated: unauthenticated,
		Idle:            idle,
		Operation:       operation,
		Write:           write,
	}, nil
}

func durationFromEnv(name string, fallback time.Duration) (time.Duration, error) {
	configured := os.Getenv(name)
	if configured == "" {
		return fallback, nil
	}
	duration, err := time.ParseDuration(configured)
	if err != nil || duration <= 0 {
		return 0, fmt.Errorf("invalid %s %q: expected a positive duration", name, configured)
	}
	return duration, nil
}
