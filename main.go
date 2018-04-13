package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
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

// VERSION references the current CLI revision
const VERSION = "0.6"

var (
	reqHeaders stringSlice

	gulpConfig         = config.New
	methodFlag         = flag.String("m", "GET", "The `method` to use: GET, POST, PUT, DELETE")
	configFlag         = flag.String("c", ".gulp.yml", "The `configuration` file to use")
	insecureFlag       = flag.Bool("k", false, "Insecure TLS communication")
	responseOnlyFlag   = flag.Bool("ro", false, "Only display the response body (default)")
	statusCodeOnlyFlag = flag.Bool("sco", false, "Only display the response code")
	verboseFlag        = flag.Bool("I", false, "Display the response body along with various headers")
	noColorFlag        = flag.Bool("no-color", false, "Disables color output for the request")
	repeatFlag         = flag.Int("repeat-times", 1, "Number of `iteration`s to submit the request")
	concurrentFlag     = flag.Int("repeat-concurrent", 1, "Number of concurrent `connections` to use")
	versionFlag        = flag.Bool("version", false, "Display the current client version")
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
	if *noColorFlag || !gulpConfig.UseColor() {
		output.NoColor(true)
	}

	// Make sure that the displayFlags are set appropriately
	filterDisplayFlags()

	if *versionFlag {
		output.PrintBlock(fmt.Sprintf(`thoom.Gulp
version: %s
author: Z.d.Peacock <zdp@thoomtech.com>
link: https://github.com/thoom/gulp`, VERSION), nil)
		fmt.Println()
		os.Exit(0)
	}

	url, err := buildURL()
	if err != nil {
		output.ExitErr("", err)
	}

	if *insecureFlag || !gulpConfig.VerifyTLS() {
		if *verboseFlag {
			output.PrintWarning("TLS checking is disabled for this request", nil)
		}
		client.DisableTLSVerification()
	}

	// Don't get the post body if it's a GET/HEAD request
	body := ""
	if *methodFlag != "GET" && *methodFlag != "HEAD" {
		body = getPostBody()
	}

	iteration := 0
	for i := 0; i < *repeatFlag; i += *concurrentFlag {
		var wg sync.WaitGroup

		ci := *concurrentFlag
		if i >= *concurrentFlag {
			remaining := *repeatFlag - i
			if remaining < ci {
				ci = remaining
			}
		}

		for c := 0; c < ci; c++ {
			wg.Add(1)
			go func(url string, body string, i int, c int) {
				defer wg.Done()
				var startTimer time.Time

				b := &bytes.Buffer{}
				defer fmt.Print(b)

				if *repeatFlag > 1 {
					iteration++
					if *verboseFlag {
						output.PrintHeader(fmt.Sprintf("Iteration #%d", iteration), b)
					} else {
						fmt.Fprintf(b, "%d: ", iteration)
					}
				}

				headers, err := buildHeaders(reqHeaders, body != "")
				if err != nil {
					output.ExitErr("", err)
				}

				req, err := client.CreateRequest(*methodFlag, url, body, headers)
				if err != nil {
					output.ExitErr("", err)
				}

				if *verboseFlag && req != nil {
					block := []string{fmt.Sprintf("%s %s", *methodFlag, url)}
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
					output.PrintBlock(strings.Join(block, "\n"), b)
					fmt.Fprintln(b)
				}

				startTimer = time.Now()
				resp, err := client.CreateResponse(req)
				if err != nil {
					output.ExitErr("Something unexpected happened", err)
				}

				handleResponse(resp, time.Now().Sub(startTimer).Seconds(), b)
			}(url, body, i, c)
		}
		wg.Wait()
	}
}

func buildURL() (string, error) {
	url := ""
	path := ""
	if len(flag.Args()) > 0 {
		path = flag.Args()[0]
	}

	var err error
	if strings.HasPrefix(path, "http") {
		url = path
	} else if gulpConfig.URL != "" {
		url = gulpConfig.URL + path
	}

	if url == "" {
		if path == "" {
			err = fmt.Errorf("Need a URL to make a request")
		} else {
			err = fmt.Errorf("Invalid URL")
		}
	}

	return url, err
}

func buildHeaders(reqHeaders []string, includeJSON bool) (map[string]string, error) {
	headers := make(map[string]string)

	// Set the default User-Agent and Accept type
	headers["USER-AGENT"] = fmt.Sprintf("thoom.Gulp/%s", VERSION)
	headers["ACCEPT"] = "application/json;q=1.0, */*;q=0.8"

	if includeJSON {
		headers["CONTENT-TYPE"] = "application/json"
	}

	for k, v := range gulpConfig.Headers {
		headers[strings.ToUpper(k)] = v
	}

	for _, header := range reqHeaders {
		pieces := strings.Split(header, ":")
		if len(pieces) != 2 {
			return nil, fmt.Errorf("Could not parse header: '%s'", header)
		}

		headers[strings.ToUpper(pieces[0])] = strings.TrimSpace(pieces[1])
	}

	return headers, nil
}

func handleResponse(resp *http.Response, duration float64, writer io.Writer) {
	if writer == nil {
		writer = os.Stdout
	}

	if *statusCodeOnlyFlag {
		fmt.Fprintln(writer, resp.StatusCode)
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if *verboseFlag {
		output.PrintStoplight(fmt.Sprintf("Status: %s (%.2f seconds)\n", resp.Status, duration), resp.StatusCode >= 400, writer)
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
			fmt.Fprintln(writer, strings.ToUpper(k)+": "+resp.Header.Get(k))
		}
	}

	if *verboseFlag {
		fmt.Fprintln(writer, "")
	}

	if isJSON && *verboseFlag {
		var prettyJSON bytes.Buffer
		err := json.Indent(&prettyJSON, body, "", "  ")
		if err == nil {
			// Don't worry about pretty-printing if we got an error
			body = prettyJSON.Bytes()
		}
	}

	fmt.Fprintln(writer, string(body))
}

func getPostBody() string {
	stat, _ := os.Stdin.Stat()
	body := ""

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		var stdin []byte
		for scanner.Scan() {
			stdin = append(append(stdin, scanner.Bytes()...), []byte("\n")...)
		}

		if err := scanner.Err(); err != nil {
			output.ExitErr("Reading standard input", err)
		}

		j, err := yaml.YAMLToJSON(stdin)
		if err != nil {
			output.ExitErr("Could not parse post body", err)
		}

		body = string(j)
	}

	return body
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
		case "-I":
			*verboseFlag = true
			break
		default:
			continue
		}
		break
	}
}
