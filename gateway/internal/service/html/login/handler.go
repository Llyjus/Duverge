package login

import (
	"encoding/json"
	"errors"
	"fmt"
	"gateway/internal/client/userdata"
	internalerrors "gateway/internal/errors"
	"gateway/internal/metrics"
	serverresponse "gateway/internal/service/html/serverResponse"
	"net/http"
	"time"

	redis "github.com/redis/go-redis/v9"
)

func HandleLogin(userDataClient *userdata.Client, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// Start measuring the duration of the request
		started := time.Now()
		operation := "unknown"
		defer func() {
			if operation != "unknown" {
				metrics.RequestDuration.WithLabelValues("gateway", operation, "http_total").Observe(time.Since(started).Seconds())
			}
		}()

		fmt.Println("Received login request")

		// Decode the request body into a struct
		var req struct {
			Type     string `json:"type"`
			Username string `json:"accountId"`
			Password string `json:"password"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
		if req.Type == "login" || req.Type == "register" {
			operation = req.Type
		}
		switch req.Type {

		case "login":
			var res string
			res, err = login(userDataClient, redisClient, r.Context(), req.Username, req.Password)

			if err != nil {
				if internalerrors.IsRequestTimeout(err) {
					http.Error(w, "Request timed out", http.StatusServiceUnavailable)
					return
				}
				if errors.Is(err, internalerrors.ErrUserDataUnavailable) {
					http.Error(w, "UserData service unavailable", http.StatusServiceUnavailable)
					return
				}
				if errors.Is(err, internalerrors.ErrRedisUnavailable) {
					http.Error(w, "Redis service unavailable", http.StatusServiceUnavailable)
					return
				}
				http.Error(w, "Error occurred while logging in", http.StatusInternalServerError)
				return
			}

			switch res {
			case "":
				http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			default:

				response := serverresponse.LoginResponse{
					Success:   true,
					SessionID: res,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}

		case "register":
			var res bool
			res, err = register(userDataClient, redisClient, r.Context(), req.Username, req.Password)

			if err != nil {
				if internalerrors.IsRequestTimeout(err) {
					http.Error(w, "Request timed out", http.StatusServiceUnavailable)
					return
				}
				if errors.Is(err, internalerrors.ErrUserDataUnavailable) {
					http.Error(w, "UserData service unavailable", http.StatusServiceUnavailable)
					return
				}
				if errors.Is(err, internalerrors.ErrRedisUnavailable) {
					http.Error(w, "Redis service unavailable", http.StatusServiceUnavailable)
					return
				}
				http.Error(
					w,
					"Error occurred while registering",
					http.StatusInternalServerError)
				return
			}

			switch res {
			case true:
				response := serverresponse.RegisterResponse{
					Success: true,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			case false:
				http.Error(w, "Account already exists", http.StatusUnauthorized)
			}

		default:
			http.Error(w, "Invalid request type", http.StatusBadRequest)
		}
	}
}
