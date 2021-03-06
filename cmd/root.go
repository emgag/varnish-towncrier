package cmd

import (
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "varnish-towncrier",
	Short: "Distribute cache invalidation requests to a fleet of varnish instances.",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is /etc/varnish-towncrier.yml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName("varnish-towncrier")

	// set defaults for redis
	viper.SetDefault("redis.uri", "redis://127.0.0.1:6379")
	viper.SetDefault("redis.subscribe", []string{"varnish.purge"})

	// set defaults for varnish
	viper.SetDefault("endpoint.uri", "http://127.0.0.1:8080/")
	viper.SetDefault("endpoint.xkeyheader", "x-xkey")
	viper.SetDefault("endpoint.softxkeyheader", "x-xkey-soft")
	viper.SetDefault("endpoint.banheader", "x-ban-expression")
	viper.SetDefault("endpoint.banurlheader", "x-ban-url")

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("/etc")
		viper.AddConfigPath("$HOME/.varnish-towncrier")
		viper.AddConfigPath(".")
	}

	viper.SetEnvPrefix("vt")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetTypeByDefaultValue(true)
	viper.AutomaticEnv()

	// if a config file is found, read it in.
	err := viper.ReadInConfig()

	if err != nil {
		log.Printf("Could not open config file: %v", err)
		log.Printf("Using environment and default config only")
	}

}

func clientFlags(cmd *cobra.Command) {
	cmd.Flags().String("host", "", "HTTP Host")
	cmd.Flags().String(
		"channel",
		"",
		"Pubsub channel to publish message to (defaults to all configured channels)",
	)
}
