package router

import (
	"gateway/internal/service/html/health"
	"gateway/internal/service/html/login"
	"net/http"
)

func NewHTTPRouter() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", login.HandleLogin)

	mux.HandleFunc("/health", health.HandleHealth)

	return mux
}
