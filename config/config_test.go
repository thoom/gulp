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

	assert.True(New.UseColor())
	assert.True(New.VerifyTLS())
}

func TestLoadConfigurationMissing(t *testing.T) {
	assert := assert.New(t)

	_, err := LoadConfiguration("invalidFile.yml")
	assert.NotNil(err)
	assert.Contains(fmt.Sprintf("%s", err), "Could not load configuration")
}

func TestLoadConfigurationNoParse(t *testing.T) {
	assert := assert.New(t)
	testFile, _ := ioutil.TempFile(os.TempDir(), "test_file_prefix")

	ioutil.WriteFile(testFile.Name(), []byte{255, 253}, 0644)
	_, err := LoadConfiguration(testFile.Name())
	assert.NotNil(err)
	assert.Contains(fmt.Sprintf("%s", err), "Could not parse configuration")
}
