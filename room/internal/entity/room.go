package entity

type Room struct {
	// 5 digit unique code
	RoomID      string   `json:"roomId"`
	MaximumSize int64    `json:"maximumSize"`
	UserIDs     []string `json:"userIds"`
	HosterID    string   `json:"hosterId"`
	Status      string   `json:"status"`
}
