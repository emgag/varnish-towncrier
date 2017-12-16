package lib

import (
	"io"
	"log"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Listener struct {
	Options Options
}

func NewListener(options Options) *Listener {
	l := Listener{}
	l.Options = options

	return &l
}

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

		log.Printf("Connected to %s", l.Options.Redis.Uri)

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
