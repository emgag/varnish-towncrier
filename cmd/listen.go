package cmd

import (
	"log"

	"github.com/emgag/varnish-broadcast/internal/lib"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(listenCmd)
}

var listenCmd = &cobra.Command{
	Use:   "listen",
	Short: "Listen for incoming invalidation requests",
	Run: func(cmd *cobra.Command, args []string) {
		options := lib.Options{}
		err := viper.Unmarshal(&options)

		if err != nil {
			log.Fatal(err)
		}

		listener := lib.NewListener(options)

		if err := listener.Listen(); err != nil {
			log.Fatal(err)
		}
	},
}
