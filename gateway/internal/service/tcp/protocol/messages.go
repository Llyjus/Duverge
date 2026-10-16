package protocol

import (
	"encoding/json"
)

// E.g. {"version": "1.0", "type": "auth", "data": "{\"token\":\"abc123\"}"}
type Message struct {
	Version   string          `json:"version,omitempty"`
	RequestID string          `json:"request_id,omitempty"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data"`
}

// Chat message strtuctures
type ChatMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type RoomChatMessage struct {
	RoomID  string `json:"room_id"`
	Content string `json:"content"`
}

type PrivateChatMessage struct {
	To      string `json:"to"`
	Content string `json:"content"`
}
