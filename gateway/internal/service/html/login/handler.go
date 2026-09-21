package login

import (
	"encoding/json"
	"fmt"
	"gateway/internal/client/userdata"
	"net/http"

	redis "github.com/redis/go-redis/v9"
)

func HandleLogin(userDataClient *userdata.Client, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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

		switch req.Type {

		case "login":
			var res string
			res, err = login(userDataClient, redisClient, r.Context(), req.Username, req.Password)

			if err != nil {
				http.Error(w, "Error occurred while logging in", http.StatusInternalServerError)
				return
			}

			switch res {
			case "":
				http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			default:

				type LoginResponse struct {
					SessionID string `json:"sessionId"`
				}
				response := LoginResponse{
					SessionID: res,
				}

				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(response)
			}

		case "register":
			var res bool
			res, err = register(userDataClient, r.Context(), req.Username, req.Password)

			if err != nil {
				http.Error(
					w,
					"Error occurred while registering",
					http.StatusInternalServerError)
				return
			}

			switch res {
			case true:
				w.Write([]byte("Registration successful"))
			case false:
				http.Error(w, "Account already exists", http.StatusUnauthorized)
			}

		default:
			http.Error(w, "Invalid request type", http.StatusBadRequest)
		}
	}
}
