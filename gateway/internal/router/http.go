package router

import (
	"gateway/internal/handler"
	"net/http"
)

func NewHTTPRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", handler.HandleLogin)

	mux.HandleFunc("/health", handler.HandleHealth)

	return mux
}
