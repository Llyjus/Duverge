package protocol

import "encoding/json"

type Packet struct {
	Version string `json:"version"`
	// Type: "response", "chat", "game"
	Type    string `json:"type"`
	Success bool   `json:"success"`
	// Data contains the response data or error details when Success is false.
	Data json.RawMessage `json:"data,omitempty"`
}
