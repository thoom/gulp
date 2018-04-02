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

type gulpConfig struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Display string            `yaml:"display"`
}

var (
	reqHeaders stringSlice

	config           = gulpConfig{}
	methodFlag       = flag.String("m", "GET", "The `method` to use: GET, POST, PUT, DELETE")
	configFlag       = flag.String("c", ".gulp.yml", "The `configuration` file to use")
	responseOnlyFlag = flag.Bool("dr", false, "Only display the response body (default)")
	successOnlyFlag  = flag.Bool("ds", false, "Only display whether or not the request was successful")
	verboseFlag      = flag.Bool("dv", false, "Display the response body along with various headers")
)

func main() {
	flag.Var(&reqHeaders, "H", "Set a `request` header")
	flag.Parse()

	config = loadConfiguration(*configFlag)
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
		fmt.Println("Need something to access")
		os.Exit(1)
	}

	if *verboseFlag {
		fmt.Println("url: " + url)
	}

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := &http.Client{}
	body := ""

	// Don't get the post body if it's a GET/HEAD request
	if *methodFlag != "GET" && *methodFlag != "HEAD" {
		body = getPostBody()
	}

	resp, err := client.Do(createRequest(*methodFlag, url, body))
	if err != nil {
		fmt.Println("somethun happened: ", err)
		os.Exit(0)
	}

	handleResponse(resp)
}

func createRequest(method string, url string, body string) *http.Request {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		fmt.Println("could not build request: ", err)
		os.Exit(1)
	}

	// If the reader is empty, then we didn't have a post body
	if reader != nil {
		if *verboseFlag {
			fmt.Println("body: " + body)
		}

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

	return req
}

func handleResponse(resp *http.Response) {
	if *successOnlyFlag {
		fmt.Println(resp.StatusCode >= 200 && resp.StatusCode < 300)
		os.Exit(0)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if *verboseFlag {
		fmt.Println("Status: " + resp.Status + "\n")
	}

	isJSON := false
	for k, v := range resp.Header {
		if k == "Content-Type" && strings.Contains(strings.Join(v, ","), "application/json") {
			isJSON = true
		}
		if *verboseFlag {
			fmt.Println(k + ": " + strings.Join(v, ","))
		}
	}

	if *verboseFlag {
		fmt.Println("")
	}

	if isJSON {
		var prettyJSON bytes.Buffer
		err := json.Indent(&prettyJSON, body, "", "  ")
		if err == nil {
			// Don't worry about pretty-printing if we got an error
			body = prettyJSON.Bytes()
		}
	}

	fmt.Println(string(body))
	os.Exit(0)
}

func loadConfiguration(fileName string) gulpConfig {
	dat, err := ioutil.ReadFile(fileName)
	if err != nil {
		// If the file wasn't found and it's just the default, don't worry about it.
		if fileName == ".gulp.yml" {
			return gulpConfig{}
		}

		fmt.Println("Configuration file not found")
		os.Exit(1)
	}

	var config gulpConfig
	if yaml.Unmarshal(dat, &config) != nil {
		fmt.Println("Could not parse configuration")
		os.Exit(1)
	}

	return config
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
			fmt.Fprintln(os.Stderr, "reading standard input:", err)
			os.Exit(1)
		}

		j, err := yaml.YAMLToJSON([]byte(stdin))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		body = string(j)
	}

	return body
}
