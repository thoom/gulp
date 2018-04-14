package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func resetRedirectFlags() {
	*followRedirectFlag = false
	*disableRedirectFlag = false
}

func TestShouldFollowRedirects(t *testing.T) {
	assert := assert.New(t)
	resetRedirectFlags()

	*followRedirectFlag = true
	assert.True(shouldFollowRedirects())
}

func TestShouldFollowRedirectsDisabled(t *testing.T) {
	assert := assert.New(t)
	resetRedirectFlags()

	*disableRedirectFlag = true
	assert.False(shouldFollowRedirects())
}
func TestShouldFollowRedirectsConfig(t *testing.T) {
	assert := assert.New(t)
	resetRedirectFlags()

	assert.True(shouldFollowRedirects())
}

func TestShouldFollowRedirectsConfigDisabled(t *testing.T) {
	assert := assert.New(t)
	resetRedirectFlags()

	gulpConfig.Flags["follow_redirects"] = "false"
	assert.False(shouldFollowRedirects())
}

func TestShouldFollowRedirectsConfigDisabledFlagEnable(t *testing.T) {
	assert := assert.New(t)
	resetRedirectFlags()

	*followRedirectFlag = true
	gulpConfig.Flags["follow_redirects"] = "false"
	assert.True(shouldFollowRedirects())
}

func TestShouldFollowRedirectsFlagsMultipleFollow(t *testing.T) {
	assert := assert.New(t)

	*followRedirectFlag = true
	*disableRedirectFlag = true

	os.Args = []string{"cmd", "-no-redirect", "-follow-redirect"}
	assert.True(shouldFollowRedirects())
}

func TestShouldFollowRedirectsFlagsMultipleDisabled(t *testing.T) {
	assert := assert.New(t)

	*followRedirectFlag = true
	*disableRedirectFlag = true

	os.Args = []string{"cmd", "-follow-redirect", "-no-redirect"}
	assert.False(shouldFollowRedirects())
}

func resetDisplayFlags() {
	*responseOnlyFlag = false
	*statusCodeOnlyFlag = false
	*verboseFlag = false
}

func TestFilterDisplayFlagsResponseOnly(t *testing.T) {
	assert := assert.New(t)
	resetDisplayFlags()

	*responseOnlyFlag = true
	filterDisplayFlags()
	assert.True(*responseOnlyFlag)
	assert.False(*statusCodeOnlyFlag)
	assert.False(*verboseFlag)
}

func TestFilterDisplayFlagsStatusCode(t *testing.T) {
	assert := assert.New(t)
	resetDisplayFlags()

	*statusCodeOnlyFlag = true
	filterDisplayFlags()
	assert.False(*responseOnlyFlag)
	assert.True(*statusCodeOnlyFlag)
	assert.False(*verboseFlag)
}

func TestFilterDisplayFlagsVerbose(t *testing.T) {
	assert := assert.New(t)
	resetDisplayFlags()

	*verboseFlag = true
	filterDisplayFlags()
	assert.False(*responseOnlyFlag)
	assert.False(*statusCodeOnlyFlag)
	assert.True(*verboseFlag)
}

func TestFilterDisplayFlagsConfig(t *testing.T) {
	assert := assert.New(t)
	resetDisplayFlags()

	filterDisplayFlags()
	assert.True(*responseOnlyFlag)
	assert.False(*statusCodeOnlyFlag)
	assert.False(*verboseFlag)
}

func TestFilterDisplayFlagsConfigStatusCode(t *testing.T) {
	assert := assert.New(t)
	resetDisplayFlags()

	gulpConfig.Display = "status-code-only"
	filterDisplayFlags()
	assert.False(*responseOnlyFlag)
	assert.True(*statusCodeOnlyFlag)
	assert.False(*verboseFlag)
}

func TestFilterDisplayFlagsConfigVerbose(t *testing.T) {
	assert := assert.New(t)
	resetDisplayFlags()

	gulpConfig.Display = "verbose"
	filterDisplayFlags()
	assert.False(*responseOnlyFlag)
	assert.False(*statusCodeOnlyFlag)
	assert.True(*verboseFlag)
}

func TestFilterDisplayFlagsMultipleResponseOnly(t *testing.T) {
	assert := assert.New(t)

	*responseOnlyFlag = true
	*statusCodeOnlyFlag = true
	*verboseFlag = true

	os.Args = []string{"cmd", "-sco", "-v", "-ro"}

	filterDisplayFlags()
	assert.True(*responseOnlyFlag)
	assert.False(*statusCodeOnlyFlag)
	assert.False(*verboseFlag)
}

func TestFilterDisplayFlagsMultipleStatusCode(t *testing.T) {
	assert := assert.New(t)

	*responseOnlyFlag = true
	*statusCodeOnlyFlag = true
	*verboseFlag = true

	os.Args = []string{"cmd", "-v", "-ro", "-sco"}

	filterDisplayFlags()
	assert.False(*responseOnlyFlag)
	assert.True(*statusCodeOnlyFlag)
	assert.False(*verboseFlag)
}

func TestFilterDisplayFlagsMultipleVerbose(t *testing.T) {
	assert := assert.New(t)

	*responseOnlyFlag = true
	*statusCodeOnlyFlag = true
	*verboseFlag = true

	os.Args = []string{"cmd", "-sco", "-ro", "-v"}

	filterDisplayFlags()
	assert.False(*responseOnlyFlag)
	assert.False(*statusCodeOnlyFlag)
	assert.True(*verboseFlag)
}
