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
	// Type: "room", "private"
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// Response to the sender
type ChatResponse struct {
	RequestID string `json:"response_id,omitempty"`
	Status    string `json:"status"`
	Detail    string `json:"detail"`
}

// post the message to the room or private chat
type ChatPost struct {
	MessageID string `json:"message_id"`
	Timestamp int64  `json:"timestamp"`
	SenderID  string `json:"sender_id"`
	Content   string `json:"content"`
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
