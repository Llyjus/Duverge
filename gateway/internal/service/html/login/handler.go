package login

import (
	"encoding/json"
	"fmt"
	"gateway/internal/client/userdata"
	"net/http"
)

func HandleLogin(userDataClient *userdata.Client) http.HandlerFunc {
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

		var res bool
		switch req.Type {

		case "login":
			res, err = login(userDataClient, r.Context(), req.Username, req.Password)

			if err != nil {
				http.Error(w, "Error occurred while logging in", http.StatusInternalServerError)
				return
			}

			switch res {
			case true:
				w.Write([]byte("Login successful"))
			case false:
				http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			}

		case "register":
			res, err = register(userDataClient, r.Context(), req.Username, req.Password)

			if err != nil {
				http.Error(w, "Error occurred while registering", http.StatusInternalServerError)
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
