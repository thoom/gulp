package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ghodss/yaml"
	"github.com/thoom/gulp/client"
	"github.com/thoom/gulp/config"
	"github.com/thoom/gulp/form"
	"github.com/thoom/gulp/output"
	"github.com/thoom/gulp/template"
)

type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%s", *s)
}

// The second method is Set(value string) error
func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

var (
	reqHeaders stringSlice

	gulpConfig          = config.New
	methodFlag          = flag.String("m", "GET", "The `method` to use: ie. HEAD, GET, POST, PUT, DELETE")
	configFlag          = flag.String("c", ".gulp.yml", "The `configuration` file to use")
	clientCert          = flag.String("client-cert", "", "If using client cert auth, the cert to use. MUST be paired with -client-cert-key flag")
	clientCertKey       = flag.String("client-cert-key", "", "If using client cert auth, the key to use. MUST be paired with -client-cert flag")
	clientCA            = flag.String("custom-ca", "", "If using a custom CA certificate, the CA cert file to use for verification")
	basicAuthUser       = flag.String("basic-auth-user", "", "Username for basic authentication")
	basicAuthPass       = flag.String("basic-auth-pass", "", "Password for basic authentication")
	insecureFlag        = flag.Bool("insecure", false, "Disable TLS certificate checking")
	responseOnlyFlag    = flag.Bool("ro", false, "Only display the response body (default)")
	statusCodeOnlyFlag  = flag.Bool("sco", false, "Only display the response code")
	verboseFlag         = flag.Bool("v", false, "Display the response body along with various headers")
	timeoutFlag         = flag.String("timeout", "", "The number of `seconds` to wait before the connection times out "+fmt.Sprintf("(default %d)", config.DefaultTimeout))
	noColorFlag         = flag.Bool("no-color", false, "Disables color output for the request")
	followRedirectFlag  = flag.Bool("follow-redirect", false, "Enables following 3XX redirects (default)")
	disableRedirectFlag = flag.Bool("no-redirect", false, "Disables following 3XX redirects")
	repeatFlag          = flag.Int("repeat-times", 1, "Number of `iteration`s to submit the request")
	concurrentFlag      = flag.Int("repeat-concurrent", 1, "Number of concurrent `connections` to use")
	urlFlag             = flag.String("url", "", "The `URL` to use for the request. Alternative to requiring a URL at the end of the command")
	versionFlag         = flag.Bool("version", false, "Display the current client version")

	// New flags for template and payload file support
	fileFlag        = flag.String("file", "", "JSON, YAML, or Go template `file` to use as request body (template processing enabled when -var flags are present)")
	templateVarFlag stringSlice // Will be initialized in main() with flag.Var
	formFlag        = flag.Bool("form", false, "Send data as application/x-www-form-urlencoded instead of JSON")
)

func main() {
	flag.Var(&reqHeaders, "H", "Set a `request` header")
	flag.Var(&templateVarFlag, "var", "Set a template `variable` in the format key=value for use in Go templates")
	flag.Parse()

	if err := runGulp(); err != nil {
		output.ExitErr("", err)
	}
}

// runGulp contains the main application logic, extracted for testability
func runGulp() error {
	// Load the custom configuration
	loadedConfig, err := config.LoadConfiguration(*configFlag)
	if err != nil {
		return err
	}

	// Set the main config to the one that was loaded
	gulpConfig = loadedConfig

	// Disable color output for the request
	disableColorOutput()

	// Make sure that the displayFlags are set appropriately
	filterDisplayFlags()

	if *versionFlag {
		return handleVersionFlag()
	}

	return executeRequest()
}

// handleVersionFlag handles the version flag display and update checking
func handleVersionFlag() error {
	// Check for updates with a 3-second timeout
	currentVersion := client.GetVersion()
	updateInfo, err := client.CheckForUpdates(currentVersion, 3*time.Second)

	if err != nil {
		// If update check fails, just show the version without update info
		output.Out.PrintVersion(currentVersion)
		if *verboseFlag {
			output.Out.PrintWarning(fmt.Sprintf("Could not check for updates: %s", err))
		}
	} else {
		// Show version with update information
		output.Out.PrintVersionWithUpdates(
			currentVersion,
			updateInfo.HasUpdate,
			updateInfo.LatestVersion,
			updateInfo.UpdateURL,
		)
	}
	os.Exit(0)
	return nil // This line won't be reached but makes the function signature consistent
}

// executeRequest handles the HTTP request execution logic
func executeRequest() error {
	url, err := client.BuildURL(getPath(*urlFlag, flag.Args()), gulpConfig.URL)
	if err != nil {
		return err
	}

	// Don't check the TLS bro
	disableTLSVerify()

	// If the disableRedirectFlag is false and follow redirects is false, then set the flag to true
	followRedirect := shouldFollowRedirects()

	var body []byte
	var formContentType string
	// Don't get the post body if it's a GET/HEAD request
	if *methodFlag != "GET" && *methodFlag != "HEAD" {
		var err error
		body, err = getPostBody()
		if err != nil {
			return err
		}

		// Process form data if form flag is set
		if *formFlag && body != nil {
			body, formContentType, err = form.ProcessFormData(body)
			if err != nil {
				return err
			}
		}
	}

	// Build request headers
	headers, err := client.BuildHeaders(reqHeaders, gulpConfig.Headers, body != nil)
	if err != nil {
		return err
	}

	// Set form content type if processing form data
	if *formFlag && formContentType != "" {
		headers["CONTENT-TYPE"] = formContentType
	} else if !*formFlag {
		// Convert the YAML/JSON body if necessary (only when not in form mode)
		body, err = convertJSONBody(body, headers)
		if err != nil {
			return err
		}
	}

	return executeRequestsWithConcurrency(url, body, headers, followRedirect)
}

// executeRequestsWithConcurrency handles the concurrent request execution
func executeRequestsWithConcurrency(url string, body []byte, headers map[string]string, followRedirect bool) error {
	maxChan := make(chan bool, *concurrentFlag)
	var wg sync.WaitGroup
	for i := 0; i < *repeatFlag; i++ {
		wg.Add(1)
		maxChan <- true
		go func(iteration int, maxChan chan bool, wg *sync.WaitGroup) {
			defer wg.Done()
			defer func(maxChan chan bool) { <-maxChan }(maxChan)
			if *repeatFlag > 1 {
				iteration++
			}
			processRequest(url, body, headers, iteration, followRedirect)
		}(i, maxChan, &wg)
	}
	wg.Wait()
	return nil
}

func getPath(urlFlag string, args []string) string {
	path := urlFlag
	if len(args) > 0 {
		path = args[0]
	}

	return path
}

func processRequest(url string, body []byte, headers map[string]string, iteration int, followRedirect bool) {
	if err := executeHTTPRequest(url, body, headers, iteration, followRedirect); err != nil {
		output.ExitErr("", err)
	}
}

// executeHTTPRequest performs the actual HTTP request - extracted for testability
func executeHTTPRequest(url string, body []byte, headers map[string]string, iteration int, followRedirect bool) error {
	var startTimer time.Time

	// Build client auth configuration
	clientAuth := client.BuildClientAuth(*clientCert, *clientCertKey, *clientCA, *basicAuthUser, *basicAuthPass, gulpConfig.ClientAuth)

	req, err := client.CreateRequest(*methodFlag, url, body, headers, clientAuth)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	b := &bytes.Buffer{}
	defer fmt.Print(b)
	bo := &output.BuffOut{Out: b, Err: b}

	startTimer = time.Now()
	reqClient, err := client.CreateClient(followRedirect, calculateTimeout(), clientAuth)
	if err != nil {
		return fmt.Errorf("could not create client: %w", err)
	}

	resp, err := reqClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}

	// If we got a request, output what was created
	printRequest(iteration, url, resp.Request.Header, req.ContentLength, req.Proto, bo)
	handleResponse(resp, time.Since(startTimer).Seconds(), bo)
	return nil
}

func printRequest(iteration int, url string, headers map[string][]string, contentLength int64, protocol string, bo *output.BuffOut) {
	if !*verboseFlag {
		printIterationPrefix(iteration, bo)
		return
	}

	printIterationHeader(iteration, bo)

	if len(headers) == 0 {
		bo.PrintHeader(fmt.Sprintf("%s %s", *methodFlag, url))
		return
	}

	requestInfo := buildRequestInfo(url, protocol, headers, contentLength)
	bo.PrintBlock(requestInfo)
	fmt.Fprintln(bo.Out)
}

// printIterationPrefix prints the iteration number for non-verbose mode
func printIterationPrefix(iteration int, bo *output.BuffOut) {
	if iteration > 0 {
		fmt.Fprintf(bo.Out, "%d: ", iteration)
	}
}

// printIterationHeader prints the iteration header for verbose mode
func printIterationHeader(iteration int, bo *output.BuffOut) {
	if iteration > 0 {
		bo.PrintHeader(fmt.Sprintf("Iteration #%d", iteration))
	}
}

// buildRequestInfo builds the complete request info block with headers
func buildRequestInfo(url, protocol string, headers map[string][]string, contentLength int64) string {
	// Add standard headers that aren't automatically included
	enrichedHeaders := enrichHeaders(headers, contentLength)

	// Build the info block
	block := []string{
		fmt.Sprintf("%s %s", *methodFlag, url),
		"PROTOCOL: " + protocol,
	}

	// Add sorted headers
	sortedHeaders := getSortedHeaders(enrichedHeaders)
	block = append(block, sortedHeaders...)

	return strings.Join(block, "\n")
}

// enrichHeaders adds standard headers that may be missing from the request
func enrichHeaders(headers map[string][]string, contentLength int64) map[string][]string {
	// Create a copy to avoid modifying the original
	enriched := make(map[string][]string)
	for k, v := range headers {
		enriched[k] = v
	}

	// Add headers that aren't automatically captured
	enriched["Content-Length"] = []string{strconv.FormatInt(contentLength, 10)}
	enriched["Accept-Encoding"] = []string{"gzip"}

	return enriched
}

// getSortedHeaders returns header lines sorted alphabetically
func getSortedHeaders(headers map[string][]string) []string {
	headerKeys := make([]string, 0, len(headers))
	for k := range headers {
		headerKeys = append(headerKeys, k)
	}
	sort.Strings(headerKeys)

	var headerLines []string
	for _, k := range headerKeys {
		for _, v := range headers[k] {
			headerLines = append(headerLines, strings.ToUpper(k)+": "+v)
		}
	}

	return headerLines
}

func handleResponse(resp *http.Response, duration float64, bo *output.BuffOut) {
	if *statusCodeOnlyFlag {
		fmt.Fprintln(bo.Out, resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if *verboseFlag {
		bo.PrintStoplight(fmt.Sprintf("Status: %s (%.2f seconds)\n", resp.Status, duration), resp.StatusCode >= 400)
		printResponseHeaders(resp.Header, bo)
		fmt.Fprintln(bo.Out, "")
	}

	formattedBody := formatResponseBody(body, resp.Header)
	fmt.Fprintln(bo.Out, string(formattedBody))
}

// printResponseHeaders prints response headers in sorted order
func printResponseHeaders(headers http.Header, bo *output.BuffOut) {
	headerKeys := make([]string, 0, len(headers))
	for k := range headers {
		headerKeys = append(headerKeys, k)
	}
	sort.Strings(headerKeys)

	for _, k := range headerKeys {
		fmt.Fprintln(bo.Out, strings.ToUpper(k)+": "+headers.Get(k))
	}
}

// formatResponseBody formats the response body, applying JSON pretty-printing if applicable
func formatResponseBody(body []byte, headers http.Header) []byte {
	if !*verboseFlag {
		return body
	}

	// Check if content is JSON
	contentType := headers.Get("Content-Type")
	if !strings.Contains(contentType, "json") {
		return body
	}

	// Try to pretty-print JSON
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, body, "", "  "); err == nil {
		return prettyJSON.Bytes()
	}

	// Return original if pretty-printing failed
	return body
}

func getPostBody() ([]byte, error) {
	// Priority order: file > stdin

	// Handle file input
	if *fileFlag != "" {
		return template.ProcessTemplate(*fileFlag, templateVarFlag)
	}

	// Handle stdin (existing behavior)
	return getPostBodyFromStdin()
}

// getPostBodyFromStdin handles reading from stdin - extracted for better testability
func getPostBodyFromStdin() ([]byte, error) {
	stat, _ := os.Stdin.Stat()

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		return readAndProcessStdin()
	}

	return nil, nil
}

// readAndProcessStdin reads from stdin and optionally processes it as a template
func readAndProcessStdin() ([]byte, error) {
	scanner := bufio.NewScanner(os.Stdin)
	var stdin []byte
	first := true
	for scanner.Scan() {
		if first {
			first = false
		} else {
			stdin = append(stdin, []byte("\n")...)
		}

		stdin = append(stdin, scanner.Bytes()...)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("reading standard input: %s", err)
	}

	// If template variables are provided, process stdin as a template
	if len(templateVarFlag) > 0 {
		return template.ProcessStdin(stdin, templateVarFlag)
	}

	return stdin, nil
}

func convertJSONBody(body []byte, headers map[string]string) ([]byte, error) {
	// Determine if we should convert the body to JSON
	if !strings.Contains(headers["CONTENT-TYPE"], "json") {
		return body, nil
	}

	j, err := yaml.YAMLToJSON(body)
	if err != nil {
		return nil, fmt.Errorf("could not parse post body: %s", err)
	}

	return j, nil
}

func disableColorOutput() {
	if *noColorFlag || !gulpConfig.UseColor() {
		output.NoColor(true)
	}
}

func disableTLSVerify() {
	if *insecureFlag || !gulpConfig.VerifyTLS() {
		if *verboseFlag {
			output.Out.PrintWarning("TLS checking is disabled for this request")
		}
		client.DisableTLSVerification()
	}
}

func calculateTimeout() int {
	if *timeoutFlag == "" {
		return gulpConfig.GetTimeout()
	}

	i, err := strconv.Atoi(*timeoutFlag)
	if err != nil {
		return gulpConfig.GetTimeout()
	}

	return i
}

func shouldFollowRedirects() bool {
	flagCount := countRedirectFlags()

	// No redirect flags set - use config
	if flagCount == 0 {
		return gulpConfig.FollowRedirects()
	}

	// Only one flag set - use it directly
	if flagCount == 1 {
		return *followRedirectFlag
	}

	// Multiple flags set - use the last one specified
	return getLastRedirectFlagFromArgs()
}

// countRedirectFlags returns how many redirect-related flags are set
func countRedirectFlags() int {
	count := 0
	if *disableRedirectFlag {
		count++
	}
	if *followRedirectFlag {
		count++
	}
	return count
}

// getLastRedirectFlagFromArgs parses command line args to find the last redirect flag
func getLastRedirectFlagFromArgs() bool {
	totalArgs := len(os.Args[1:])
	*disableRedirectFlag = false
	*followRedirectFlag = false

	for i := totalArgs; i > 0; i-- {
		switch os.Args[i] {
		case "-no-redirect":
			*disableRedirectFlag = true
			return false
		case "-follow-redirect":
			*followRedirectFlag = true
			return true
		}
	}

	// Fallback (shouldn't reach here if count was correct)
	return *followRedirectFlag
}

type DisplayMode int

const (
	DisplayResponseOnly DisplayMode = iota
	DisplayStatusCode
	DisplayVerbose
)

func filterDisplayFlags() {
	flagCount := countDisplayFlags()

	// No display flags set - use config
	if flagCount == 0 {
		setDisplayModeFromConfig()
		return
	}

	// Only one flag set - already correct
	if flagCount == 1 {
		return
	}

	// Multiple flags set - use the last one specified
	setDisplayModeFromLastArg()
}

// countDisplayFlags returns how many display-related flags are set
func countDisplayFlags() int {
	count := 0
	if *responseOnlyFlag {
		count++
	}
	if *statusCodeOnlyFlag {
		count++
	}
	if *verboseFlag {
		count++
	}
	return count
}

// setDisplayModeFromConfig sets the display mode based on configuration
func setDisplayModeFromConfig() {
	switch gulpConfig.Display {
	case "status-code-only":
		*statusCodeOnlyFlag = true
	case "verbose":
		*verboseFlag = true
	default:
		*responseOnlyFlag = true
	}
}

// setDisplayModeFromLastArg finds the last display flag in args and sets only that one
func setDisplayModeFromLastArg() {
	// Reset all flags first
	*responseOnlyFlag = false
	*statusCodeOnlyFlag = false
	*verboseFlag = false

	totalArgs := len(os.Args[1:])
	for i := totalArgs; i > 0; i-- {
		switch os.Args[i] {
		case "-ro":
			*responseOnlyFlag = true
			return
		case "-sco":
			*statusCodeOnlyFlag = true
			return
		case "-v":
			*verboseFlag = true
			return
		}
	}
}
