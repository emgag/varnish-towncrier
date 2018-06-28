package cmd

import (
	"fmt"

	"github.com/emgag/varnish-towncrier/internal/lib/version"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of varnish-towncrier",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("varnish-towncrier %s -- %s\n", version.Version, version.Commit)
	},
}
