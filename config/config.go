package config

import (
	"fmt"
	"io/ioutil"

	"github.com/ghodss/yaml"
)

// Config contains configuration data
type Config struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Display string            `json:"display"`
	Flags   map[string]string `json:"flags"`
}

// New creates a default configuration object
var New Config

func newConfig() Config {
	flags := make(map[string]string)
	flags["use_color"] = "true"
	flags["verify_tls"] = "true"

	return Config{Flags: flags}
}

// UseColor adds a switch for whether or not to colorize the output
func (gc Config) UseColor() bool {
	return gc.Flags["use_color"] != "false"
}

// TLSVerify determines whether or not to verify that a TLS cert is valid
func (gc Config) TLSVerify() bool {
	return gc.Flags["verify_tls"] != "false"
}

func init() {
	New = newConfig()
}

// LoadConfiguration builds a configuration object based on the fileName passed
func LoadConfiguration(fileName string) (Config, error) {
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		// If the file wasn't found and it's just the default, don't worry about it.
		if fileName == ".gulp.yml" {
			return New, nil
		}

		return New, fmt.Errorf("Could not load configuration '%s'", fileName)
	}

	var gulpConfig Config
	if yaml.Unmarshal(dat, &gulpConfig) != nil {
		return New, fmt.Errorf("Could not parse configuration")
	}

	return gulpConfig, nil
}
