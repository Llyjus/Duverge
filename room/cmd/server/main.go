package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"room/internal/client"
	"room/internal/server"
	pb "room/proto/gen"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

func main() {
	// Redis
	redisClient := client.NewClient()
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		fmt.Println("Error connecting to Redis:", err)
		return
	}

	// gRPC listener
	listener, err := net.Listen("tcp", ":26528")
	if err != nil {
		fmt.Println("Error creating listener:", err)
		return
	}
	defer listener.Close()

	// Register metrics endpoint.
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	go func() {
		fmt.Println("Room metrics server running on :9091")

		if err := http.ListenAndServe(":9091", mux); err != nil {
			fmt.Println("Metrics server failed:", err)
		}
	}()

	// gRPC server
	grpcServer := grpc.NewServer()

	// Room RPC server
	roomServer := server.NewRoomServer(redisClient)

	// Register RoomService
	pb.RegisterRoomServiceServer(grpcServer, roomServer)

	// Start server
	go func() {
		fmt.Println("Room gRPC server running on :26528")

		if err := grpcServer.Serve(listener); err != nil {
			fmt.Println("gRPC server failed:", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	fmt.Println("Shutting down Room server...")

	grpcServer.GracefulStop()

	fmt.Println("Room server stopped")
}
