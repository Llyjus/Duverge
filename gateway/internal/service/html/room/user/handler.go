package user

import (
	"encoding/json"
	"fmt"
	"gateway/internal/client/room"
	"gateway/internal/service/auth"
	serverresponse "gateway/internal/service/html/serverResponse"
	"net/http"

	redis "github.com/redis/go-redis/v9"
)

func HandleRoom(roomClient *room.Client, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received room request")

		// Decode the request body into a struct
		var req struct {
			Type      string `json:"type"`
			SessionID string `json:"sessionId"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		userId, err := auth.GetUserIDFromSessionID(redisClient, r.Context(), req.SessionID)
		if err != nil {
			http.Error(w, "Error occurred while fetching user ID", http.StatusInternalServerError)
			return
		}

		switch req.Type {

		case "createRoom":
			var res string
			res, err = createRoom(roomClient, r.Context(), userId)

			if err != nil {
				http.Error(w, "Error occurred while getting room", http.StatusInternalServerError)
				return
			}

			if res != "" {
				response := serverresponse.CreateRoomResponse{
					RoomID: res,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			} else {
				http.Error(w, "Failed to create room", http.StatusUnauthorized)
			}

		default:

			http.Error(w, "Invalid request type", http.StatusBadRequest)
		}
	}
}
