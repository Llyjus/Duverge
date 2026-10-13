package room

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
			RoomID    string `json:"roomId"`
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

		case "getRoomInfo":
			var res []string
			res, err = getRoomInfo(roomClient, r.Context(), req.RoomID)

			if err != nil {
				http.Error(w, "Error occurred while getting room", http.StatusInternalServerError)
				return
			}

			if res != nil {
				response := serverresponse.GetRoomInfoResponse{
					Success: true,
					RoomID:  res,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			} else {
				http.Error(w, "Invalid room ID", http.StatusUnauthorized)
			}

		case "joinRoom":
			var res bool
			res, err = joinRoom(roomClient, r.Context(), req.RoomID, userId)

			if err != nil {
				http.Error(w, "Error occurred while joining room", http.StatusInternalServerError)
				return
			}

			if res {
				response := serverresponse.JoinRoomResponse{
					Result: res,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			} else {
				http.Error(w, "Failed to join room", http.StatusUnauthorized)
			}

		case "leaveRoom":
			var res bool
			res, err = leaveRoom(roomClient, r.Context(), req.RoomID, userId)

			if err != nil {
				http.Error(w, "Error occurred while leaving room", http.StatusInternalServerError)
				return
			}

			if res {
				response := serverresponse.LeaveRoomResponse{
					Result: res,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			} else {
				http.Error(w, "Failed to leave room", http.StatusUnauthorized)
			}

		case "startGame":
			var res bool
			res, err = startGame(roomClient, r.Context(), req.RoomID, userId)

			if err != nil {
				http.Error(w, "Error occurred while starting game", http.StatusInternalServerError)
				return
			}

			if res {
				response := serverresponse.StartGameResponse{
					Result: res,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			} else {
				http.Error(w, "Failed to start game", http.StatusUnauthorized)
			}

		default:

			http.Error(w, "Invalid request type", http.StatusBadRequest)
		}
	}
}
