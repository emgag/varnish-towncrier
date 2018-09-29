package cmd

import (
	"log"

	"github.com/emgag/varnish-towncrier/internal/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	clientFlags(xkeyCmd)
	xkeyCmd.Flags().Bool("soft", false, "Issue a soft ban")

	rootCmd.AddCommand(xkeyCmd)
}

var xkeyCmd = &cobra.Command{
	Use:       "xkey [flags] key [key...]",
	Short:     "Invalidate selected surrogate keys on all registered instances",
	Args:      cobra.MinimumNArgs(1),
	ValidArgs: []string{"key"},
	Run: func(cmd *cobra.Command, args []string) {
		options := lib.Options{}
		err := viper.Unmarshal(&options)

		if err != nil {
			log.Fatal(err)
		}

		host, _ := cmd.Flags().GetString("host")
		channels := []string{}

		if publishChannel, _ := cmd.Flags().GetString("channel"); publishChannel != "" {
			channels = []string{publishChannel}
		} else {
			channels = options.Redis.Subscribe
		}

		client := lib.NewClient(options)
		var xkeyFunc func([]string, string, []string) error

		if ret, _ := cmd.Flags().GetBool("soft"); ret {
			xkeyFunc = client.XkeySoft
		} else {
			xkeyFunc = client.Xkey
		}

		if err := xkeyFunc(channels, host, args); err != nil {
			log.Fatalf("Error connecting to redis: %s", err)
		}

	},
}
