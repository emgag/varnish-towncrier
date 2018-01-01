package cmd

import (
	"log"

	"github.com/emgag/varnish-broadcast/internal/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	clientFlags(purgeCommand)

	rootCmd.AddCommand(purgeCommand)
}

var purgeCommand = &cobra.Command{
	Use:       "purge [flags] --host http.host path [path...]",
	Short:     "Issue purge request to all registered instances",
	Args:      cobra.MinimumNArgs(1),
	ValidArgs: []string{"path"},
	Run: func(cmd *cobra.Command, args []string) {
		options := lib.Options{}
		err := viper.Unmarshal(&options)

		if err != nil {
			log.Fatal(err)
		}

		hostname, _ := cmd.Flags().GetString("hostname")
		channels := []string{}

		if publishChannel, _ := cmd.Flags().GetString("channel"); publishChannel != "" {
			channels = []string{publishChannel}
		} else {
			channels = options.Redis.Subscribe
		}

		client := lib.NewClient(options)

		for _, path := range args {
			if err := client.Purge(channels, hostname, path); err != nil {
				log.Fatalf("Error connecting to redis: %s", err)
			}
		}

	},
}
