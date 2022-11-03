package cmd

import (
	"log"

	"github.com/emgag/varnish-towncrier/internal/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	clientFlags(purgeCommand)

	rootCmd.AddCommand(purgeCommand)
}

var purgeCommand = &cobra.Command{
	Use:       "purge [flags] path [path...]",
	Short:     "Issue purge request to all registered instances",
	Args:      cobra.MinimumNArgs(1),
	ValidArgs: []string{"path"},
	Run: func(cmd *cobra.Command, args []string) {
		options := lib.Options{}
		err := viper.Unmarshal(&options)

		if err != nil {
			log.Fatal(err)
		}

		host, _ := cmd.Flags().GetString("host")
		var channels []string

		if publishChannel, _ := cmd.Flags().GetString("channel"); publishChannel != "" {
			channels = []string{publishChannel}
		} else {
			channels = options.Redis.Subscribe
		}

		client := lib.NewClient(options)

		if err := client.Purge(channels, host, args); err != nil {
			log.Fatalf("Error connecting to redis: %s", err)
		}
	},
}
