package entity

type Room struct {
	// 5 digit unique code
	RoomID      string   `json:"roomId"`
	MaximumSize int64    `json:"maximumSize"`
	UserIDs     []string `json:"userIds"`
	HosterID    string   `json:"hosterId"`
	Prepared    []string `json:"prepared"`
	Status      string   `json:"status"`
}

// game can be started when all users are prepared
const (
	// status: "waiting", "in_game", "finished"
	RoomStatusWaiting  = "waiting"
	RoomStatusInGame   = "in_game"
	RoomStatusFinished = "finished"
)
