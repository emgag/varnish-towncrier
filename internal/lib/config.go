package lib

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Options contains all settings read from the configuration file
type Options struct {
	Redis struct {
		URI       string   `yaml:"uri"`
		Password  string   `yaml:"password,omitempty"`
		Subscribe []string `yaml:"subscribe"`
	}
	Endpoint struct {
		URI            string `yaml:"uri"`
		XkeyHeader     string `yaml:"xkeyheader"`
		SoftXkeyHeader string `yaml:"softxkeyheader"`
	}
}

// LoadConfig loads settings from the configuration file
func LoadConfig(Filename string) (Options, error) {
	options := Options{}

	// read option file
	config, err := ioutil.ReadFile(Filename)

	if err != nil {
		return options, err
	}

	err = yaml.Unmarshal(config, &options)

	if err != nil {
		return options, err
	}

	return options, nil
}
