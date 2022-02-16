package config

import (
	"fmt"
	"io/ioutil"
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

	ioutil.WriteFile(testFile.Name(), []byte{255, 253}, 0644)
	_, err := LoadConfiguration(testFile.Name())
	assert.NotNil(err)
	assert.Contains(fmt.Sprintf("%s", err), "could not parse configuration")
}

func TestLoadConfigurationMissingTimeout(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := os.CreateTemp(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	ioutil.WriteFile(testFile.Name(), []byte("url: some_url"), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal("some_url", config.URL)
	assert.Equal(DefaultTimeout, config.GetTimeout())
}

func TestLoadConfigurationFlagsOK(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := os.CreateTemp(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	ioutil.WriteFile(testFile.Name(), []byte("url: some_url"), 0644)
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

	ioutil.WriteFile(testFile.Name(), []byte("url: some_url\ntimeout: 100"), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal("some_url", config.URL)
	assert.Equal(100, config.GetTimeout())
}

func TestLoadConfigurationInvalidTimeout(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := os.CreateTemp(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	ioutil.WriteFile(testFile.Name(), []byte("url: some_url\ntimeout: abc"), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal("some_url", config.URL)
	assert.Equal(DefaultTimeout, config.GetTimeout())
}

func TestLoadConfigurationLoadFlagsNegative(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := os.CreateTemp(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	ioutil.WriteFile(testFile.Name(), []byte(`
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

	ioutil.WriteFile(testFile.Name(), []byte(`
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
