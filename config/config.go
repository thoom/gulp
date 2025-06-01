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
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers"`
	Display string            `json:"display"`
	Timeout string            `json:"timeout"`

	// Phase 2: Clean auth structure
	Auth AuthConfig `json:"auth"`

	// Phase 2: Clean boolean flags
	Flags ConfigFlags `json:"flags"`

	// v1.0 additions
	Output  string        `json:"output"`  // New unified output control
	Method  string        `json:"method"`  // Default HTTP method
	Data    DataConfig    `json:"data"`    // Enhanced data input options
	Request RequestConfig `json:"request"` // Request-specific settings
	Repeat  RepeatConfig  `json:"repeat"`  // Load testing configuration
}

// AuthConfig provides clean nested authentication structure
type AuthConfig struct {
	Basic       BasicAuthConfig `json:"basic"`
	Certificate CertAuthConfig  `json:"certificate"`
}

// BasicAuthConfig handles basic authentication
type BasicAuthConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// CertAuthConfig handles certificate authentication
type CertAuthConfig struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
	CA   string `json:"ca"`
}

// UseAuth determines whether certificate authentication should be used
func (ac *AuthConfig) UseAuth() bool {
	cert := strings.TrimSpace(ac.Certificate.Cert)
	key := strings.TrimSpace(ac.Certificate.Key)
	return cert != "" && key != ""
}

// UseBasicAuth determines whether basic authentication should be used
func (ac *AuthConfig) UseBasicAuth() bool {
	username := strings.TrimSpace(ac.Basic.Username)
	password := strings.TrimSpace(ac.Basic.Password)
	return username != "" && password != ""
}

// ConfigFlags contains valid configuration flags with proper boolean types
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

	// Initialize v1.0 defaults
	repeat := RepeatConfig{
		Times:      1,
		Concurrent: 1,
	}

	request := RequestConfig{
		Insecure:        false,
		FollowRedirects: nil, // Use global flag setting
		NoRedirects:     nil, // Use global flag setting
	}

	data := DataConfig{
		Variables: make(map[string]string),
		Form:      make(map[string]string),
		FormMode:  false,
	}

	// Initialize clean auth structure
	auth := AuthConfig{
		Basic:       BasicAuthConfig{},
		Certificate: CertAuthConfig{},
	}

	return &Config{
		Flags:   flags,
		Method:  "GET",
		Output:  "body", // Default to body output
		Data:    data,
		Request: request,
		Repeat:  repeat,
		Auth:    auth,
	}
}

// FollowRedirects determines whether or not to follow 301/302 redirects
func (gc *Config) FollowRedirects() bool {
	// Check request-specific override first
	if gc.Request.NoRedirects != nil && *gc.Request.NoRedirects {
		return false
	}
	if gc.Request.FollowRedirects != nil && *gc.Request.FollowRedirects {
		return true
	}
	// Fall back to global flag
	return gc.Flags.FollowRedirects
}

// UseColor adds a switch for whether or not to colorize the output
func (gc *Config) UseColor() bool {
	return gc.Flags.UseColor
}

// VerifyTLS determines whether or not to verify that a TLS cert is valid
func (gc *Config) VerifyTLS() bool {
	// Check request-specific override first
	if gc.Request.Insecure {
		return false
	}
	// Fall back to global flag
	return gc.Flags.VerifyTLS
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

	// Start with defaults and merge in loaded configuration
	gulpConfig := newConfig()
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

Example of valid YAML configuration (v1.0 with Phase 2 improvements):
---
# Basic configuration
url: https://api.example.com
method: GET                    # Default HTTP method
timeout: "30"                  # Request timeout in seconds
output: body                   # Output mode: body, status, verbose

# Optional request headers
headers:
  Authorization: Bearer your-token-here
  X-Custom-Header: some-value

# Phase 2: Clean authentication structure
auth:
  basic:
    username: api-user
    password: secret-password
  certificate:
    cert: /path/to/client-cert.pem
    key: /path/to/client-key.pem
    ca: /path/to/ca-cert.pem

# Optional data input configuration
data:
  body: "@data.json"           # Request body from file or inline
  template: "@template.json"   # Template file to process
  variables:                   # Template variables
    name: "John Doe"
    environment: "production"
  form:                        # Form data fields
    username: "john"
    email: "john@example.com"
  form_mode: false             # Process stdin as form data

# Optional request-specific settings
request:
  insecure: false              # Disable TLS verification
  follow_redirects: true       # Override global redirect setting

# Optional load testing configuration
repeat:
  times: 1                     # Number of requests to make
  concurrent: 1                # Number of concurrent connections

# Phase 2: Clean boolean flags (no more strings!)
flags:
  follow_redirects: true       # Real boolean
  use_color: true              # Real boolean  
  verify_tls: true             # Real boolean

# Legacy compatibility - still supported but prefer new 'auth' structure
# client_auth:
#   username: api-user
#   password: secret-password
#   cert: /path/to/client-cert.pem
#   key: /path/to/client-key.pem
#   ca: /path/to/ca-cert.pem

# Legacy display setting (use 'output' instead)
# display: verbose
---

For more examples, see: https://github.com/thoom/gulp#configuration`, fileName, parseErr)
}

// cleanupConfigurationFields trims whitespace from all string fields in the configuration
func cleanupConfigurationFields(config *Config) {
	config.URL = strings.TrimSpace(config.URL)
	config.Display = strings.TrimSpace(config.Display)
	config.Timeout = strings.TrimSpace(config.Timeout)
	config.Output = strings.TrimSpace(config.Output)
	config.Method = strings.TrimSpace(config.Method)

	// Clean auth configuration
	config.Auth.Basic.Username = strings.TrimSpace(config.Auth.Basic.Username)
	config.Auth.Basic.Password = strings.TrimSpace(config.Auth.Basic.Password)
	config.Auth.Certificate.Cert = strings.TrimSpace(config.Auth.Certificate.Cert)
	config.Auth.Certificate.Key = strings.TrimSpace(config.Auth.Certificate.Key)
	config.Auth.Certificate.CA = strings.TrimSpace(config.Auth.Certificate.CA)

	// Clean data configuration
	config.Data.Body = strings.TrimSpace(config.Data.Body)
	config.Data.Template = strings.TrimSpace(config.Data.Template)

	// Clean up headers
	for k, v := range config.Headers {
		delete(config.Headers, k)
		config.Headers[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}

	// Clean up variables
	for k, v := range config.Data.Variables {
		delete(config.Data.Variables, k)
		config.Data.Variables[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}

	// Clean up form data
	for k, v := range config.Data.Form {
		delete(config.Data.Form, k)
		config.Data.Form[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
}

// DataConfig handles various data input methods
type DataConfig struct {
	Body      string            `json:"body"`      // Request body content or @file reference
	Template  string            `json:"template"`  // Template file reference
	Variables map[string]string `json:"variables"` // Template variables
	Form      map[string]string `json:"form"`      // Form data fields
	FormMode  bool              `json:"form_mode"` // Process stdin as form data
}

// RequestConfig contains request-specific settings
type RequestConfig struct {
	Insecure        bool  `json:"insecure"`         // Disable TLS verification
	FollowRedirects *bool `json:"follow_redirects"` // Override global setting
	NoRedirects     *bool `json:"no_redirects"`     // Explicit disable redirects
}

// RepeatConfig handles load testing settings
type RepeatConfig struct {
	Times      int `json:"times"`      // Number of times to repeat request
	Concurrent int `json:"concurrent"` // Number of concurrent connections
}

// GetMethod returns the configured HTTP method or default
func (gc *Config) GetMethod() string {
	if gc.Method == "" {
		return "GET"
	}
	return gc.Method
}

// GetOutput returns the configured output mode
func (gc *Config) GetOutput() string {
	if gc.Output == "" {
		return "body"
	}
	return gc.Output
}

// GetRepeatTimes returns the number of times to repeat requests
func (gc *Config) GetRepeatTimes() int {
	if gc.Repeat.Times <= 0 {
		return 1
	}
	return gc.Repeat.Times
}

// GetRepeatConcurrent returns the number of concurrent connections
func (gc *Config) GetRepeatConcurrent() int {
	if gc.Repeat.Concurrent <= 0 {
		return 1
	}
	return gc.Repeat.Concurrent
}

// GetAuthConfig returns the authentication configuration
func (gc *Config) GetAuthConfig() AuthConfig {
	return gc.Auth
}
