package entity

type UserStatus struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
}

// status: "free", "in_room", "in_game"
