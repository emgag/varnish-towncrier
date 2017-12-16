package lib

import (
	"encoding/json"
)

type Client struct {
	Options Options
}

func (c *Client) Ban(channels []string, host string, expression string) error {
	return c.Do(channels, Request{Command: "ban", Host: host, Expression: expression})
}

func (c *Client) BanURL(channels []string, host string, pattern string) error {
	return c.Do(channels, Request{Command: "ban.url", Host: host, Pattern: pattern})
}

func (c *Client) Purge(channels []string, host string, path string) error {
	return c.Do(channels, Request{Command: "purge", Host: host, Path: path})
}

func (c *Client) Xkey(channels []string, host string, keys []string) error {
	return c.Do(channels, Request{Command: "xkey", Host: host, Keys: keys})
}

func (c *Client) XkeySoft(channels []string, host string, keys []string) error {
	return c.Do(channels, Request{Command: "xkey.soft", Host: host, Keys: keys})
}

func (c *Client) Do(channels []string, req Request) error {

	redis, err := NewRedisConn(c.Options)

	if err != nil {
		return err
	}

	message, err := json.Marshal(req)

	for _, channel := range channels {
		redis.Do("publish", channel, string(message))
	}

	return nil
}

func NewClient(options Options) *Client {
	return &Client{Options: options}
}
