package entity

type UserStatus struct {
	UserID string `json:"user_id"`
	Status string `json:"status"`
}

const (
	// status: "in_lobby", "in_room", "in_game"
	UserStatusInLobby = "in_lobby"
	UserStatusInRoom  = "in_room"
	UserStatusInGame  = "in_game"
)
