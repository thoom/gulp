package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

var buildVersion string

// DisableTLSVerification disables TLS verification
func DisableTLSVerification() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

// CreateRequest will create a request object
func CreateRequest(method string, url string, body string, headers map[string]string) (*http.Request, error) {
	var reader io.Reader

	// Don't build the read if using a GET/HEAD request
	if method != "GET" && method != "HEAD" && body != "" {
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

func CreateClient(followRedirects bool, timeout int) *http.Client {
	if !followRedirects {
		return &http.Client{
			Timeout: time.Duration(timeout) * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	return &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
}
