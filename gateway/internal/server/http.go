package server

import (
	"fmt"
	"gateway/internal/client/userdata"
	"gateway/internal/router"
	"net/http"
)

func HttpServer(userDataClient *userdata.Client) {
	// HTTP server transferred from nginx
	httpServer := &http.Server{
		Addr:    "0.0.0.0:26526",
		Handler: router.NewHTTPRouter(userDataClient),
	}

	go func() {
		fmt.Println("[HTTP] new thread is running...")

		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("[HTTP] error: %v\n", err)
		}
	}()
}
