package client

import (
	"fmt"
	"strings"
	"time"
)

var defaultVersion string

// DefaultVersion references the current CLI revision
func init() {
	defaultVersion = time.Now().Format("20060102.0304PM.MST") + "-SNAPSHOT"
}

// BuildURL will compute the final URL to use in the request
func BuildURL(path string, configURL string) (string, error) {
	url := ""

	var err error
	if strings.HasPrefix(path, "http") {
		url = path
	} else if configURL != "" {
		url = configURL + path
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

// BuildHeaders will return a map[string]string of headers
func BuildHeaders(reqHeaders []string, configHeaders map[string]string, includeJSON bool) (map[string]string, error) {
	headers := make(map[string]string)

	// Set the default User-Agent and Accept type
	headers["USER-AGENT"] = fmt.Sprintf("thoom.Gulp/%s", GetVersion())
	headers["ACCEPT"] = "application/json;q=1.0, */*;q=0.8"

	if includeJSON {
		headers["CONTENT-TYPE"] = "application/json"
	}

	for k, v := range configHeaders {
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

// GetVersion builds the version from the build branch
func GetVersion() string {
	version := buildVersion
	if version == "" {
		version = defaultVersion
	}

	return version
}
