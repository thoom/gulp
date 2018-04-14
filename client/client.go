package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var buildVersion string

// DisableTLSVerification disables TLS verification
func DisableTLSVerification() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

// CreateRequest will create a request object
func CreateRequest(method string, url string, body string, headers map[string]string) (*http.Request, error) {
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("Could not build request: %s", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

// CreateResponse processes the request and returns the response
func CreateResponse(request *http.Request, followRedirects bool) (*http.Response, error) {
	httpClient := &http.Client{}
	if !followRedirects {
		httpClient = &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	return httpClient.Do(request)
}
