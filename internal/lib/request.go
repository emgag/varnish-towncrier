package lib

import (
	"encoding/json"
	"errors"
	"strings"
)

var validCommands = map[string]bool{
	"ban":       true,
	"ban.url":   true,
	"purge":     true,
	"xkey":      true,
	"xkey.soft": true,
}

// Request contains the purge message received from or sent to redis
type Request struct {
	Host    string   `json:"host"`
	Command string   `json:"command"`
	Value   []string `json:"value"`
}

// Validate validates the request
func (r *Request) Validate() (bool, error) {
	messages := []string{}

	if r.Command == "" {
		messages = append(messages, "command: missing")
	} else if !validCommands[r.Command] {
		messages = append(messages, "Unknown command: "+r.Command)
	}

	if r.Host == "" {
		messages = append(messages, "host: missing")
	}

	if len(r.Value) == 0 {
		messages = append(messages, "value: empty")
	}

	if len(messages) > 0 {
		return false, errors.New(strings.Join(messages, ", "))
	}

	return true, nil
}

// NewRequest create a new Request instance
func NewRequest(jsonInput string) (*Request, error) {
	req := Request{}

	if err := json.Unmarshal([]byte(jsonInput), &req); err != nil {
		return nil, err
	}

	if valid, err := req.Validate(); !valid {
		return nil, err
	}

	return &req, nil
}
