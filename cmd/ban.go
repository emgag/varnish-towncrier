package cmd

import (
	"log"

	"github.com/emgag/varnish-towncrier/internal/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	clientFlags(banCmd)
	banCmd.Flags().Bool(
		"url",
		false,
		"Submit an URL ban (regular expression pattern matched on path instead of VCL expression)",
	)

	rootCmd.AddCommand(banCmd)
}

var banCmd = &cobra.Command{
	Use:       "ban [flags] expression [expression...]",
	Short:     "Issue ban request to all registered instances",
	Args:      cobra.MinimumNArgs(1),
	ValidArgs: []string{"expression"},
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
		var banFunc func([]string, string, []string) error

		if ret, _ := cmd.Flags().GetBool("url"); ret {
			banFunc = client.BanURL
		} else {
			banFunc = client.Ban
		}

		if err := banFunc(channels, host, args); err != nil {
			log.Fatalf("Error connecting to redis: %s", err)
		}
	},
}
