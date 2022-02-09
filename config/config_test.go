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
	testFile, _ := ioutil.TempFile(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	ioutil.WriteFile(testFile.Name(), []byte{255, 253}, 0644)
	_, err := LoadConfiguration(testFile.Name())
	assert.NotNil(err)
	assert.Contains(fmt.Sprintf("%s", err), "could not parse configuration")
}

func TestLoadConfigurationMissingTimeout(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := ioutil.TempFile(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	ioutil.WriteFile(testFile.Name(), []byte("url: some_url"), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal("some_url", config.URL)
	assert.Equal(DefaultTimeout, config.GetTimeout())
}

func TestLoadConfigurationFoundTimeout(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := ioutil.TempFile(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	ioutil.WriteFile(testFile.Name(), []byte("url: some_url\ntimeout: 100"), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal("some_url", config.URL)
	assert.Equal(100, config.GetTimeout())
}

func TestLoadConfigurationInvalidTimeout(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := ioutil.TempFile(os.TempDir(), "test_file_prefix")
	defer testFile.Close()

	ioutil.WriteFile(testFile.Name(), []byte("url: some_url\ntimeout: abc"), 0644)
	config, _ := LoadConfiguration(testFile.Name())
	assert.Equal("some_url", config.URL)
	assert.Equal(DefaultTimeout, config.GetTimeout())
}
