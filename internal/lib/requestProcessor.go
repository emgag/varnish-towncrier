package lib

import (
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/emgag/varnish-towncrier/internal/lib/version"
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

	if req.Host != "" {
		httpReq.Host = req.Host
	} else {
		httpReq.Host = targetURL.Host
	}

	httpReq.Header = make(http.Header)
	httpReq.Header.Set("User-Agent", "varnish-towncrier/"+version.Version)
	httpReq.URL = targetURL

	client := &http.Client{
		Timeout: time.Second * 5,
	}

	// xkey and xkey.soft commands allow submitting multiple values (surrogate keys) in a single request,
	// ban, ban.url and purge need to issue multiple http requests.
	switch req.Command {
	case "purge":
		for _, path := range req.Value {
			targetURL.Path = path

			log.Printf("Purging path %s from %s", path, httpReq.Host)

			_, err = client.Do(httpReq)

			if err != nil {
				log.Printf("Sending request failed: %v", err)
				return err
			}
		}

	case "ban":
		fallthrough
	case "ban.url":

		headerMap := map[string]struct {
			Header string
			Status string
		}{
			"ban":     {rp.Config.Endpoint.BanHeader, "Banning with expression"},
			"ban.url": {rp.Config.Endpoint.BanURLHeader, "Banning URL"},
		}

		httpReq.Method = "BAN"

		for _, expression := range req.Value {
			httpReq.Header.Set(headerMap[req.Command].Header, expression)

			log.Printf("%s %s from %s", headerMap[req.Command].Status, expression, httpReq.Host)

			_, err = client.Do(httpReq)

			if err != nil {
				log.Printf("Sending request failed: %v", err)
				return err
			}
		}

	case "xkey":
		fallthrough
	case "xkey.soft":
		headerMap := map[string]struct {
			Header string
			Status string
		}{
			"xkey":      {rp.Config.Endpoint.XkeyHeader, "Purging"},
			"xkey.soft": {rp.Config.Endpoint.SoftXkeyHeader, "Soft-Purging"},
		}

		for _, t := range req.Value {
			httpReq.Header.Add(headerMap[req.Command].Header, t)
		}

		log.Printf(
			"%s tags %s from %s",
			headerMap[req.Command].Status,
			strings.Join(req.Value, ", "),
			httpReq.Host,
		)

		_, err = client.Do(httpReq)

		if err != nil {
			log.Printf("Sending request failed: %v", err)
			return err
		}
	}

	return nil
}

// NewRequestProcessor creates a new RequestProcessor
func NewRequestProcessor(options Options) *RequestProcessor {
	rp := RequestProcessor{options}
	return &rp
}
