package cmd

import (
	"fmt"

	"github.com/emgag/varnish-broadcast/internal/lib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of varnish-broadcast",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("varnish-broadcast -- %s\n", lib.Version)
	},
}
