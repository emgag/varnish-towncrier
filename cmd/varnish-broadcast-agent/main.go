package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/emgag/varnish-broadcast/internal/app"
)

func main() {

	// "listen" command
	listenCommand := flag.NewFlagSet("listen", flag.ExitOnError)

	if len(os.Args) < 2 {
		fmt.Println("listen subcommand is required")
		os.Exit(1)
	}

	configFilePath := listenCommand.String("config", "/etc/varnish/broadcast-agent.yml", "Config file to use.")

	switch os.Args[1] {
	case "listen":
		listenCommand.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if listenCommand.Parsed() {
		options, err := app.LoadConfig(*configFilePath)

		if err != nil {
			log.Fatal(err)
		}

		listener := app.NewListener(options)

		if err := listener.Listen(); err != nil {
			log.Fatal(err)
		}
	}

}
