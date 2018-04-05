package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ghodss/yaml"
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

	config           = NewConfig()
	methodFlag       = flag.String("m", "GET", "The `method` to use: GET, POST, PUT, DELETE")
	configFlag       = flag.String("c", ".gulp.yml", "The `configuration` file to use")
	insecureFlag     = flag.Bool("k", false, "Insecure TLS communication")
	responseOnlyFlag = flag.Bool("ro", false, "Only display the response body (default)")
	successOnlyFlag  = flag.Bool("so", false, "Only display whether or not the request was successful")
	verboseFlag      = flag.Bool("I", false, "Display the response body along with various headers")
	repeatFlag       = flag.Int("repeat", 1, "Number of `iteration`s to submit the request")
	groupFlag        = flag.Int("repeat-group", 1, "Number of `concurrent connections` when grouping")
)

func main() {
	flag.Var(&reqHeaders, "H", "Set a `request` header")
	flag.Parse()

	config = loadConfiguration(*configFlag)
	// Set the flag based on the configuration if none of the flags are set
	if !*responseOnlyFlag && !*successOnlyFlag && !*verboseFlag {
		switch config.Display {
		case "success-only":
			*successOnlyFlag = true
		case "verbose":
			*verboseFlag = true
		default:
			*responseOnlyFlag = true
		}
	}

	url := ""
	path := ""
	if len(flag.Args()) > 0 {
		path = flag.Args()[0]
	}

	if strings.HasPrefix(path, "http") {
		url = path
	} else if config.URL != "" {
		url = config.URL + path
	}

	if url == "" {
		ExitErr("Need a URL to make a request", nil)
	}

	if *insecureFlag || !config.TLSVerify() {
		if *verboseFlag {
			PrintWarning("TLS checking is disabled for this request", nil)
		}
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	// Don't get the post body if it's a GET/HEAD request
	body := ""
	if *methodFlag != "GET" && *methodFlag != "HEAD" {
		body = getPostBody()
	}

	iteration := 0
	for i := 0; i < *repeatFlag; i += *groupFlag {
		var wg sync.WaitGroup

		ci := *groupFlag
		if i >= *groupFlag {
			remaining := *repeatFlag - i
			if remaining < ci {
				ci = remaining
			}
		}

		for c := 0; c < ci; c++ {
			wg.Add(1)
			go func(url string, body string, i int) {
				defer wg.Done()
				var startTimer, endTimer time.Time

				b := &bytes.Buffer{}
				defer fmt.Print(b)

				if *repeatFlag > 1 {
					iteration++
					PrintHeader(fmt.Sprintf("Iteration #%d", iteration), b)
				}

				req := createRequest(*methodFlag, url, body, b)
				client := &http.Client{}

				startTimer = time.Now()
				resp, err := client.Do(req)
				endTimer = time.Now()

				if err != nil {
					ExitErr("Something unexpected happened", err)
				}

				handleResponse(resp, endTimer.Sub(startTimer).Seconds(), b)
			}(url, body, i)
		}
		wg.Wait()
	}
}

func createRequest(method string, url string, body string, writer io.Writer) *http.Request {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		ExitErr("Could not build request", err)
	}

	// Set the default User-Agent
	req.Header.Set("User-Agent", "thoom.Gulp/0.2")

	// If the reader is empty, then we didn't have a post body
	if reader != nil {
		// We onlly allow json bodies
		req.Header.Set("Content-Type", "application/json")
	}

	for k, v := range config.Headers {
		req.Header.Set(k, v)
	}

	for _, header := range reqHeaders {
		pieces := strings.Split(header, ":")
		req.Header.Set(pieces[0], strings.TrimSpace(pieces[1]))
	}

	if *verboseFlag {
		if writer == nil {
			writer = os.Stdout
		}

		block := []string{fmt.Sprintf("%s %s", *methodFlag, url)}
		for k, v := range req.Header {
			for _, h := range v {
				block = append(block, strings.ToUpper(k)+": "+h)
			}
		}
		PrintBlock(strings.Join(block, "\n"), writer)
		fmt.Fprintln(writer)

		// Output the post body if one was passed in
		// if reader != nil {
		// 	var prettyJSON bytes.Buffer
		// 	err := json.Indent(&prettyJSON, []byte(body), "", "  ")
		// 	if err == nil {
		// 		// Don't worry about pretty-printing if we got an error
		// 		fmt.Println(string(prettyJSON.Bytes()) + "\n")
		// 	}
		// }
	}

	return req
}

func handleResponse(resp *http.Response, duration float64, writer io.Writer) {
	if writer == nil {
		writer = os.Stdout
	}

	if *successOnlyFlag {
		fmt.Fprintln(writer, resp.StatusCode < 400)
		return
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if *verboseFlag {
		PrintStoplight(fmt.Sprintf("Status: %s (%.2f seconds)\n", resp.Status, duration), resp.StatusCode >= 400, writer)
	}

	isJSON := false
	for k, v := range resp.Header {
		if k == "Content-Type" && strings.Contains(strings.Join(v, ","), "application/json") {
			isJSON = true
		}
		if *verboseFlag {
			fmt.Fprintln(writer, strings.ToUpper(k)+": "+strings.Join(v, ","))
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

func loadConfiguration(fileName string) GulpConfig {
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		// If the file wasn't found and it's just the default, don't worry about it.
		if fileName == ".gulp.yml" {
			return config
		}

		ExitErr(fmt.Sprintf("Could not load configuration '%s'", fileName), nil)
	}

	var gulpConfig GulpConfig
	if yaml.Unmarshal(dat, &gulpConfig) != nil {
		ExitErr("Could not parse configuration", nil)
	}

	return gulpConfig
}

func getPostBody() string {
	stat, _ := os.Stdin.Stat()
	body := ""

	if (stat.Mode() & os.ModeCharDevice) == 0 {
		scanner := bufio.NewScanner(os.Stdin)
		stdin := ""
		for scanner.Scan() {
			stdin += scanner.Text() + "\n"
		}

		if err := scanner.Err(); err != nil {
			ExitErr("Reading standard input", err)
		}

		j, err := yaml.YAMLToJSON([]byte(stdin))
		if err != nil {
			ExitErr("Could not parse post body", err)
		}

		body = string(j)
	}

	return body
}
