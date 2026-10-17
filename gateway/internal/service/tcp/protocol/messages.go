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

// Error response structure
type ErrorResponse struct {
	ErrorCode string `json:"error_code"`
	Detail    string `json:"detail"`
}

// Chat message strtuctures
type ChatMessage struct {
	// Type: "room", "private"
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Response to the sender
type ChatResponse struct {
	Success   bool   `json:"success"`
	ErrorCode string `json:"error_code,omitempty"`
}

// post the message to the room or private chat
type ChatPost struct {
	MessageID string `json:"message_id"`
	SenderID  string `json:"sender_id"`
	// Type: "text", "emoji"
	Type    string `json:"type"`
	Content string `json:"content"`
}

// Room chat message structure
type RoomChatMessage struct {
	RoomID string `json:"room_id"`
	// Type: "text", "emoji"
	Type    string `json:"type"`
	Content string `json:"content"`
}

// Private chat message structure
type PrivateChatMessage struct {
	To string `json:"to"`
	// Type: "text", "emoji"
	Type    string `json:"type"`
	Content string `json:"content"`
}
