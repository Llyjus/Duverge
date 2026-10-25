package router

import (
	"gateway/internal/client/room"
	"gateway/internal/client/userdata"
	"gateway/internal/service/html/health"
	"gateway/internal/service/html/login"
	"gateway/internal/service/html/room/endGame"
	rm "gateway/internal/service/html/room/room"
	"gateway/internal/service/html/room/user"
	prep "gateway/internal/service/html/room/userPreparation"
	"gateway/internal/service/html/sessionLogin"
	"net/http"
	"time"

	redis "github.com/redis/go-redis/v9"
)

func NewHTTPRouter(userDataClient *userdata.Client,
	redisClient *redis.Client,
	roomClient *room.Client,
	requestTimeout time.Duration,
) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/login", login.HandleLogin(userDataClient, redisClient))
	mux.HandleFunc("/health", health.HandleHealth)

	mux.HandleFunc("/login_session", sessionLogin.HandleLoginSession(redisClient))

	// createRoom
	mux.HandleFunc("/room/user", user.HandleRoom(roomClient, redisClient))

	// setPreparation
	mux.HandleFunc("/room/userPreparation", prep.HandleRoom(roomClient, redisClient))

	// endGame
	mux.HandleFunc("/room/endGame", endGame.HandleRoom(roomClient, redisClient))

	// getRoomInfo, joinRoom, leaveRoom, startGame
	mux.HandleFunc("/room/room", rm.HandleRoom(roomClient, redisClient))
	return http.TimeoutHandler(mux, requestTimeout, "Request timed out\n")
}
