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
	assert.False(New.ClientAuth.UseAuth())
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
	assert.Contains(errStr, "client_auth:")
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
client_cert_auth:
  cert:     
  key:  
  `), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal(config.Flags.FollowRedirects, "false")
	assert.Equal(config.Flags.UseColor, "false")
	assert.Equal(config.Flags.VerifyTLS, "false")
	assert.Empty(config.ClientAuth.Cert)
	assert.Empty(config.ClientAuth.Key)
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
client_auth:
  cert: someFile.pem
  key: CLIENT_CERT_KEY
  `), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal(config.Flags.FollowRedirects, "true")
	assert.Equal(config.Flags.UseColor, "true")
	assert.Equal(config.Flags.VerifyTLS, "true")
	assert.Equal("someFile.pem", config.ClientAuth.Cert)
	assert.Equal("CLIENT_CERT_KEY", config.ClientAuth.Key)
}
