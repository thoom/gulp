package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func DisableTLSVerification() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

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

func CreateResponse(request *http.Request) (*http.Response, error) {
	httpClient := &http.Client{}
	return httpClient.Do(request)
}
