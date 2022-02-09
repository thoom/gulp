package config

import (
	"fmt"
	"io/ioutil"
	"strconv"

	"github.com/ghodss/yaml"
)

// Config contains configuration data
type Config struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Display string            `json:"display"`
	Timeout string            `json:"timeout"`
	Flags   map[string]string `json:"flags"`
}

// DefaultTimeout is 5 minutes (300 seconds)
const DefaultTimeout = 300

// New creates a default configuration object
var New *Config

func newConfig() *Config {
	flags := make(map[string]string)
	flags["follow_redirects"] = "true"
	flags["use_color"] = "true"
	flags["verify_tls"] = "true"

	return &Config{Flags: flags}
}

// FollowRedirects determines whether or not to follow 301/302 redirects
func (gc *Config) FollowRedirects() bool {
	return gc.Flags["follow_redirects"] != "false"
}

// UseColor adds a switch for whether or not to colorize the output
func (gc *Config) UseColor() bool {
	return gc.Flags["use_color"] != "false"
}

// VerifyTLS determines whether or not to verify that a TLS cert is valid
func (gc *Config) VerifyTLS() bool {
	return gc.Flags["verify_tls"] != "false"
}

// GetTimeout Parses the config string and returns the default if the value wasn't passed
func (gc *Config) GetTimeout() int {
	// If the timeout is empty, just return 300
	if gc.Timeout == "" {
		return DefaultTimeout
	}

	i, err := strconv.Atoi(gc.Timeout)
	if err != nil {
		// For now, if the timeout is not valid, then return 300
		return DefaultTimeout
	}

	return i
}

func init() {
	New = newConfig()
}

// LoadConfiguration builds a configuration object based on the fileName passed
func LoadConfiguration(fileName string) (*Config, error) {
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		// If the file wasn't found and it's just the default, don't worry about it.
		if fileName == ".gulp.yml" {
			return New, nil
		}

		return nil, fmt.Errorf("could not load configuration '%s'", fileName)
	}

	var gulpConfig *Config
	if yaml.Unmarshal(dat, &gulpConfig) != nil {
		return nil, fmt.Errorf("could not parse configuration")
	}

	return gulpConfig, nil
}
