package cmd

import (
	"fmt"

	"github.com/emgag/varnish-towncrier/internal/lib"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of varnish-towncrier",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("varnish-towncrier -- %s\n", lib.Version)
	},
}
