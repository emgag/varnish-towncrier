package main

import (
	"log"
	"os"

	"github.com/emgag/varnish-broadcast/internal/lib"
	"github.com/urfave/cli"
)

func main() {

	app := cli.NewApp()
	app.Name = "varnish-broadcast"
	app.Usage = "Distribute cache invalidation requests to a fleet of varnish instances."
	app.Version = "0.1"

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: "/etc/varnish-broadcast.yml",
			Usage: "Load configuration from `FILE`",
		},
	}

	hostFlag := cli.StringFlag{
		Name:  "host",
		Usage: "HTTP Host",
	}

	pubsubChannelFlag := cli.StringFlag{
		Name:  "channel",
		Usage: "Pubsub channel to publish message to (defaults to all configured channels)",
	}

	app.Commands = []cli.Command{
		{
			Name:     "listen",
			Usage:    "Listen for incoming invalidation requests",
			Category: "Agent",
			Action: func(c *cli.Context) error {
				options, err := lib.LoadConfig(c.GlobalString("config"))

				if err != nil {
					log.Fatal(err)
				}

				listener := lib.NewListener(options)

				if err := listener.Listen(); err != nil {
					log.Fatal(err)
				}

				return nil
			},
		},
		{
			Name:      "ban",
			Usage:     "Issue ban request to all registered instances",
			Category:  "Client",
			ArgsUsage: "expression",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "url",
					Usage: "Submit an URL ban (regular expression pattern matched on path instead of VCL expression) ",
				},
				hostFlag,
				pubsubChannelFlag,
			},
			Action: func(c *cli.Context) error {
				log.Fatal("ban: TBD")

				return nil
			},
		},
		{
			Name:      "purge",
			Usage:     "Issue purge request to all registered instances",
			Category:  "Client",
			ArgsUsage: "path",
			Flags: []cli.Flag{
				hostFlag,
				pubsubChannelFlag,
			},
			Action: func(c *cli.Context) error {
				log.Fatal("purge: TBD")

				return nil
			},
		},
		{
			Name:      "xkey",
			Usage:     "Invalidate selected surrogate keys on all registered instances",
			Category:  "Client",
			ArgsUsage: "tags...",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "soft",
					Usage: "Issue a soft ban",
				},
				hostFlag,
				pubsubChannelFlag,
			},
			Action: func(c *cli.Context) error {
				log.Fatal("xkey: TBD")

				return nil
			},
		},
	}

	app.Run(os.Args)

}
