package auth

type AuthMessage struct {
	Type         string `json:"type"`
	LoginSession string `json:"token"`
}
