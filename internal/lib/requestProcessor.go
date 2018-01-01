package lib

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestProcessor converts an incoming pubsub message on redis to a purge request to be sent to varnish
type RequestProcessor struct {
	Config Options
}

// Process parses the request and sends it to varnish
func (rp *RequestProcessor) Process(jsonInput string) error {
	req, err := NewRequest(jsonInput)

	if err != nil {
		log.Printf("Invalid request: %v", req)
		return err
	}

	return rp.Send(req)
}

// Send sends a purge request to varnish
func (rp *RequestProcessor) Send(req *Request) error {

	targetURL, err := url.Parse(rp.Config.Endpoint.URI)

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

// NewRequestProcessor creates a new RequestProcessor
func NewRequestProcessor(options Options) *RequestProcessor {
	rp := RequestProcessor{options}
	return &rp
}
