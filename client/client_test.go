package client

import (
	"net/http"
	"net/http/httputil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDisableTLS(t *testing.T) {
	assert := assert.New(t)

	DisableTLSVerification()
	assert.True(http.DefaultTransport.(*http.Transport).TLSClientConfig.InsecureSkipVerify)
}

func TestCreateRequest(t *testing.T) {
	assert := assert.New(t)

	method := "GET"
	url := "http://test.ex.io"
	headers := map[string]string{}
	headers["X-Test-Header"] = "abc123def"

	req, err := CreateRequest(method, url, "", headers)
	assert.Nil(err)
	assert.Equal(url, req.URL.String())
	assert.Equal(method, req.Method)
	assert.Equal(1, len(req.Header))
	assert.EqualValues(headers["X-Test-Header"], req.Header.Get("X-Test-Header"))
	assert.Nil(req.Body)
}

func TestCreateRequestGetWithBody(t *testing.T) {
	assert := assert.New(t)

	method := "GET"
	url := "http://test.ex.io"
	body := "body!"

	req, err := CreateRequest(method, url, body, map[string]string{})
	assert.Nil(err)
	assert.Equal(url, req.URL.String())
	assert.Equal(method, req.Method)
	assert.Empty(req.Header)
	assert.Nil(req.Body)
}

func TestCreateRequestPostWithBody(t *testing.T) {
	assert := assert.New(t)

	method := "POST"
	url := "http://test.ex.io"
	body := "body!"

	req, err := CreateRequest(method, url, body, map[string]string{})
	assert.Nil(err)
	assert.Equal(url, req.URL.String())
	assert.Equal(method, req.Method)
	assert.Empty(req.Header)

	// Hacky way to get the body for now
	requestDump, _ := httputil.DumpRequest(req, true)
	reqDumpStr := strings.Split(string(requestDump), "\n")
	assert.Equal(body, reqDumpStr[len(reqDumpStr)-1])
}

func TestCreateClient(t *testing.T) {
	assert := assert.New(t)

	client := CreateClient(false, 10)
	assert.Equal(time.Duration(10)*time.Second, client.Timeout)
}

func TestCreateClientFollowRedirects(t *testing.T) {
	assert := assert.New(t)
	client := CreateClient(true, 10)
	assert.Equal(time.Duration(10)*time.Second, client.Timeout)
}
