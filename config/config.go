package config

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/ghodss/yaml"
)

// Config contains configuration data
type Config struct {
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers"`
	Display    string            `json:"display"`
	Timeout    string            `json:"timeout"`
	ClientCert ClientCertAuth    `json:"client_cert_auth"`
	Flags      ConfigFlags       `json:"flags"`
}

// ClientCertAuth leads to files with PEM-encoded data tied to client cert authentication
type ClientCertAuth struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
}

// ConfigFlags contains valid configuration flags
type ConfigFlags struct {
	FollowRedirects bool `json:"follow_redirects"`
	UseColor        bool `json:"use_color"`
	VerifyTLS       bool `json:"verify_tls"`
}

// DefaultTimeout is 5 minutes (300 seconds)
const DefaultTimeout = 300

// New creates a default configuration object
var New *Config

func newConfig() *Config {
	flags := ConfigFlags{
		FollowRedirects: true,
		UseColor:        true,
		VerifyTLS:       true,
	}

	return &Config{Flags: flags}
}

// FollowRedirects determines whether or not to follow 301/302 redirects
func (gc *Config) FollowRedirects() bool {
	return gc.Flags.FollowRedirects
}

// UseColor adds a switch for whether or not to colorize the output
func (gc *Config) UseColor() bool {
	return gc.Flags.UseColor
}

// VerifyTLS determines whether or not to verify that a TLS cert is valid
func (gc *Config) VerifyTLS() bool {
	return gc.Flags.VerifyTLS
}

func (gc *Config) UseClientCertAuth() bool {
	return strings.TrimSpace(gc.ClientCert.Cert) != "" && strings.TrimSpace(gc.ClientCert.Key) != ""
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

	// Clean up spaced padding
	if gulpConfig.ClientCert.Cert != "" {
		gulpConfig.ClientCert.Cert = strings.TrimSpace(gulpConfig.ClientCert.Cert)
	}

	// Clean up spaced padding
	if gulpConfig.ClientCert.Key != "" {
		gulpConfig.ClientCert.Key = strings.TrimSpace(gulpConfig.ClientCert.Key)
	}

	return gulpConfig, nil
}
