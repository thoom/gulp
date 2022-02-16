package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/thoom/gulp/config"
	"github.com/thoom/gulp/output"

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

	gulpConfig.Flags.FollowRedirects = false
	assert.False(shouldFollowRedirects())
}

func TestShouldFollowRedirectsConfigDisabledFlagEnable(t *testing.T) {
	assert := assert.New(t)
	resetRedirectFlags()

	*followRedirectFlag = true
	gulpConfig.Flags.FollowRedirects = false
	assert.True(shouldFollowRedirects())
}

func TestShouldFollowRedirectsFlagsMultipleFollow(t *testing.T) {
	assert := assert.New(t)

	*followRedirectFlag = true
	*disableRedirectFlag = true

	os.Args = []string{"cmd", "-no-redirect", "-follow-redirect", "/foo/path"}
	assert.True(shouldFollowRedirects())
}

func TestShouldFollowRedirectsFlagsMultipleDisabled(t *testing.T) {
	assert := assert.New(t)

	*followRedirectFlag = true
	*disableRedirectFlag = true

	os.Args = []string{"cmd", "-follow-redirect", "-no-redirect", "/foo/path"}
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

	os.Args = []string{"cmd", "-sco", "-v", "-ro", "-no-redirect"}

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

	os.Args = []string{"cmd", "-v", "-ro", "-sco", "-no-redirect"}

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

	os.Args = []string{"cmd", "-sco", "-ro", "-v", "-no-redirect"}

	filterDisplayFlags()
	assert.False(*responseOnlyFlag)
	assert.False(*statusCodeOnlyFlag)
	assert.True(*verboseFlag)
}

func TestHandleResponse(t *testing.T) {
	resetDisplayFlags()
	assert := assert.New(t)
	*verboseFlag = true

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte("{\"salutation\":\"hello world\"}"))
	}

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	// Disable color output for now
	output.NoColor(true)

	b := &bytes.Buffer{}
	bo := &output.BuffOut{Out: b, Err: b}
	handleResponse(w.Result(), 10, bo)
	assert.Equal(200, w.Result().StatusCode)
	assert.Equal("Status: 200 OK (10.00 seconds)\n\nCONTENT-TYPE: application/json\n\n{\n  \"salutation\": \"hello world\"\n}\n", b.String())
}

func TestHandleResponseStatusCode(t *testing.T) {
	resetDisplayFlags()

	assert := assert.New(t)
	*statusCodeOnlyFlag = true

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}

	req := httptest.NewRequest("GET", "http://api.ex.io/foo", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	b := &bytes.Buffer{}
	bo := &output.BuffOut{Out: b, Err: b}
	handleResponse(w.Result(), 10, bo)
	assert.Equal(200, w.Result().StatusCode)
	assert.Equal("200\n", b.String())
}

func TestGetPostBodyEmpty(t *testing.T) {
	assert := assert.New(t)

	body, err := getPostBody(os.Stdin)
	assert.Nil(err)
	assert.Empty(body)
}

func TestGetPostBody(t *testing.T) {
	assert := assert.New(t)

	testFile, _ := os.CreateTemp(os.TempDir(), "test_post_body")
	ioutil.WriteFile(testFile.Name(), []byte("salutation: hello world\nvalediction: goodbye world"), 0644)

	f, _ := os.Open(testFile.Name())
	defer f.Close()

	body, err := getPostBody(f)
	assert.Nil(err)
	assert.Equal("salutation: hello world\nvalediction: goodbye world", string(body))
}

func TestConvertJSONBody(t *testing.T) {
	assert := assert.New(t)

	yaml := `
salutation: hello world
valediction: goodbye world
`
	body, err := convertJSONBody([]byte(yaml), map[string]string{"CONTENT-TYPE": "application/json"})
	assert.Nil(err)
	assert.Equal("{\"salutation\":\"hello world\",\"valediction\":\"goodbye world\"}", string(body))
}

func TestConvertJSONBodyNotJSON(t *testing.T) {
	assert := assert.New(t)

	body, err := convertJSONBody([]byte("Not JSON, but plain text"), map[string]string{"CONTENT-TYPE": "text/plain"})
	assert.Nil(err)
	assert.Equal("Not JSON, but plain text", string(body))
}

func TestConvertJSONBodyInvalidJson(t *testing.T) {
	assert := assert.New(t)

	body, err := convertJSONBody([]byte{255, 253}, map[string]string{"CONTENT-TYPE": "application/json"})
	assert.Nil(body)
	assert.Contains(fmt.Sprintf("%s", err), "could not parse post body: yaml:")
}

func TestDisableColorOutput(t *testing.T) {
	assert := assert.New(t)

	gulpConfig.Flags.UseColor = true
	*noColorFlag = true
	disableColorOutput()
	assert.True(color.NoColor)
}

func TestDisableColorOutputConfig(t *testing.T) {
	assert := assert.New(t)

	gulpConfig.Flags.UseColor = false
	*noColorFlag = false
	disableColorOutput()
	assert.True(color.NoColor)
}

func TestDisableTLSVerify(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	bo := output.BuffOut{Out: b, Err: b}
	output.Out = &bo

	*insecureFlag = true
	*verboseFlag = true
	gulpConfig.Flags.VerifyTLS = true
	disableTLSVerify()

	assert.Equal("WARNING: TLS CHECKING IS DISABLED FOR THIS REQUEST\n", b.String())
}

func TestDisableTLSVerifyConfig(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	bo := output.BuffOut{Out: b, Err: b}
	output.Out = &bo

	*insecureFlag = false
	*verboseFlag = true
	gulpConfig.Flags.VerifyTLS = false
	disableTLSVerify()

	assert.Equal("WARNING: TLS CHECKING IS DISABLED FOR THIS REQUEST\n", b.String())
}

func TestPrintRequestNotVerboseRepeat1(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	bo := &output.BuffOut{Out: b, Err: b}

	*verboseFlag = false
	printRequest(0, "http://test.fake", map[string][]string{}, 0, "", bo)
	assert.Equal("", b.String())
}

func TestPrintRequestNotVerboseRepeat7(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	bo := &output.BuffOut{Out: b, Err: b}

	*verboseFlag = false
	printRequest(7, "http://test.fake", map[string][]string{}, 0, "", bo)
	assert.Equal("7: ", b.String())
}

func TestPrintRequestVerboseRepeat0(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	bo := &output.BuffOut{Out: b, Err: b}

	*verboseFlag = true
	printRequest(0, "http://test.fake", map[string][]string{}, 0, "", bo)
	assert.Equal("\nGET http://test.fake\n\n", b.String())
}

func TestPrintRequestVerboseRepeat7(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	bo := &output.BuffOut{Out: b, Err: b}

	*verboseFlag = true
	printRequest(7, "http://test.fake", map[string][]string{}, 0, "", bo)
	assert.Equal("\nIteration #7\n\n\nGET http://test.fake\n\n", b.String())
}

func TestPrintRequestVerboseRepeat0Headers(t *testing.T) {
	assert := assert.New(t)

	b := &bytes.Buffer{}
	bo := &output.BuffOut{Out: b, Err: b}
	*verboseFlag = true

	headers := map[string][]string{}
	headers["X-TEST"] = []string{"abc123def"}

	printRequest(0, "http://test.fake", headers, 9, "HTTP 1.1", bo)
	assert.Equal("\nGET http://test.fake  \n\nPROTOCOL: HTTP 1.1    \nACCEPT-ENCODING: gzip \nCONTENT-LENGTH: 9     \nX-TEST: abc123def     \n\n", b.String())
}

func TestCalculateTimeout(t *testing.T) {
	assert := assert.New(t)

	*timeoutFlag = ""
	gulpConfig = config.New
	assert.Equal(config.DefaultTimeout, calculateTimeout())
}

func TestCalculateTimeoutFlag(t *testing.T) {
	assert := assert.New(t)

	gulpConfig = config.New
	*timeoutFlag = "100"
	assert.Equal(100, calculateTimeout())
}

func TestCalculateTimeoutFlagInvalid(t *testing.T) {
	assert := assert.New(t)

	gulpConfig = config.New
	*timeoutFlag = "abc123"
	assert.Equal(config.DefaultTimeout, calculateTimeout())
}
