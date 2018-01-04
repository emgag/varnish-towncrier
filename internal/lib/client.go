package lib

import (
	"encoding/json"
)

// Client is the client used to connect to redis and send pubsub messages
type Client struct {
	Options Options
}

// Ban issues a ban request, expression being a complete VCL ban expression
func (c *Client) Ban(channels []string, host string, value []string) error {
	return c.Do(channels, Request{Command: "ban", Host: host, Value: value})
}

// BanURL issues a ban request with pattern matching the URL
func (c *Client) BanURL(channels []string, host string, value []string) error {
	return c.Do(channels, Request{Command: "ban.url", Host: host, Value: value})
}

// Purge issues a purge request for the supplied path
func (c *Client) Purge(channels []string, host string, value []string) error {
	return c.Do(channels, Request{Command: "purge", Host: host, Value: value})
}

// Xkey issues a purge request for supplied surrogate keys
func (c *Client) Xkey(channels []string, host string, value []string) error {
	return c.Do(channels, Request{Command: "xkey", Host: host, Value: value})
}

// XkeySoft issues a soft-purge request for supplied surrogate keys
func (c *Client) XkeySoft(channels []string, host string, value []string) error {
	return c.Do(channels, Request{Command: "xkey.soft", Host: host, Value: value})
}

// Do sends a request to supplied pubsub channels
func (c *Client) Do(channels []string, req Request) error {

	redis, err := NewRedisConn(c.Options)

	if err != nil {
		return err
	}

	message, _ := json.Marshal(req)

	for _, channel := range channels {
		redis.Do("publish", channel, string(message))
	}

	return nil
}

// NewClient creates a new instance of Client
func NewClient(options Options) *Client {
	return &Client{Options: options}
}
