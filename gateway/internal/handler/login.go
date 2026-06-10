package handler

import (
	"fmt"
	"net/http"
)

func HandleLogin(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Received login request")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Login successful"))

}
