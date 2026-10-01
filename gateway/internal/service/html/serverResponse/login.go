package serverresponse

type LoginResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"sessionId"`
}

type RegisterResponse struct {
	Success bool `json:"success"`
}

type SessionLoginResponse struct {
	Success bool `json:"success"`
}
