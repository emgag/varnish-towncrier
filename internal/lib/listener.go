package lib

import (
	"io"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

// Listener is used to connect to redis pubsub and listen for incoming requests
type Listener struct {
	Options Options
}

// Listen starts listening for incoming requests
func (l *Listener) Listen() error {

	rp := NewRequestProcessor(l.Options)

	for {
		log.Printf("Connecting to redis...")

		c, err := NewRedisConn(l.Options)

		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		defer c.Close()

		log.Printf("Connected to %s", l.Options.Redis.URI)

		psc := redis.PubSubConn{Conn: c}
		psc.Subscribe(redis.Args{}.AddFlat(l.Options.Redis.Subscribe)...)

	Receive:
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				go rp.Process(string(v.Data))
			case redis.Subscription:
				log.Printf("%s: %s (%d)\n", v.Kind, v.Channel, v.Count)
			case error:
				if c.Err() == io.EOF {
					log.Print("Lost connection to redis, reconnecting...")
				} else {
					log.Print(c.Err())
					log.Print(v)
				}

				time.Sleep(5 * time.Second)
				c.Close()
				break Receive
			}
		}
	}

}

// NewListener creates a new Listener instance
func NewListener(options Options) *Listener {
	l := Listener{}
	l.Options = options

	return &l
}
