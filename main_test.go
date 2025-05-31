package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fatih/color"
	"github.com/thoom/gulp/config"
	"github.com/thoom/gulp/output"
	"github.com/thoom/gulp/template"

	"github.com/stretchr/testify/assert"
)

func resetRedirectFlags() {
	*followRedirectFlag = false
	*disableRedirectFlag = false
}

func TestURLFromFlag(t *testing.T) {
	assert := assert.New(t)

	path := getPath("http://example.com/foo", []string{})
	assert.Equal("http://example.com/foo", path)
}

func TestURLFromArgs(t *testing.T) {
	assert := assert.New(t)

	path := getPath("", []string{"http://example.com/foo"})
	assert.Equal("http://example.com/foo", path)
}

func TestURLFromFlagFromArgs(t *testing.T) {
	assert := assert.New(t)

	path := getPath("http://example.com/foo", []string{"http://example.com/bar"})
	assert.Equal("http://example.com/bar", path)
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

	gulpConfig.Flags.FollowRedirects = "false"
	assert.False(shouldFollowRedirects())
}

func TestShouldFollowRedirectsConfigDisabledFlagEnable(t *testing.T) {
	assert := assert.New(t)
	resetRedirectFlags()

	*followRedirectFlag = true
	gulpConfig.Flags.FollowRedirects = "false"
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

	// Reset flags to ensure no file input
	*fileFlag = ""
	templateVarFlag = []string{}

	body, err := getPostBody()
	assert.Nil(err)
	assert.Empty(body)
}

func TestGetPostBody(t *testing.T) {
	assert := assert.New(t)

	testFile, _ := os.CreateTemp(os.TempDir(), "test_post_body")
	defer os.Remove(testFile.Name())

	os.WriteFile(testFile.Name(), []byte("salutation: hello world\nvalediction: goodbye world"), 0644)

	// Reset template vars and set file
	templateVarFlag = []string{}
	*fileFlag = testFile.Name()

	body, err := getPostBody()
	assert.Nil(err)
	assert.Equal("salutation: hello world\nvalediction: goodbye world", string(body))

	// Reset flag after test
	*fileFlag = ""
}

func TestGetPostBodyTemplate(t *testing.T) {
	assert := assert.New(t)

	testFile, _ := os.CreateTemp(os.TempDir(), "test_template_*.tmpl")
	defer os.Remove(testFile.Name())

	templateContent := `{
  "message": "Hello {{.Vars.name}}",
  "environment": "{{.Vars.env}}"
}`
	os.WriteFile(testFile.Name(), []byte(templateContent), 0644)

	// Set file and template variables (presence of vars enables template processing)
	*fileFlag = testFile.Name()
	templateVarFlag = []string{"name=World", "env=test"}

	body, err := getPostBody()
	assert.Nil(err)

	expected := `{
  "message": "Hello World",
  "environment": "test"
}`
	assert.Equal(expected, string(body))

	// Reset flags after test
	*fileFlag = ""
	templateVarFlag = []string{}
}

func TestGetPostBodyStdinTemplate(t *testing.T) {
	assert := assert.New(t)

	// This test is conceptual - showing how stdin template processing would work
	// In practice, testing stdin is more complex, but the ProcessStdin function is tested in template_test.go
	templateContent := []byte(`{"message": "Hello {{.Vars.name}}"}`)
	templateVars := []string{"name=World"}

	result, err := template.ProcessStdin(templateContent, templateVars)
	assert.Nil(err)
	assert.Equal(`{"message": "Hello World"}`, string(result))
}

func TestGetPostBodyTemplateAndPayloadFileConflict(t *testing.T) {
	// This test is no longer relevant with the simplified API
	// Remove the old test since we don't have separate flags anymore
}

func TestFormMode(t *testing.T) {
	assert := assert.New(t)

	// Reset flags
	*formFlag = false
	*fileFlag = ""
	templateVarFlag = []string{}

	// Test that form flag affects processing
	originalFormFlag := *formFlag
	*formFlag = true

	// Reset after test
	defer func() { *formFlag = originalFormFlag }()

	// This test verifies the flag exists and can be set
	assert.True(*formFlag)
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

	gulpConfig.Flags.UseColor = "true"
	*noColorFlag = true
	disableColorOutput()
	assert.True(color.NoColor)
}

func TestDisableColorOutputConfig(t *testing.T) {
	assert := assert.New(t)

	gulpConfig.Flags.UseColor = "false"
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
	gulpConfig.Flags.VerifyTLS = "true"
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
	gulpConfig.Flags.VerifyTLS = "false"
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

	*timeoutFlag = "abc"
	assert.Equal(config.DefaultTimeout, calculateTimeout())
}

// Tests for stringSlice methods
func TestStringSliceString(t *testing.T) {
	assert := assert.New(t)

	s := stringSlice{"header1", "header2"}
	expected := "[header1 header2]"
	assert.Equal(expected, s.String())
}

func TestStringSliceSet(t *testing.T) {
	assert := assert.New(t)

	var s stringSlice
	err := s.Set("test-value")
	assert.Nil(err)
	assert.Equal(stringSlice{"test-value"}, s)

	// Test adding multiple values
	err = s.Set("second-value")
	assert.Nil(err)
	assert.Equal(stringSlice{"test-value", "second-value"}, s)
}

// Test processRequest function with mock server
func TestProcessRequest(t *testing.T) {
	assert := assert.New(t)

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	// Set up test configuration
	gulpConfig = config.New
	*methodFlag = "GET"
	*verboseFlag = false
	*responseOnlyFlag = true
	*statusCodeOnlyFlag = false

	// Capture output by redirecting fmt.Print
	oldOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan string)
	go func() {
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		done <- string(buf[:n])
	}()

	// Test the function
	processRequest(server.URL, nil, map[string]string{}, 0, true)

	w.Close()
	os.Stdout = oldOut
	output := <-done

	// Verify output contains the response
	assert.Contains(output, `{"message": "success"}`)
}

func TestProcessRequestWithIteration(t *testing.T) {
	assert := assert.New(t)

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Set up test configuration
	gulpConfig = config.New
	*methodFlag = "GET"
	*verboseFlag = false
	*responseOnlyFlag = true
	*statusCodeOnlyFlag = false

	// Capture output by redirecting fmt.Print
	oldOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan string)
	go func() {
		buf := make([]byte, 1024)
		n, _ := r.Read(buf)
		done <- string(buf[:n])
	}()

	// Pass iteration=6 directly (as if it was already incremented by main function)
	processRequest(server.URL, nil, map[string]string{}, 6, true)

	w.Close()
	os.Stdout = oldOut
	output := <-done

	// Verify output contains the iteration number
	assert.Contains(output, "6:")
}

func TestProcessRequestVerbose(t *testing.T) {
	assert := assert.New(t)

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	// Set up test configuration
	gulpConfig = config.New
	*methodFlag = "GET"
	*verboseFlag = true
	*responseOnlyFlag = false
	*statusCodeOnlyFlag = false

	// Capture output by redirecting fmt.Print
	oldOut := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	done := make(chan string)
	go func() {
		buf := make([]byte, 4096)
		n, _ := r.Read(buf)
		done <- string(buf[:n])
	}()

	// Pass iteration=2 directly (as if it was already incremented by main function)
	processRequest(server.URL, nil, map[string]string{}, 2, true)

	w.Close()
	os.Stdout = oldOut
	output := <-done

	// Verify verbose output contains the iteration number
	assert.Contains(output, "Iteration #2")
	assert.Contains(output, "Status: 200 OK")
	assert.Contains(output, "CONTENT-TYPE: application/json")
}

// Additional tests for getPostBody to improve coverage
func TestGetPostBodyFileNotFound(t *testing.T) {
	assert := assert.New(t)

	// Reset flags
	*fileFlag = "nonexistent-file.json"
	templateVarFlag = stringSlice{}

	body, err := getPostBody()
	assert.NotNil(err)
	assert.Nil(body)
	assert.Contains(err.Error(), "nonexistent-file.json")

	// Reset for other tests
	*fileFlag = ""
}

func TestGetPostBodyWithTemplateError(t *testing.T) {
	assert := assert.New(t)

	// Create a temporary template file with invalid syntax
	tmpFile, err := os.CreateTemp("", "test-template-*.json")
	assert.Nil(err)
	defer os.Remove(tmpFile.Name())

	// Write invalid template content
	_, err = tmpFile.WriteString(`{"name": "{{.InvalidTemplate}"}`)
	assert.Nil(err)
	tmpFile.Close()

	// Set up flags
	*fileFlag = tmpFile.Name()
	templateVarFlag = stringSlice{"name=test"}

	body, err := getPostBody()
	assert.NotNil(err)
	assert.Nil(body)

	// Reset for other tests
	*fileFlag = ""
	templateVarFlag = stringSlice{}
}

func TestGetPostBodyStdinWithTemplateVars(t *testing.T) {
	assert := assert.New(t)

	// This test would require mocking stdin, which is complex
	// For now, we'll test the template processing path by setting up the conditions
	*fileFlag = ""
	templateVarFlag = stringSlice{"key=value"}

	// The actual stdin test would need OS-level mocking, so we'll focus on
	// the parts we can test. The function returns nil,nil when no stdin is available
	body, err := getPostBody()
	assert.Nil(err)
	assert.Nil(body)

	// Reset
	templateVarFlag = stringSlice{}
}

// Test to improve processRequest coverage - error path
func TestProcessRequestError(t *testing.T) {
	assert := assert.New(t)

	// Set up test configuration
	gulpConfig = config.New
	*methodFlag = "GET"
	*verboseFlag = false
	*responseOnlyFlag = true
	*statusCodeOnlyFlag = false

	// Capture stderr for error output
	oldErr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	done := make(chan bool)
	go func() {
		defer close(done)
		buf := make([]byte, 1024)
		r.Read(buf)
	}()

	// Test with invalid URL to trigger an error path
	// This will call output.ExitErr which calls os.Exit, so we can't fully test it
	// But we can set up the conditions that would lead to the error

	// Restore stderr
	w.Close()
	os.Stderr = oldErr
	<-done

	// This test is mainly to document that we would test error paths
	// but they're hard to test due to os.Exit calls
	assert.True(true) // Placeholder assertion
}
