package login

import (
	"fmt"
	"net/http"
)

func HandleLogin(w http.ResponseWriter, r *http.Request) {

	fmt.Println("Received login request")

	res := login()
	w.Write([]byte(res))

}
