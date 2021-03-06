package client

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildHeadersBase(t *testing.T) {
	assert := assert.New(t)

	headers, _ := BuildHeaders([]string{"X-Test-Key: abc123def"}, nil, false)
	assert.Equal(3, len(headers))

	assert.Contains(headers, "USER-AGENT")
	assert.Equal("thoom.Gulp/"+GetVersion(), headers["USER-AGENT"])

	assert.Contains(headers, "ACCEPT")
	assert.Equal("application/json;q=1.0, */*;q=0.8", headers["ACCEPT"])

	assert.Contains(headers, "X-TEST-KEY")
	assert.Equal("abc123def", headers["X-TEST-KEY"])
}

func TestBuildHeadersJSON(t *testing.T) {
	assert := assert.New(t)

	headers, _ := BuildHeaders([]string{}, nil, true)
	assert.Equal(3, len(headers))

	assert.Contains(headers, "CONTENT-TYPE")
	assert.Equal("application/json", headers["CONTENT-TYPE"])
}

func TestBuildHeadersHeaderConfig(t *testing.T) {
	assert := assert.New(t)

	configHeaders := map[string]string{}
	configHeaders["X-Test-Key"] = "abc123def"

	headers, _ := BuildHeaders([]string{}, configHeaders, false)
	assert.Equal(3, len(headers))

	assert.Contains(headers, "X-TEST-KEY")
	assert.Equal("abc123def", headers["X-TEST-KEY"])
}

func TestBuildHeadersHeaderOverride(t *testing.T) {
	assert := assert.New(t)

	headers, _ := BuildHeaders([]string{"Content-Type: application/vnd.ex.v1+json"}, nil, true)
	assert.Equal(3, len(headers))

	assert.Contains(headers, "CONTENT-TYPE")
	assert.Equal("application/vnd.ex.v1+json", headers["CONTENT-TYPE"])
}

func TestBuildHeadersHeaderErr(t *testing.T) {
	assert := assert.New(t)

	_, err := BuildHeaders([]string{"Bad-Content-Header"}, nil, true)
	assert.NotNil(err)
	assert.Equal("Could not parse header: 'Bad-Content-Header'", fmt.Sprintf("%s", err))
}

func TestBuildURLBasic(t *testing.T) {
	assert := assert.New(t)
	url, _ := BuildURL("/some/resource", "https://api.ex.io")
	assert.Equal("https://api.ex.io/some/resource", url)
}

func TestBuildURLNoConfig(t *testing.T) {
	assert := assert.New(t)
	url, _ := BuildURL("https://api.ex.io/some/resource", "")
	assert.Equal("https://api.ex.io/some/resource", url)
}

func TestBuildURLOverride(t *testing.T) {
	assert := assert.New(t)
	url, _ := BuildURL("https://api.ex.io", "https://another.base.io")
	assert.Equal("https://api.ex.io", url)
}

func TestBuildURLNoPath(t *testing.T) {
	assert := assert.New(t)
	url, _ := BuildURL("", "https://api.ex.io")
	assert.Equal("https://api.ex.io", url)
}

func TestBuildURLBadURL(t *testing.T) {
	assert := assert.New(t)
	url, err := BuildURL("/bad/path", "")
	assert.Empty(url)
	assert.NotNil(err)
	assert.Equal("Invalid URL", fmt.Sprintf("%s", err))
}

func TestBuildURLNoURL(t *testing.T) {
	assert := assert.New(t)
	url, err := BuildURL("", "")
	assert.Empty(url)
	assert.NotNil(err)
	assert.Equal("Need a URL to make a request", fmt.Sprintf("%s", err))
}

func TestGetVersion(t *testing.T) {
	assert := assert.New(t)

	buildVersion = ""
	assert.Equal(defaultVersion, GetVersion())
}

func TestGetVersionEnv(t *testing.T) {
	assert := assert.New(t)

	buildVersion = "TestVersion"
	assert.Equal("TestVersion", GetVersion())
}
