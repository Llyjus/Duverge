package main

import (
	"fmt"
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
		Handler: http.HandlerFunc(handleLogin),
	}

	go func() {
		fmt.Println("another thread is running on: 26526")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[HTTP] error: %v\n", err)
		}
	}()

	<-quit
	fmt.Println("Shutting down server...")
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received login request")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login successful"))
}
