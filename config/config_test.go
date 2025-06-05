package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Make sure that the default config is set up to use color and verify TLS
func TestNewConfig(t *testing.T) {
	assert := assert.New(t)

	assert.True(New.FollowRedirects())
	assert.True(New.UseColor())
	assert.True(New.VerifyTLS())
	assert.False(New.Auth.UseAuth())
	assert.Equal(DefaultTimeout, New.GetTimeout())
}

func TestLoadConfigurationDefault(t *testing.T) {
	assert := assert.New(t)

	config, _ := LoadConfiguration(".gulp.yml")
	assert.NotNil(config)
	assert.EqualValues(config, New)
}

func TestLoadConfigurationMissing(t *testing.T) {
	assert := assert.New(t)

	_, err := LoadConfiguration("invalidFile.yml")
	assert.NotNil(err)
	assert.Contains(fmt.Sprintf("%s", err), "could not load configuration")
}

func TestLoadConfigurationNoParse(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := os.CreateTemp(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	// Write invalid YAML content (binary data that can't be parsed as YAML)
	os.WriteFile(testFile.Name(), []byte{255, 253}, 0644)
	_, err := LoadConfiguration(testFile.Name())
	assert.NotNil(err)

	errStr := fmt.Sprintf("%s", err)
	// Verify the error message includes the filename
	assert.Contains(errStr, testFile.Name())
	// Verify it still contains the original error indication
	assert.Contains(errStr, "could not parse configuration")
	// Verify it includes helpful example
	assert.Contains(errStr, "Example of valid YAML configuration")
	assert.Contains(errStr, "url: https://api.example.com")
	assert.Contains(errStr, "auth:")
	assert.Contains(errStr, "https://github.com/thoom/gulp#configuration")
}

func TestLoadConfigurationInvalidYAMLSyntax(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := os.CreateTemp(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	// Write YAML with syntax error (invalid indentation)
	invalidYaml := `
url: https://api.example.com
headers:
  Authorization: Bearer token
    X-Invalid-Indent: this is incorrectly indented
flags:
  use_color: true
`
	os.WriteFile(testFile.Name(), []byte(invalidYaml), 0644)
	_, err := LoadConfiguration(testFile.Name())
	assert.NotNil(err)

	errStr := fmt.Sprintf("%s", err)
	// Verify the error message includes the filename
	assert.Contains(errStr, testFile.Name())
	// Verify it contains helpful information
	assert.Contains(errStr, "could not parse configuration")
	assert.Contains(errStr, "Example of valid YAML configuration")
	assert.Contains(errStr, "# Basic configuration")
}

func TestLoadConfigurationMissingTimeout(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := os.CreateTemp(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	os.WriteFile(testFile.Name(), []byte("url: some_url"), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal("some_url", config.URL)
	assert.Equal(DefaultTimeout, config.GetTimeout())
}

func TestLoadConfigurationFlagsOK(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := os.CreateTemp(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	os.WriteFile(testFile.Name(), []byte("url: some_url"), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal("some_url", config.URL)
	assert.True(config.FollowRedirects())
	assert.True(config.UseColor())
	assert.True(config.VerifyTLS())
}

func TestLoadConfigurationFoundTimeout(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := os.CreateTemp(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	os.WriteFile(testFile.Name(), []byte("url: some_url\ntimeout: 100"), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal("some_url", config.URL)
	assert.Equal(100, config.GetTimeout())
}

func TestLoadConfigurationInvalidTimeout(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := os.CreateTemp(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	os.WriteFile(testFile.Name(), []byte("url: some_url\ntimeout: abc"), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal("some_url", config.URL)
	assert.Equal(DefaultTimeout, config.GetTimeout())
}

func TestLoadConfigurationLoadFlagsNegative(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := os.CreateTemp(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	os.WriteFile(testFile.Name(), []byte(`
flags:
  follow_redirects: false
  use_color: false
  verify_tls: false
auth:
  certificate:
    cert: ""
    key: ""
  `), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal(config.Flags.FollowRedirects, false)
	assert.Equal(config.Flags.UseColor, false)
	assert.Equal(config.Flags.VerifyTLS, false)
	assert.Empty(config.Auth.Certificate.Cert)
	assert.Empty(config.Auth.Certificate.Key)
}

func TestLoadConfigurationLoadFlagsPositive(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := os.CreateTemp(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	os.WriteFile(testFile.Name(), []byte(`
flags:
  follow_redirects: true
  use_color: true
  verify_tls: true
auth:
  certificate:
    cert: someFile.pem
    key: CLIENT_CERT_KEY
  `), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal(config.Flags.FollowRedirects, true)
	assert.Equal(config.Flags.UseColor, true)
	assert.Equal(config.Flags.VerifyTLS, true)
	assert.Equal("someFile.pem", config.Auth.Certificate.Cert)
	assert.Equal("CLIENT_CERT_KEY", config.Auth.Certificate.Key)
}

func TestLoadConfigurationBasicAuth(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := os.CreateTemp(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	os.WriteFile(testFile.Name(), []byte(`
auth:
  basic:
    username: testuser
    password: testpass
  `), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal("testuser", config.Auth.Basic.Username)
	assert.Equal("testpass", config.Auth.Basic.Password)
	assert.True(config.Auth.UseBasicAuth())
}

func TestBuildConfigurationError(t *testing.T) {
	assert := assert.New(t)

	err := buildConfigurationError("test.yml", fmt.Errorf("invalid YAML"))
	assert.Contains(err.Error(), "could not parse configuration file 'test.yml'")
	assert.Contains(err.Error(), "invalid YAML")
	assert.Contains(err.Error(), "Example of valid YAML configuration")
	assert.Contains(err.Error(), "github.com/thoom/gulp#configuration")
}

func TestCleanupConfigurationFields(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Auth: AuthConfig{
			Certificate: CertAuthConfig{
				Cert: "  /path/to/cert.pem  ",
				Key:  "\t/path/to/key.pem\t",
				CA:   "  /path/to/ca.pem  ",
			},
			Basic: BasicAuthConfig{
				Username: " user ",
				Password: " pass ",
			},
		},
	}

	cleanupConfigurationFields(config)

	assert.Equal("/path/to/cert.pem", config.Auth.Certificate.Cert)
	assert.Equal("/path/to/key.pem", config.Auth.Certificate.Key)
	assert.Equal("/path/to/ca.pem", config.Auth.Certificate.CA)
	assert.Equal("user", config.Auth.Basic.Username)
	assert.Equal("pass", config.Auth.Basic.Password)
}

func TestAuthConfigUseAuth(t *testing.T) {
	// Test with cert and key
	auth := AuthConfig{
		Certificate: CertAuthConfig{
			Cert: "cert.pem",
			Key:  "key.pem",
		},
	}
	assert.True(t, auth.UseAuth())

	// Test without cert
	auth.Certificate.Cert = ""
	assert.False(t, auth.UseAuth())

	// Test without key
	auth.Certificate.Cert = "cert.pem"
	auth.Certificate.Key = ""
	assert.False(t, auth.UseAuth())
}

func TestAuthConfigUseBasicAuth(t *testing.T) {
	// Test with username and password
	auth := AuthConfig{
		Basic: BasicAuthConfig{
			Username: "user",
			Password: "pass",
		},
	}
	assert.True(t, auth.UseBasicAuth())

	// Test without username
	auth.Basic.Username = ""
	assert.False(t, auth.UseBasicAuth())

	// Test without password
	auth.Basic.Username = "user"
	auth.Basic.Password = ""
	assert.False(t, auth.UseBasicAuth())
}

func TestGetAuthConfig(t *testing.T) {
	config := &Config{
		Auth: AuthConfig{
			Basic: BasicAuthConfig{
				Username: "testuser",
				Password: "testpass",
			},
			Certificate: CertAuthConfig{
				Cert: "cert.pem",
				Key:  "key.pem",
				CA:   "ca.pem",
			},
		},
	}

	auth := config.GetAuthConfig()
	assert.Equal(t, "testuser", auth.Basic.Username)
	assert.Equal(t, "testpass", auth.Basic.Password)
	assert.Equal(t, "cert.pem", auth.Certificate.Cert)
	assert.Equal(t, "key.pem", auth.Certificate.Key)
	assert.Equal(t, "ca.pem", auth.Certificate.CA)
}

// New tests for 0% coverage functions

func TestGetMethod(t *testing.T) {
	assert := assert.New(t)

	// Test with empty method (should return default)
	config := &Config{Method: ""}
	assert.Equal("GET", config.GetMethod())

	// Test with configured method
	config.Method = "POST"
	assert.Equal("POST", config.GetMethod())

	// Test with other methods
	config.Method = "PUT"
	assert.Equal("PUT", config.GetMethod())

	config.Method = "DELETE"
	assert.Equal("DELETE", config.GetMethod())
}

func TestGetOutput(t *testing.T) {
	assert := assert.New(t)

	// Test with empty output (should return default)
	config := &Config{Output: ""}
	assert.Equal("body", config.GetOutput())

	// Test with configured output modes
	config.Output = "status"
	assert.Equal("status", config.GetOutput())

	config.Output = "verbose"
	assert.Equal("verbose", config.GetOutput())

	config.Output = "body"
	assert.Equal("body", config.GetOutput())
}

func TestGetRepeatTimes(t *testing.T) {
	assert := assert.New(t)

	// Test with zero/negative times (should return default)
	config := &Config{Repeat: RepeatConfig{Times: 0}}
	assert.Equal(1, config.GetRepeatTimes())

	config.Repeat.Times = -5
	assert.Equal(1, config.GetRepeatTimes())

	// Test with valid times
	config.Repeat.Times = 10
	assert.Equal(10, config.GetRepeatTimes())

	config.Repeat.Times = 100
	assert.Equal(100, config.GetRepeatTimes())
}

func TestGetRepeatConcurrent(t *testing.T) {
	assert := assert.New(t)

	// Test with zero/negative concurrent (should return default)
	config := &Config{Repeat: RepeatConfig{Concurrent: 0}}
	assert.Equal(1, config.GetRepeatConcurrent())

	config.Repeat.Concurrent = -3
	assert.Equal(1, config.GetRepeatConcurrent())

	// Test with valid concurrent values
	config.Repeat.Concurrent = 5
	assert.Equal(5, config.GetRepeatConcurrent())

	config.Repeat.Concurrent = 20
	assert.Equal(20, config.GetRepeatConcurrent())
}

func TestFollowRedirectsRequestOverride(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Flags:   ConfigFlags{FollowRedirects: true},
		Request: RequestConfig{},
	}

	// Test global flag default
	assert.True(config.FollowRedirects())

	// Test NoRedirects override (pointer to true means NO redirects)
	noRedirects := true
	config.Request.NoRedirects = &noRedirects
	assert.False(config.FollowRedirects())

	// Test FollowRedirects override (pointer to false when global is true)
	followRedirects := false
	config.Request.NoRedirects = nil
	config.Request.FollowRedirects = &followRedirects
	assert.False(config.FollowRedirects())

	// Test FollowRedirects override (pointer to true)
	followRedirects = true
	config.Request.FollowRedirects = &followRedirects
	assert.True(config.FollowRedirects())

	// Test with global flag false and request override true
	config.Flags.FollowRedirects = false
	config.Request.NoRedirects = nil
	config.Request.FollowRedirects = &followRedirects // still true
	assert.True(config.FollowRedirects())
}

func TestVerifyTLSRequestOverride(t *testing.T) {
	assert := assert.New(t)

	config := &Config{
		Flags:   ConfigFlags{VerifyTLS: true},
		Request: RequestConfig{Insecure: false},
	}

	// Test global flag default
	assert.True(config.VerifyTLS())

	// Test request insecure override
	config.Request.Insecure = true
	assert.False(config.VerifyTLS())

	// Test with global flag false
	config.Flags.VerifyTLS = false
	config.Request.Insecure = false
	assert.False(config.VerifyTLS())
}
