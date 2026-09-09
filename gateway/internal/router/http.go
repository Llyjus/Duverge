package router

import (
	"gateway/internal/client/userdata"
	"gateway/internal/service/html/health"
	"gateway/internal/service/html/login"
	"net/http"
)

func NewHTTPRouter(userDataClient *userdata.Client) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", login.HandleLogin(userDataClient))

	mux.HandleFunc("/health", health.HandleHealth)

	return mux
}
