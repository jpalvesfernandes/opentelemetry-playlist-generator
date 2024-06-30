package models

import "encoding/json"

type Request struct {
	Taste json.RawMessage `json:"taste"`
	Token json.RawMessage `json:"token"`
}
