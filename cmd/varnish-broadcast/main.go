package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

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
		Usage: "HTTP Host (required)",
	}

	pubsubChannelFlag := cli.StringFlag{
		Name:  "channel",
		Usage: "Pubsub channel to publish message to (defaults to all configured channels)",
	}

	app.Commands = []cli.Command{
		{
			Name:      "listen",
			Usage:     "Listen for incoming invalidation requests",
			Category:  "Agent",
			ArgsUsage: " ",
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
			Before: func(c *cli.Context) error {
				messages := []string{}

				if !c.IsSet(hostFlag.Name) {
					messages = append(messages, "host is required")
				}

				if !c.Args().Present() {
					messages = append(messages, "expression is required")
				}

				if len(messages) > 0 {
					fmt.Printf("%s\n\n", strings.Join(messages, ", "))
					return errors.New(strings.Join(messages, ", "))
				}

				return nil
			},
			Action: func(c *cli.Context) error {
				options, err := lib.LoadConfig(c.GlobalString("config"))

				if err != nil {
					log.Fatal(err)
				}

				channels := []string{}

				if c.String(pubsubChannelFlag.Name) != "" {
					channels = []string{c.String(pubsubChannelFlag.Name)}
				} else {
					channels = options.Redis.Subscribe
				}

				client := lib.NewClient(options)

				if c.Bool("url") {
					err = client.BanURL(channels, c.String(hostFlag.Name), c.Args().First())
				} else {
					err = client.Ban(channels, c.String(hostFlag.Name), c.Args().First())
				}

				if err != nil {
					log.Fatalf("Error connecting to redis: %s", err)
				}

				return err
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
			Before: func(c *cli.Context) error {
				messages := []string{}

				if !c.IsSet(hostFlag.Name) {
					messages = append(messages, "host is required")
				}

				if !c.Args().Present() {
					messages = append(messages, "path is required")
				}

				if len(messages) > 0 {
					fmt.Printf("%s\n\n", strings.Join(messages, ", "))
					return errors.New(strings.Join(messages, ", "))
				}

				return nil
			},
			Action: func(c *cli.Context) error {
				options, err := lib.LoadConfig(c.GlobalString("config"))

				if err != nil {
					log.Fatal(err)
				}

				channels := []string{}

				if c.String(pubsubChannelFlag.Name) != "" {
					channels = []string{c.String(pubsubChannelFlag.Name)}
				} else {
					channels = options.Redis.Subscribe
				}

				client := lib.NewClient(options)
				err = client.Purge(channels, c.String(hostFlag.Name), c.Args().First())

				if err != nil {
					log.Fatalf("Error connecting to redis: %s", err)
				}

				return err
			},
		},
		{
			Name:      "xkey",
			Usage:     "Invalidate selected surrogate keys on all registered instances",
			Category:  "Client",
			ArgsUsage: "keys...",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "soft",
					Usage: "Issue a soft ban",
				},
				hostFlag,
				pubsubChannelFlag,
			},
			Before: func(c *cli.Context) error {
				messages := []string{}

				if !c.IsSet(hostFlag.Name) {
					messages = append(messages, "host is required")
				}

				if !c.Args().Present() {
					messages = append(messages, "keys are required")
				}

				if len(messages) > 0 {
					fmt.Printf("%s\n\n", strings.Join(messages, ", "))
					return errors.New(strings.Join(messages, ", "))
				}

				return nil
			},
			Action: func(c *cli.Context) error {
				options, err := lib.LoadConfig(c.GlobalString("config"))

				if err != nil {
					log.Fatal(err)
				}

				channels := []string{}

				if c.String(pubsubChannelFlag.Name) != "" {
					channels = []string{c.String(pubsubChannelFlag.Name)}
				} else {
					channels = options.Redis.Subscribe
				}

				client := lib.NewClient(options)

				if c.Bool("soft") {
					err = client.XkeySoft(channels, c.String(hostFlag.Name), c.Args())
				} else {
					err = client.Xkey(channels, c.String(hostFlag.Name), c.Args())
				}

				if err != nil {
					log.Fatalf("Error connecting to redis: %s", err)
				}

				return err
			},
		},
	}

	app.Run(os.Args)

}
