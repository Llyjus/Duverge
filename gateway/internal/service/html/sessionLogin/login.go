package sessionLogin

import (
	"encoding/json"
	"fmt"
	"gateway/internal/service/auth"
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

		accountID, err := auth.SessionAuthentication(redisClient, req.SessionID)
		if err != nil {
			http.Error(w, "Invalid session ID", http.StatusUnauthorized)
			return
		}

		fmt.Println("Authenticated user:", accountID)

	}
}
