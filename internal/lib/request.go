package lib

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Request struct {
	Command    string   `json:"command"`
	Expression string   `json:"expression"`
	Host       string   `json:"host"`
	Path       string   `json:"path"`
	Pattern    string   `json:"pattern"`
	Keys       []string `json:"keys"`
}

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

type RequestProcessor struct {
	Config Options
}

func (rp *RequestProcessor) Process(jsonInput string) error {
	req, err := NewRequest(jsonInput)

	if err != nil {
		log.Printf("Invalid request: %v", req)
		return err
	}

	return rp.Send(req)
}

func (rp *RequestProcessor) Send(req *Request) error {

	targetURL, err := url.Parse(rp.Config.Endpoint.Uri)

	if err != nil {
		log.Print(err)
		return err
	}

	httpReq := &http.Request{}
	httpReq.Method = "PURGE"
	httpReq.Host = req.Host
	httpReq.Header = make(http.Header)
	httpReq.URL = targetURL

	switch req.Command {
	case "purge":
		targetURL.Path = req.Path

		log.Printf("Purging path %s from %s", req.Path, req.Host)

	case "ban":
		httpReq.Method = "BAN"
		targetURL.Path = "/"
		httpReq.Header.Add("X-Ban-Expression", req.Expression)

		log.Printf("Banning with expression %s", req.Expression)

	case "ban.url":
		httpReq.Method = "BAN"
		targetURL.Path = "/" + req.Pattern

		log.Printf("Banning URL %s from %s", req.Pattern, req.Host)

	case "xkey":
		for _, t := range req.Keys {
			httpReq.Header.Add(rp.Config.Endpoint.XkeyHeader, t)
		}

		log.Printf("Purging tags %s from %s", strings.Join(req.Keys, ", "), req.Host)

	case "xkey.soft":
		for _, t := range req.Keys {
			httpReq.Header.Add(rp.Config.Endpoint.SoftXkeyHeader, t)
		}

		log.Printf("Soft-purging tags %s from %s", strings.Join(req.Keys, ", "), req.Host)
	}

	client := &http.Client{
		Timeout: time.Second * 5,
	}

	_, err = client.Do(httpReq)

	if err != nil {
		log.Printf("Sending request failed: %v", err)
		return err
	}

	return nil
}

func NewRequestProcessor(options Options) *RequestProcessor {
	rp := RequestProcessor{options}
	return &rp
}
