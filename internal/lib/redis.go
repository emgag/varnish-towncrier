package lib

import (
	"time"

	"github.com/gomodule/redigo/redis"
)

// NewRedisConn creates new redis connection
func NewRedisConn(options Options) (redis.Conn, error) {
	var dialOptions []redis.DialOption
	{
		redis.DialConnectTimeout(5 * time.Second)
		redis.DialReadTimeout(2 * time.Second)
		redis.DialWriteTimeout(2 * time.Second)
	}

	if options.Redis.Password != "" {
		dialOptions = append(dialOptions, redis.DialPassword(options.Redis.Password))
	}

	return redis.DialURL(options.Redis.URI, dialOptions...)
}
