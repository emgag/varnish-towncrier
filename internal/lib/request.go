package lib

import (
	"encoding/json"
	"errors"
	"strings"
)

// Request contains the purge message received from or sent to redis
type Request struct {
	Command    string   `json:"command"`
	Expression string   `json:"expression"`
	Host       string   `json:"host"`
	Path       string   `json:"path"`
	Pattern    string   `json:"pattern"`
	Keys       []string `json:"keys"`
}

// Validate validates the request
func (r *Request) Validate() (bool, error) {
	messages := []string{}

	if r.Command == "" {
		messages = append(messages, "command: missing")
	}

	if r.Host == "" {
		messages = append(messages, "host: missing")
	}

	switch r.Command {
	case "ban":
		if r.Expression == "" {
			messages = append(messages, "expression: missing")
		}

	case "ban.url":
		if r.Pattern == "" {
			messages = append(messages, "pattern: missing")
		}

	case "purge":
		if r.Path == "" {
			messages = append(messages, "path: missing")
		}

	case "xkey", "xkey.soft":
		if len(r.Keys) == 0 {
			messages = append(messages, "keys: missing")
		}

	default:
		messages = append(messages, "Unknown command: "+r.Command)
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
