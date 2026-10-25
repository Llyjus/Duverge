package sessionLogin

import (
	"encoding/json"
	"errors"
	"fmt"
	internalerrors "gateway/internal/errors"
	"gateway/internal/service/auth"
	serverresponse "gateway/internal/service/html/serverResponse"
	"net/http"

	redis "github.com/redis/go-redis/v9"
)

func HandleLoginSession(redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received login request by session")

		//decode
		var req struct {
			SessionID string `json:"sessionId"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Authenticate the session ID
		accountID, err := auth.SessionAuthentication(r.Context(), redisClient, req.SessionID)
		if err != nil {
			if errors.Is(err, internalerrors.ErrRedisUnavailable) {
				http.Error(w, "Session service unavailable", http.StatusServiceUnavailable)
				return
			}
			http.Error(w, "Unable to authenticate session", http.StatusInternalServerError)
			return
		}

		// Create the response
		var response serverresponse.SessionLoginResponse
		if accountID == "" {
			http.Error(w, "Invalid session ID or session expired", http.StatusUnauthorized)
		} else {
			response = serverresponse.SessionLoginResponse{
				Success: true,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)

	}
}
