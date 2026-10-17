package protocol

import "encoding/json"

type Response struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}
