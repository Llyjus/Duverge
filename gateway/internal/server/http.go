package server

import (
	"fmt"
	"gateway/internal/handler"
	"net/http"
)

func HttpServer() {
	// HTTP server transferred from nginx
	httpServer := &http.Server{
		Addr:    "127.0.0.1:26526",
		Handler: http.HandlerFunc(handler.HandleLogin),
	}

	go func() {
		fmt.Println("another thread is running...")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[HTTP] error: %v\n", err)
		}
	}()
}
