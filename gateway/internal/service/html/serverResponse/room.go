package serverresponse

type GetRoomInfoResponse struct {
	Success bool     `json:"success"`
	RoomID  []string `json:"roomId"`
}

type CreateRoomResponse struct {
	RoomID string `json:"roomId"`
}

type JoinRoomResponse struct {
	Result bool `json:"result"`
}

type LeaveRoomResponse struct {
	Result bool `json:"result"`
}

type SetPreparationResponse struct {
	Status bool `json:"status"`
}

type StartGameResponse struct {
	Result bool `json:"result"`
}

type EndGameResponse struct {
	Result bool `json:"result"`
}
