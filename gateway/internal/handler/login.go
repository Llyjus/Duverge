package handler

import (
	"fmt"
	"gateway/internal/service"
	"net/http"
)

func HandleLogin(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Received login request")

	res := service.Login()
	w.Write([]byte(res))

}
