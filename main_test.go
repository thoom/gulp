package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildHeadersBase(t *testing.T) {
	assert := assert.New(t)

	gulpConfig.Headers = map[string]string{}
	headers, _ := buildHeaders([]string{"X-Test-Key: abc123def"}, false)
	assert.Equal(3, len(headers))

	assert.Contains(headers, "USER-AGENT")
	assert.Equal("thoom.Gulp/"+VERSION, headers["USER-AGENT"])

	assert.Contains(headers, "ACCEPT")
	assert.Equal("application/json;q=1.0, */*;q=0.8", headers["ACCEPT"])

	assert.Contains(headers, "X-TEST-KEY")
	assert.Equal("abc123def", headers["X-TEST-KEY"])
}

func TestBuildHeadersJSON(t *testing.T) {
	assert := assert.New(t)

	headers, _ := buildHeaders([]string{}, true)
	assert.Equal(3, len(headers))

	assert.Contains(headers, "CONTENT-TYPE")
	assert.Equal("application/json", headers["CONTENT-TYPE"])
}

func TestBuildHeadersHeaderConfig(t *testing.T) {
	assert := assert.New(t)

	gulpConfig.Headers = map[string]string{}
	gulpConfig.Headers["X-Test-Key"] = "abc123def"

	headers, _ := buildHeaders([]string{}, false)
	assert.Equal(3, len(headers))

	assert.Contains(headers, "X-TEST-KEY")
	assert.Equal("abc123def", headers["X-TEST-KEY"])
}

func TestBuildHeadersHeaderOverride(t *testing.T) {
	assert := assert.New(t)

	gulpConfig.Headers = map[string]string{}
	headers, _ := buildHeaders([]string{"Content-Type: application/vnd.ex.v1+json"}, true)
	assert.Equal(3, len(headers))

	assert.Contains(headers, "CONTENT-TYPE")
	assert.Equal("application/vnd.ex.v1+json", headers["CONTENT-TYPE"])
}

func TestBuildHeadersHeaderErr(t *testing.T) {
	assert := assert.New(t)

	_, err := buildHeaders([]string{"Bad-Content-Header"}, true)
	assert.NotNil(err)
	assert.Equal("Could not parse header: 'Bad-Content-Header'", fmt.Sprintf("%s", err))
}

func reset() {
	*responseOnlyFlag = false
	*statusCodeOnlyFlag = false
	*verboseFlag = false
}

func TestFilterDisplayFlagsResponseOnly(t *testing.T) {
	assert := assert.New(t)
	reset()

	*responseOnlyFlag = true
	filterDisplayFlags()
	assert.True(*responseOnlyFlag)
	assert.False(*statusCodeOnlyFlag)
	assert.False(*verboseFlag)
}

func TestFilterDisplayFlagsStatusCode(t *testing.T) {
	assert := assert.New(t)
	reset()

	*statusCodeOnlyFlag = true
	filterDisplayFlags()
	assert.False(*responseOnlyFlag)
	assert.True(*statusCodeOnlyFlag)
	assert.False(*verboseFlag)
}

func TestFilterDisplayFlagsVerbose(t *testing.T) {
	assert := assert.New(t)
	reset()

	*verboseFlag = true
	filterDisplayFlags()
	assert.False(*responseOnlyFlag)
	assert.False(*statusCodeOnlyFlag)
	assert.True(*verboseFlag)
}

func TestFilterDisplayFlagsConfig(t *testing.T) {
	assert := assert.New(t)
	reset()

	filterDisplayFlags()
	assert.True(*responseOnlyFlag)
	assert.False(*statusCodeOnlyFlag)
	assert.False(*verboseFlag)
}

func TestFilterDisplayFlagsConfigStatusCode(t *testing.T) {
	assert := assert.New(t)
	reset()

	gulpConfig.Display = "status-code-only"
	filterDisplayFlags()
	assert.False(*responseOnlyFlag)
	assert.True(*statusCodeOnlyFlag)
	assert.False(*verboseFlag)
}

func TestFilterDisplayFlagsConfigVerbose(t *testing.T) {
	assert := assert.New(t)
	reset()

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

	os.Args = []string{"cmd", "-sco", "-I", "-ro"}

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

	os.Args = []string{"cmd", "-I", "-ro", "-sco"}

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

	os.Args = []string{"cmd", "-sco", "-ro", "-I"}

	filterDisplayFlags()
	assert.False(*responseOnlyFlag)
	assert.False(*statusCodeOnlyFlag)
	assert.True(*verboseFlag)
}
