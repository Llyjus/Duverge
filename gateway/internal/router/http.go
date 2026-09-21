package router

import (
	"gateway/internal/client/userdata"
	"gateway/internal/service/html/health"
	"gateway/internal/service/html/login"
	"gateway/internal/service/html/sessionLogin"
	"net/http"

	redis "github.com/redis/go-redis/v9"
)

func NewHTTPRouter(userDataClient *userdata.Client, redisClient *redis.Client) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", login.HandleLogin(userDataClient, redisClient))
	mux.HandleFunc("/health", health.HandleHealth)

	mux.HandleFunc("/login_session", sessionLogin.HandleLoginSession(redisClient))

	return mux
}
