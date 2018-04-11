package main

import (
	"fmt"
	"io/ioutil"
	
	"github.com/ghodss/yaml"
)

// GulpConfig contains configuration data
type GulpConfig struct {
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Display string            `json:"display"`
	Flags   map[string]string `json:"flags"`
}

var NewConfig GulpConfig

// NewConfig creates a new instance of GulpConfig
func newConfig() GulpConfig {
	flags := make(map[string]string)
	flags["use_color"] = "true"
	flags["verify_tls"] = "true"

	return GulpConfig{Flags: flags}
}

// UseColor adds a switch for whether or not to colorize the output
func (gc GulpConfig) UseColor() bool {
	return gc.Flags["use_color"] != "false"
}

// TLSVerify determines whether or not to verify that a TLS cert is valid
func (gc GulpConfig) TLSVerify() bool {
	return gc.Flags["verify_tls"] != "false"
}

func init() {
	NewConfig = newConfig()
}

func LoadConfiguration(fileName string) GulpConfig {
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		// If the file wasn't found and it's just the default, don't worry about it.
		if fileName == ".gulp.yml" {
			return config
		}

		ExitErr(fmt.Sprintf("Could not load configuration '%s'", fileName), nil)
	}

	var gulpConfig GulpConfig
	if yaml.Unmarshal(dat, &gulpConfig) != nil {
		ExitErr("Could not parse configuration", nil)
	}

	return gulpConfig
}
