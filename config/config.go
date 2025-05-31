package config

import (
	"fmt"
	"os"
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
	ClientAuth ClientAuth        `json:"client_auth"`
	Flags      ConfigFlags       `json:"flags"`
}

// ClientAuth leads to files with PEM-encoded data tied to client cert authentication
type ClientAuth struct {
	Cert     string `json:"cert"`
	Key      string `json:"key"`
	CA       string `json:"ca"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// ConfigFlags contains valid configuration flags
// These are strings not bool bc otherwise we don't know if the config file is missing the flag or is set to false
type ConfigFlags struct {
	FollowRedirects string `json:"follow_redirects"`
	UseColor        string `json:"use_color"`
	VerifyTLS       string `json:"verify_tls"`
}

// DefaultTimeout is 5 minutes (300 seconds)
const DefaultTimeout = 300

// New creates a default configuration object
var New *Config

func newConfig() *Config {
	flags := ConfigFlags{
		FollowRedirects: "true",
		UseColor:        "true",
		VerifyTLS:       "true",
	}

	return &Config{Flags: flags}
}

// FollowRedirects determines whether or not to follow 301/302 redirects
func (gc *Config) FollowRedirects() bool {
	return gc.Flags.FollowRedirects != "false"
}

// UseColor adds a switch for whether or not to colorize the output
func (gc *Config) UseColor() bool {
	return gc.Flags.UseColor != "false"
}

// VerifyTLS determines whether or not to verify that a TLS cert is valid
func (gc *Config) VerifyTLS() bool {
	return gc.Flags.VerifyTLS != "false"
}

func (gc *ClientAuth) UseAuth() bool {
	return strings.TrimSpace(gc.Cert) != "" && strings.TrimSpace(gc.Key) != ""
}

// UseBasicAuth determines whether or not to use basic authentication
func (gc *ClientAuth) UseBasicAuth() bool {
	return strings.TrimSpace(gc.Username) != "" && strings.TrimSpace(gc.Password) != ""
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
	dat, err := os.ReadFile(fileName)
	if err != nil {
		// If the file wasn't found and it's just the default, don't worry about it.
		if fileName == ".gulp.yml" {
			return New, nil
		}

		return nil, fmt.Errorf("could not load configuration '%s'", fileName)
	}

	var gulpConfig *Config
	if err := yaml.Unmarshal(dat, &gulpConfig); err != nil {
		return nil, buildConfigurationError(fileName, err)
	}

	// Clean up field padding
	cleanupConfigurationFields(gulpConfig)

	return gulpConfig, nil
}

// buildConfigurationError creates a detailed error message with examples
func buildConfigurationError(fileName string, parseErr error) error {
	return fmt.Errorf(`could not parse configuration file '%s': %v

Example of valid YAML configuration:
---
# Basic configuration
url: https://api.example.com
timeout: "30"

# Optional request headers
headers:
  Authorization: Bearer your-token-here
  X-Custom-Header: some-value

# Optional client authentication (supports both certificate and basic auth)
client_auth:
  # Client certificate authentication
  cert: /path/to/client-cert.pem
  key: /path/to/client-key.pem
  ca: /path/to/ca-cert.pem
  # Basic authentication
  username: your-username
  password: your-password

# Optional flags (all default to true)
flags:
  follow_redirects: "true"
  use_color: "true"
  verify_tls: "true"

# Optional display setting
display: verbose  # or "status-code-only"
---

For more examples, see: https://github.com/thoom/gulp#configuration`, fileName, parseErr)
}

// cleanupConfigurationFields trims whitespace from all string fields in the configuration
func cleanupConfigurationFields(config *Config) {
	// Clean up client auth fields
	config.ClientAuth.Cert = strings.TrimSpace(config.ClientAuth.Cert)
	config.ClientAuth.Key = strings.TrimSpace(config.ClientAuth.Key)
	config.ClientAuth.CA = strings.TrimSpace(config.ClientAuth.CA)
	config.ClientAuth.Username = strings.TrimSpace(config.ClientAuth.Username)
	config.ClientAuth.Password = strings.TrimSpace(config.ClientAuth.Password)
}
