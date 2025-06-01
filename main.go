package main

import (
	"bufio"
	"bytes"
	"encoding/json"
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
	"github.com/spf13/cobra"
	"github.com/thoom/gulp/client"
	"github.com/thoom/gulp/config"
	"github.com/thoom/gulp/form"
	"github.com/thoom/gulp/output"
	"github.com/thoom/gulp/template"
)

// stringSlice implements pflag.Value interface for string arrays
type stringSlice []string

func (s *stringSlice) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringSlice) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func (s *stringSlice) Type() string {
	return "stringSlice"
}

// Global variables for configuration and flags
var (
	gulpConfig = config.New

	// Core flags (keep beloved short forms)
	method     string
	verbose    bool
	configFile string

	// Display flags
	outputMode   string
	responseOnly bool // legacy override
	statusOnly   bool // legacy override
	noColor      bool

	// Request configuration
	headers  []string
	timeout  string
	insecure bool
	urlFlag  string

	// Redirect flags
	followRedirects bool
	noRedirects     bool

	// Repeat/concurrency
	repeatTimes      int
	repeatConcurrent int

	// Authentication flags
	authBasic     string
	basicAuthUser string
	basicAuthPass string
	clientCert    string
	clientCertKey string
	customCA      string

	// Data input flags (new hybrid approach)
	bodyData     string
	templateFile string
	templateVars []string
	formFields   []string
	formMode     bool

	// Legacy file flag (for backwards compatibility)
	fileFlag string

	// Version flag
	versionFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "gulp [flags] [URL]",
	Short: "A fast HTTP client for APIs and web services",
	Long: `GULP is a powerful HTTP client designed for API testing and automation.
It supports JSON, YAML, form data, templates, and more.

Examples:
  # Simple GET request
  gulp https://api.example.com

  # POST with stdin data
  gulp -m POST https://api.example.com < data.json

  # POST with file input
  gulp -m POST --body @data.json https://api.example.com

  # Basic authentication
  gulp --auth-basic=user:pass https://api.example.com

  # Template processing with variables
  gulp --template @template.json --var name=John https://api.example.com

  # Form data submission
  gulp -m POST --form name=John --form age=30 https://api.example.com

Core Options:
  -m, --method METHOD       HTTP method (GET, POST, PUT, DELETE)
  -v, --verbose            Show detailed request/response info  
  -c, --config FILE        Configuration file (.gulp.yml)

Data Input:
  --body DATA              Request body (@file, @-, or inline)
  --template FILE          Process file as Go template
  --form FIELD=VAL         Add form field (repeat for multiple)
  --var KEY=VAL            Template variable (repeat for multiple)

Authentication:
  --auth-basic=USER:PASS   Basic authentication
  --auth-cert=CERT,KEY     Client certificate  
  --basic-auth-user USER   Basic auth username
  --basic-auth-pass PASS   Basic auth password
  --client-cert FILE       Client certificate file
  --client-cert-key FILE   Client certificate key

Request Options:
  -H, --header HEADER       Request header (repeat for multiple)
  -t, --timeout SECONDS     Request timeout in seconds (default 10)
  -i, --insecure           Disable TLS certificate verification
  -u, --url URL            Request URL (alternative to positional)

Output & Display:
  -o, --output MODE         Output mode: body, status, verbose
  -n, --no-color           Disable colored output
  -ro, --ro                 (Legacy) Only display response body
  -sco, --sco              (Legacy) Only display status code

Redirect Options:
  -fr, --follow-redirects   Enable following redirects
  -nr, --no-redirects       Disable following redirects

Load Testing:
  -rt, --repeat-times TIMES  Number of requests to make
  -rc, --repeat-concurrent CONCURRENT  Number of concurrent connections

Other Options:
  -f, --file FILE          (Legacy) File input (use --body instead)
  -v, --version            Show version information

Global Flags:
  -h, --help               Help for any command

Additional help topics:
  help [command]            Help about any command

Use "gulp [command] --help" for more information about a command.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runGulp(args)
	},
}

func init() {
	// Set custom help function for better organization
	rootCmd.SetHelpFunc(customHelpFunc)

	// === CORE OPTIONS ===
	rootCmd.Flags().StringVarP(&method, "method", "m", "GET", "HTTP method (GET, POST, PUT, DELETE)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed request/response info")
	rootCmd.Flags().StringVarP(&configFile, "config", "c", ".gulp.yml", "Configuration file (.gulp.yml)")

	// === OUTPUT & DISPLAY ===
	rootCmd.Flags().StringVar(&outputMode, "output", "", "Output mode: body, status, verbose")
	rootCmd.Flags().BoolVar(&responseOnly, "ro", false, "(Legacy) Only display response body")
	rootCmd.Flags().BoolVar(&statusOnly, "sco", false, "(Legacy) Only display status code")
	rootCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// === DATA INPUT ===
	rootCmd.Flags().StringVar(&bodyData, "body", "", "Request body (@file, @-, or inline)")
	rootCmd.Flags().StringVar(&templateFile, "template", "", "Process file as Go template")
	rootCmd.Flags().StringArrayVar(&templateVars, "var", []string{}, "Template variable (repeat for multiple)")
	rootCmd.Flags().StringArrayVar(&formFields, "form", []string{}, "Form field key=value or key=@file")
	rootCmd.Flags().BoolVar(&formMode, "form-mode", false, "Process stdin as form data")

	// === AUTHENTICATION ===
	rootCmd.Flags().StringVar(&authBasic, "auth-basic", "", "Basic authentication user:pass")
	rootCmd.Flags().StringVar(&basicAuthUser, "basic-auth-user", "", "Basic auth username")
	rootCmd.Flags().StringVar(&basicAuthPass, "basic-auth-pass", "", "Basic auth password")
	rootCmd.Flags().StringVar(&clientCert, "client-cert", "", "Client certificate file")
	rootCmd.Flags().StringVar(&clientCertKey, "client-cert-key", "", "Client certificate key file")
	rootCmd.Flags().StringVar(&customCA, "custom-ca", "", "Custom CA certificate file")

	// === REQUEST OPTIONS ===
	rootCmd.Flags().StringArrayVarP(&headers, "header", "H", []string{}, "Request header (repeat for multiple)")
	rootCmd.Flags().StringVar(&timeout, "timeout", "", fmt.Sprintf("Request timeout in seconds (default %d)", config.DefaultTimeout))
	rootCmd.Flags().BoolVar(&insecure, "insecure", false, "Disable TLS certificate verification")
	rootCmd.Flags().StringVar(&urlFlag, "url", "", "Request URL (alternative to positional)")

	// === REDIRECT OPTIONS ===
	rootCmd.Flags().BoolVar(&followRedirects, "follow-redirects", false, "Enable following redirects")
	rootCmd.Flags().BoolVar(&noRedirects, "no-redirects", false, "Disable following redirects")

	// === LOAD TESTING ===
	rootCmd.Flags().IntVar(&repeatTimes, "repeat-times", 1, "Number of requests to make")
	rootCmd.Flags().IntVar(&repeatConcurrent, "repeat-concurrent", 1, "Number of concurrent connections")

	// === LEGACY / COMPATIBILITY ===
	rootCmd.Flags().StringVar(&fileFlag, "file", "", "(Legacy) File input (use --body instead)")

	// === OTHER ===
	rootCmd.Flags().BoolVar(&versionFlag, "version", false, "Show version information")

	// Mark mutually exclusive flags
	rootCmd.MarkFlagsMutuallyExclusive("follow-redirects", "no-redirects")
	rootCmd.MarkFlagsMutuallyExclusive("body", "template", "form")
}

// customHelpFunc provides organized help output for Phase 3
func customHelpFunc(cmd *cobra.Command, args []string) {
	// Print description and examples
	fmt.Printf("GULP is a powerful HTTP client designed for API testing and automation.\n")
	fmt.Printf("It supports JSON, YAML, form data, templates, and more.\n\n")

	fmt.Printf("Examples:\n")
	fmt.Printf("  # Simple GET request\n")
	fmt.Printf("  gulp https://api.example.com\n\n")
	fmt.Printf("  # POST with stdin data\n")
	fmt.Printf("  gulp -m POST https://api.example.com < data.json\n\n")
	fmt.Printf("  # POST with file input\n")
	fmt.Printf("  gulp -m POST --body @data.json https://api.example.com\n\n")
	fmt.Printf("  # Basic authentication\n")
	fmt.Printf("  gulp --auth-basic=user:pass https://api.example.com\n\n")
	fmt.Printf("  # Template processing with variables\n")
	fmt.Printf("  gulp --template @template.json --var name=John https://api.example.com\n\n")
	fmt.Printf("  # Form data submission\n")
	fmt.Printf("  gulp -m POST --form name=John --form age=30 https://api.example.com\n\n")

	fmt.Printf("Usage:\n  %s\n\n", cmd.UseLine())

	// Core Options
	fmt.Printf("Core Options:\n")
	printFlag(cmd, "method", "m")
	printFlag(cmd, "verbose", "v")
	printFlag(cmd, "config", "c")
	fmt.Println()

	// Data Input
	fmt.Printf("Data Input:\n")
	printFlag(cmd, "body", "")
	printFlag(cmd, "template", "")
	printFlag(cmd, "var", "")
	printFlag(cmd, "form", "")
	printFlag(cmd, "form-mode", "")
	fmt.Println()

	// Authentication
	fmt.Printf("Authentication:\n")
	printFlag(cmd, "auth-basic", "")
	printFlag(cmd, "basic-auth-user", "")
	printFlag(cmd, "basic-auth-pass", "")
	printFlag(cmd, "client-cert", "")
	printFlag(cmd, "client-cert-key", "")
	printFlag(cmd, "custom-ca", "")
	fmt.Println()

	// Request Options
	fmt.Printf("Request Options:\n")
	printFlag(cmd, "header", "H")
	printFlag(cmd, "timeout", "")
	printFlag(cmd, "insecure", "")
	printFlag(cmd, "url", "")
	fmt.Println()

	// Output & Display
	fmt.Printf("Output & Display:\n")
	printFlag(cmd, "output", "")
	printFlag(cmd, "no-color", "")
	printFlag(cmd, "ro", "")
	printFlag(cmd, "sco", "")
	fmt.Println()

	// Redirect Options
	fmt.Printf("Redirect Options:\n")
	printFlag(cmd, "follow-redirects", "")
	printFlag(cmd, "no-redirects", "")
	fmt.Println()

	// Load Testing
	fmt.Printf("Load Testing:\n")
	printFlag(cmd, "repeat-times", "")
	printFlag(cmd, "repeat-concurrent", "")
	fmt.Println()

	// Other Options
	fmt.Printf("Other Options:\n")
	printFlag(cmd, "version", "")
	printFlag(cmd, "file", "")
}

// printFlag formats and prints a flag with its usage
func printFlag(cmd *cobra.Command, flagName, shorthand string) {
	flag := cmd.Flags().Lookup(flagName)
	if flag == nil {
		return
	}

	name := "--" + flagName
	if shorthand != "" {
		name = "-" + shorthand + ", " + name
	}

	// Pad the name to align descriptions
	fmt.Printf("  %-25s %s\n", name, flag.Usage)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		output.ExitErr("", err)
	}
}

// runGulp contains the main application logic
func runGulp(args []string) error {
	// Handle version flag first
	if versionFlag {
		return handleVersionFlag()
	}

	// Load configuration
	loadedConfig, err := config.LoadConfiguration(configFile)
	if err != nil {
		return err
	}
	gulpConfig = loadedConfig

	// Apply configuration defaults before processing flags
	applyConfigurationDefaults()

	// Process flags and configuration
	processDisplayFlags()
	disableColorOutput()

	// Get the target URL
	url, err := getTargetURL(args)
	if err != nil {
		return err
	}

	// Build request
	body, headers, err := processRequestData()
	if err != nil {
		return err
	}

	// Configure TLS and redirects
	disableTLSVerify()
	followRedirect := shouldFollowRedirects()

	return executeRequestsWithConcurrency(url, body, headers, followRedirect)
}

// applyConfigurationDefaults applies config values when flags weren't explicitly set
func applyConfigurationDefaults() {
	// Apply method from config if not set via flag
	if method == "GET" && gulpConfig.GetMethod() != "GET" {
		method = gulpConfig.GetMethod()
	}

	// Apply output mode from config if not set via flag
	if outputMode == "" && gulpConfig.GetOutput() != "body" {
		outputMode = gulpConfig.GetOutput()
	}

	// Apply repeat settings from config if not set via flags
	if repeatTimes == 1 && gulpConfig.GetRepeatTimes() != 1 {
		repeatTimes = gulpConfig.GetRepeatTimes()
	}
	if repeatConcurrent == 1 && gulpConfig.GetRepeatConcurrent() != 1 {
		repeatConcurrent = gulpConfig.GetRepeatConcurrent()
	}

	// Apply request settings from config if not set via flags
	if !insecure && gulpConfig.Request.Insecure {
		insecure = true
	}
}

// getTargetURL determines the target URL from flags or arguments
func getTargetURL(args []string) (string, error) {
	var targetURL string

	// Command line args take precedence over flags
	if len(args) > 0 {
		targetURL = args[0]
	} else if urlFlag != "" {
		targetURL = urlFlag
	} else if gulpConfig.URL != "" {
		targetURL = gulpConfig.URL
	} else {
		return "", fmt.Errorf("need a URL: provide via argument, --url flag, or config file")
	}

	return client.BuildURL(targetURL, gulpConfig.URL)
}

// processDisplayFlags handles the display flag logic with precedence
func processDisplayFlags() {
	// Handle new --output flag first (highest precedence)
	if outputMode != "" {
		switch strings.ToLower(outputMode) {
		case "body":
			responseOnly = true
			statusOnly = false
			verbose = false
		case "status":
			responseOnly = false
			statusOnly = true
			verbose = false
		case "verbose":
			responseOnly = false
			statusOnly = false
			verbose = true
		}
		return
	}

	// Handle legacy override flags
	if responseOnly || statusOnly || verbose {
		return // Explicit flags take precedence
	}

	// Use new config.Output field first (v1.0)
	if gulpConfig.Output != "" && gulpConfig.Output != "body" {
		switch strings.ToLower(gulpConfig.Output) {
		case "status":
			statusOnly = true
		case "verbose":
			verbose = true
		default:
			responseOnly = true
		}
		return
	}

	// Fall back to legacy config.Display field for backwards compatibility
	switch gulpConfig.Display {
	case "status-code-only":
		statusOnly = true
	case "verbose":
		verbose = true
	default:
		responseOnly = true // Default behavior
	}
}

// processRequestData handles the enhanced hybrid data input approach
func processRequestData() ([]byte, map[string]string, error) {
	// Get request body
	body, err := getRequestBody()
	if err != nil {
		return nil, nil, err
	}

	// Build headers
	headerMap, err := client.BuildHeaders(headers, gulpConfig.Headers, body != nil)
	if err != nil {
		return nil, nil, err
	}

	// Process form data if needed
	if formMode && body != nil {
		processedBody, contentType, err := form.ProcessFormData(body)
		if err != nil {
			return nil, nil, err
		}
		headerMap["Content-Type"] = contentType
		return processedBody, headerMap, nil
	}

	// Handle form fields
	if len(formFields) > 0 {
		return processFormFields(headerMap)
	}

	// Convert JSON/YAML if needed
	if body != nil && !formMode {
		convertedBody, err := convertJSONBody(body, headerMap)
		if err != nil {
			return nil, nil, err
		}
		return convertedBody, headerMap, nil
	}

	return body, headerMap, nil
}

// buildAuthConfig creates authentication configuration from flags
func buildAuthConfig() (config.AuthConfig, error) {
	auth := gulpConfig.GetAuthConfig()

	// Handle --auth-basic convenience flag
	if authBasic != "" {
		parts := strings.SplitN(authBasic, ":", 2)
		if len(parts) != 2 {
			return auth, fmt.Errorf("--auth-basic must be in format 'username:password'")
		}
		auth.Basic.Username = parts[0]
		auth.Basic.Password = parts[1]
	}

	// Apply command line overrides
	if clientCert != "" {
		auth.Certificate.Cert = clientCert
	}
	if clientCertKey != "" {
		auth.Certificate.Key = clientCertKey
	}
	if customCA != "" {
		auth.Certificate.CA = customCA
	}
	if basicAuthUser != "" {
		auth.Basic.Username = basicAuthUser
	}
	if basicAuthPass != "" {
		auth.Basic.Password = basicAuthPass
	}

	return auth, nil
}

// getRequestBody implements the enhanced hybrid approach for request data
func getRequestBody() ([]byte, error) {
	// Skip body for GET/HEAD requests
	if method == "GET" || method == "HEAD" {
		return nil, nil
	}

	// Priority: CLI flags > config data > stdin
	// 1. Check CLI flags first
	if bodyData != "" {
		return processBodyFlag(bodyData)
	}

	if templateFile != "" {
		return processTemplateFlag(templateFile)
	}

	// 2. Check configuration data settings
	if gulpConfig.Data.Body != "" {
		// Check if config body contains template variables and we have variables to substitute
		if len(gulpConfig.Data.Variables) > 0 || len(templateVars) > 0 {
			// Merge config variables with CLI variables (CLI takes precedence)
			allVars := make([]string, 0, len(gulpConfig.Data.Variables)+len(templateVars))

			// Add config variables first
			for key, value := range gulpConfig.Data.Variables {
				allVars = append(allVars, fmt.Sprintf("%s=%s", key, value))
			}

			// Add CLI variables (these will override config variables with same keys)
			allVars = append(allVars, templateVars...)

			// Process as inline template
			return template.ProcessInlineTemplate(gulpConfig.Data.Body, allVars)
		}

		// No variables, process as regular body data
		return processBodyFlag(gulpConfig.Data.Body)
	}

	if gulpConfig.Data.Template != "" {
		// Merge config variables with any CLI variables (CLI takes precedence)
		configVars := make([]string, 0, len(gulpConfig.Data.Variables)+len(templateVars))

		// Add config variables first
		for key, value := range gulpConfig.Data.Variables {
			configVars = append(configVars, fmt.Sprintf("%s=%s", key, value))
		}

		// Add CLI variables (these will override config variables with same keys)
		configVars = append(configVars, templateVars...)

		// Process template with merged variables
		return template.ProcessTemplate(gulpConfig.Data.Template, configVars)
	}

	// Check config form data
	if len(gulpConfig.Data.Form) > 0 && len(formFields) == 0 {
		// Convert config form data to CLI format for processing
		for key, value := range gulpConfig.Data.Form {
			formFields = append(formFields, fmt.Sprintf("%s=%s", key, value))
		}
		if gulpConfig.Data.FormMode {
			formMode = true
		}
	}

	// 3. Check legacy file flag support
	if fileFlag != "" {
		if len(templateVars) > 0 {
			// Template processing enabled by presence of vars
			return template.ProcessTemplate(fileFlag, templateVars)
		}
		return os.ReadFile(fileFlag)
	}

	// 4. Check for stdin
	return getPostBodyFromStdin()
}

// processBodyFlag handles the --body flag (@file, @-, or inline)
func processBodyFlag(data string) ([]byte, error) {
	if strings.HasPrefix(data, "@") {
		filename := data[1:]
		if filename == "-" {
			return readAndProcessStdin()
		}
		return os.ReadFile(filename)
	}
	// Inline data
	return []byte(data), nil
}

// processTemplateFlag handles the --template flag
func processTemplateFlag(tmplFile string) ([]byte, error) {
	var filename string
	if strings.HasPrefix(tmplFile, "@") {
		filename = tmplFile[1:]
	} else {
		filename = tmplFile
	}

	if filename == "-" {
		// Process stdin as template
		content, err := readAndProcessStdin()
		if err != nil {
			return nil, err
		}
		return template.ProcessStdin(content, templateVars)
	}

	return template.ProcessTemplate(filename, templateVars)
}

// processFormFields handles --form fields
func processFormFields(headers map[string]string) ([]byte, map[string]string, error) {
	// Convert form fields to form data format
	var formData []string
	for _, field := range formFields {
		formData = append(formData, field)
	}

	formInput := strings.Join(formData, "\n")
	processedBody, contentType, err := form.ProcessFormData([]byte(formInput))
	if err != nil {
		return nil, nil, err
	}

	headers["Content-Type"] = contentType
	return processedBody, headers, nil
}

// Legacy functions updated to work with new flag system

func disableColorOutput() {
	if noColor || !gulpConfig.UseColor() {
		output.NoColor(true)
	}
}

func disableTLSVerify() {
	if insecure || !gulpConfig.VerifyTLS() {
		client.DisableTLSVerification = true
		if verbose {
			output.Out.PrintWarning("TLS CHECKING IS DISABLED FOR THIS REQUEST")
		}
	}
}

func shouldFollowRedirects() bool {
	if followRedirects {
		return true
	}
	if noRedirects {
		return false
	}
	return gulpConfig.FollowRedirects()
}

func calculateTimeout() int {
	if timeout != "" {
		if val, err := parseTimeout(timeout); err == nil {
			return val
		}
	}
	return gulpConfig.GetTimeout()
}

func parseTimeout(timeoutStr string) (int, error) {
	// Handle simple integer (seconds)
	if val, err := parseInt(timeoutStr); err == nil {
		return val, nil
	}

	// Handle duration strings like "30s", "1m", etc.
	if duration, err := time.ParseDuration(timeoutStr); err == nil {
		return int(duration.Seconds()), nil
	}

	return 0, fmt.Errorf("invalid timeout format: %s", timeoutStr)
}

func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func handleVersionFlag() error {
	currentVersion := client.GetVersion()
	updateInfo, err := client.CheckForUpdates(currentVersion, 3*time.Second)

	if err != nil {
		output.Out.PrintVersion(currentVersion)
		if verbose {
			output.Out.PrintWarning(fmt.Sprintf("Could not check for updates: %s", err))
		}
	} else {
		output.Out.PrintVersionWithUpdates(
			currentVersion,
			updateInfo.HasUpdate,
			updateInfo.LatestVersion,
			updateInfo.UpdateURL,
		)
	}
	os.Exit(0)
	return nil
}

func executeRequestsWithConcurrency(url string, body []byte, headers map[string]string, followRedirect bool) error {
	maxChan := make(chan bool, repeatConcurrent)
	var wg sync.WaitGroup
	for i := 0; i < repeatTimes; i++ {
		wg.Add(1)
		maxChan <- true
		go func(iteration int, maxChan chan bool, wg *sync.WaitGroup) {
			defer wg.Done()
			defer func(maxChan chan bool) { <-maxChan }(maxChan)
			if repeatTimes > 1 {
				iteration++
			}
			processRequest(url, body, headers, iteration, followRedirect)
		}(i, maxChan, &wg)
	}
	wg.Wait()
	return nil
}

func processRequest(url string, body []byte, headers map[string]string, iteration int, followRedirect bool) {
	if err := executeHTTPRequest(url, body, headers, iteration, followRedirect); err != nil {
		output.ExitErr("", err)
	}
}

func executeHTTPRequest(url string, body []byte, headers map[string]string, iteration int, followRedirect bool) error {
	// Build auth config
	auth, err := buildAuthConfig()
	if err != nil {
		return fmt.Errorf("could not build auth config: %w", err)
	}

	// Create HTTP client
	timeout := calculateTimeout()
	httpClient, err := client.CreateClient(followRedirect, timeout, auth)
	if err != nil {
		return fmt.Errorf("could not create HTTP client: %w", err)
	}

	// Create request
	req, err := client.CreateRequest(method, url, body, headers, auth)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	// Print request if verbose
	if verbose || repeatTimes > 1 {
		printRequest(iteration, url, req.Header, req.ContentLength, req.Proto, output.Out)
	}

	// Execute request
	start := time.Now()
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	duration := time.Since(start).Seconds()
	handleResponse(resp, duration, output.Out)
	return nil
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
	if len(templateVars) > 0 {
		return template.ProcessStdin(stdin, templateVars)
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

func printRequest(iteration int, url string, headers map[string][]string, contentLength int64, protocol string, bo *output.BuffOut) {
	if !verbose {
		printIterationPrefix(iteration, bo)
		return
	}

	printIterationHeader(iteration, bo)

	if len(headers) == 0 {
		bo.PrintHeader(fmt.Sprintf("%s %s", method, url))
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
		fmt.Sprintf("%s %s", method, url),
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
	if statusOnly {
		fmt.Fprintln(bo.Out, resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	if verbose {
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
	if !verbose {
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
