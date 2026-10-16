package handler

import (
	"encoding/json"
)

// E.g. {"version": "1.0", "type": "auth", "data": "{\"token\":\"abc123\"}"}
type Message struct {
	Version string          `json:"version,omitempty"`
	Type    string          `json:"type"`
	Data    json.RawMessage `json:"data"`
}
