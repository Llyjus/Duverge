package endGame

import (
	"encoding/json"
	"fmt"
	"gateway/internal/client/room"
	serverresponse "gateway/internal/service/html/serverResponse"
	"net/http"

	redis "github.com/redis/go-redis/v9"
)

// not called by user, called by game service when game ends

func HandleRoom(roomClient *room.Client, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Received room request")

		// Decode the request body into a struct
		var req struct {
			Type   string `json:"type"`
			RoomID string `json:"roomId"`
		}
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		switch req.Type {

		case "endGame":
			var res bool
			res, err = EndGame(roomClient, r.Context(), req.RoomID)

			if err != nil {
				http.Error(w, "Error occurred while ending game", http.StatusInternalServerError)
				return
			}

			response := serverresponse.SetPreparationResponse{
				Status: res,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(response)

		default:

			http.Error(w, "Invalid request type", http.StatusBadRequest)
		}
	}
}
