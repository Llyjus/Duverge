package main

import (
	"fmt"
	"gateway/internal/handler"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	fmt.Println("Gateway is running on port 26526...")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// HTTP server transferred from nginx
	httpServer := &http.Server{
		Addr:    "127.0.0.1:26526",
		Handler: http.HandlerFunc(handler.HandleLogin),
	}

	go func() {
		fmt.Println("another thread is running on: 26526")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[HTTP] error: %v\n", err)
		}
	}()

	// TCP server
	tcpListener, err := net.Listen("tcp", ":26527")
	if err != nil {
		panic(err)
	}

	go func() {
		fmt.Println("[TCP] game server is running on: :26527")
		for {
			conn, err := tcpListener.Accept()
			if err != nil {
				fmt.Printf("[TCP] error accepting connection: %v\n", err)
				return
			}

			// for every new connection, handle it in a separate goroutine
			go handler.HandleGameTCP(conn)
		}
	}()

	<-quit
	fmt.Println("Shutting down server...")
}
