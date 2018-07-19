package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
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
	"github.com/thoom/gulp/output"
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
	methodFlag          = flag.String("m", "GET", "The `method` to use: GET, POST, PUT, DELETE")
	configFlag          = flag.String("c", ".gulp.yml", "The `configuration` file to use")
	insecureFlag        = flag.Bool("insecure", false, "Disable TLS certificate checking")
	responseOnlyFlag    = flag.Bool("ro", false, "Only display the response body (default)")
	statusCodeOnlyFlag  = flag.Bool("sco", false, "Only display the response code")
	verboseFlag         = flag.Bool("v", false, "Display the response body along with various headers")
	timeoutFlag         = flag.String("timeout", "", "The number of `seconds` to wait before the connection times out "+fmt.Sprintf("(default %d)", config.DefaultTimeout))
	noColorFlag         = flag.Bool("no-color", false, "Disables color output for the request")
	followRedirectFlag  = flag.Bool("follow-redirect", false, "Enables following 3XX redirects")
	disableRedirectFlag = flag.Bool("no-redirect", false, "Disables following 3XX redirects")
	repeatFlag          = flag.Int("repeat-times", 1, "Number of `iteration`s to submit the request")
	concurrentFlag      = flag.Int("repeat-concurrent", 1, "Number of concurrent `connections` to use")
	versionFlag         = flag.Bool("version", false, "Display the current client version")
)

func main() {
	flag.Var(&reqHeaders, "H", "Set a `request` header")
	flag.Parse()

	// Load the custom configuration
	loadedConfig, err := config.LoadConfiguration(*configFlag)
	if err != nil {
		output.ExitErr("", err)
	}

	// Set the main config to the one that was loaded
	gulpConfig = loadedConfig

	// Disable color output for the request
	disableColorOutput()

	// Make sure that the displayFlags are set appropriately
	filterDisplayFlags()

	if *versionFlag {
		output.Out.PrintVersion(client.GetVersion())
		os.Exit(0)
	}

	path := ""
	if len(flag.Args()) > 0 {
		path = flag.Args()[0]
	}

	url, err := client.BuildURL(path, gulpConfig.URL)
	if err != nil {
		output.ExitErr("", err)
	}

	// Don't check the TLS bro
	disableTLSVerify()

	// If the disableRedirectFlag is false and follow redirects is false, then set the flag to true
	followRedirect := shouldFollowRedirects()

	var body []byte
	// Don't get the post body if it's a GET/HEAD request
	if *methodFlag != "GET" && *methodFlag != "HEAD" {
		var err error
		body, err = getPostBody(os.Stdin)
		if err != nil {
			output.ExitErr("", err)
		}
	}

	// Build request headers
	headers, err := client.BuildHeaders(reqHeaders, gulpConfig.Headers, body != nil)
	if err != nil {
		output.ExitErr("", err)
	}

	// Convert the YAML/JSON body if necessary
	convertJSONBody(body, headers)

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
}

func processRequest(url string, body []byte, headers map[string]string, iteration int, followRedirect bool) {
	var startTimer time.Time

	req, err := client.CreateRequest(*methodFlag, url, body, headers)
	if err != nil {
		output.ExitErr("", err)
	}

	b := &bytes.Buffer{}
	defer fmt.Print(b)
	bo := &output.BuffOut{Out: b, Err: b}

	// If we got a request, output what was created
	if req != nil {
		printRequest(iteration, url, headers, bo)
	}

	startTimer = time.Now()
	resp, err := client.CreateClient(followRedirect, calculateTimeout()).Do(req)
	if err != nil {
		output.ExitErr("Something unexpected happened", err)
	}

	handleResponse(resp, time.Now().Sub(startTimer).Seconds(), bo)
}

func printRequest(iteration int, url string, headers map[string]string, bo *output.BuffOut) {
	if !*verboseFlag {
		if iteration > 0 {
			fmt.Fprintf(bo.Out, "%d: ", iteration)
		}
		return
	}

	if iteration > 0 {
		bo.PrintHeader(fmt.Sprintf("Iteration #%d", iteration))
	}

	urlHeader := fmt.Sprintf("%s %s", *methodFlag, url)
	if len(headers) == 0 {
		bo.PrintHeader(urlHeader)
		return
	}

	block := []string{urlHeader}
	mk := make([]string, len(headers))
	i := 0
	for k := range headers {
		mk[i] = k
		i++
	}
	sort.Strings(mk)

	for _, k := range mk {
		block = append(block, strings.ToUpper(k)+": "+headers[k])
	}
	bo.PrintBlock(strings.Join(block, "\n"))
	fmt.Fprintln(bo.Out)
}

func handleResponse(resp *http.Response, duration float64, bo *output.BuffOut) {
	if *statusCodeOnlyFlag {
		fmt.Fprintln(bo.Out, resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if *verboseFlag {
		bo.PrintStoplight(fmt.Sprintf("Status: %s (%.2f seconds)\n", resp.Status, duration), resp.StatusCode >= 400)
	}

	isJSON := false
	mk := make([]string, len(resp.Header))
	i := 0
	for k := range resp.Header {
		mk[i] = k
		i++
	}
	sort.Strings(mk)

	for _, k := range mk {
		if k == "Content-Type" && strings.Contains(resp.Header.Get(k), "json") {
			isJSON = true
		}
		if *verboseFlag {
			fmt.Fprintln(bo.Out, strings.ToUpper(k)+": "+resp.Header.Get(k))
		}
	}

	if *verboseFlag {
		fmt.Fprintln(bo.Out, "")
	}

	if isJSON && *verboseFlag {
		var prettyJSON bytes.Buffer
		err := json.Indent(&prettyJSON, body, "", "  ")
		if err == nil {
			// Don't worry about pretty-printing if we got an error
			body = prettyJSON.Bytes()
		}
	}

	fmt.Fprintln(bo.Out, string(body))
}

func getPostBody(input *os.File) ([]byte, error) {
	stat, _ := input.Stat()

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(input)
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
			return nil, fmt.Errorf("Reading standard input: %s", err)
		}

		return stdin, nil
	}

	return nil, nil
}

func convertJSONBody(body []byte, headers map[string]string) ([]byte, error) {
	// Determine if we should convert the body to JSON
	if !strings.Contains(headers["CONTENT-TYPE"], "json") {
		return body, nil
	}

	j, err := yaml.YAMLToJSON(body)
	if err != nil {
		return nil, fmt.Errorf("Could not parse post body: %s", err)
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
	redirectFlags := 0
	if *disableRedirectFlag {
		redirectFlags++
	}

	if *followRedirectFlag {
		redirectFlags++
	}

	// If we don't have either flag set, use the config
	if redirectFlags == 0 {
		return gulpConfig.FollowRedirects()
	}

	// If only one of the flags is set, use the flag passed
	if redirectFlags > 1 {
		totalArgs := len(os.Args[1:])
		*disableRedirectFlag = false
		*followRedirectFlag = false
		for i := totalArgs; i > 0; i-- {
			switch os.Args[i] {
			case "-no-redirect":
				*disableRedirectFlag = true
				break
			case "-follow-redirect":
				*followRedirectFlag = true
				break
			default:
				continue
			}
			break
		}
	}

	if *disableRedirectFlag {
		return false
	}
	return true
}

func filterDisplayFlags() {
	displayFlags := 0
	if *responseOnlyFlag {
		displayFlags++
	}

	if *statusCodeOnlyFlag {
		displayFlags++
	}

	if *verboseFlag {
		displayFlags++
	}

	// If only one was set then we can just return
	if displayFlags == 1 {
		return
	}

	// If none were set, then use the configuration loaded
	if displayFlags == 0 {
		switch gulpConfig.Display {
		case "status-code-only":
			*statusCodeOnlyFlag = true
		case "verbose":
			*verboseFlag = true
		default:
			*responseOnlyFlag = true
		}
		return
	}

	// If multiple were set, then we need to figure out which one was the last one set and use that instead
	totalArgs := len(os.Args[1:])
	*responseOnlyFlag = false
	*statusCodeOnlyFlag = false
	*verboseFlag = false
	for i := totalArgs; i > 0; i-- {
		switch os.Args[i] {
		case "-ro":
			*responseOnlyFlag = true
			break
		case "-sco":
			*statusCodeOnlyFlag = true
			break
		case "-v":
			*verboseFlag = true
			break
		default:
			continue
		}
		break
	}
}
