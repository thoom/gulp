package client

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/thoom/gulp/config"
)

var buildVersion string

// DisableTLSVerification disables TLS verification
func DisableTLSVerification() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
}

// CreateRequest will create a request object
func CreateRequest(method, url string, body []byte, headers map[string]string) (*http.Request, error) {
	var reader io.Reader

	// Don't build the reader if using a GET/HEAD request
	if method != "GET" && method != "HEAD" && body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("could not build request: %s", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

// CreateClient will create a new http.Client with basic defaults
func CreateClient(followRedirects bool, timeout int, clientCert config.ClientAuth) (*http.Client, error) {
	tr := &http.Transport{
		DisableCompression: false,
	}

	if clientCert.UseAuth() {
		cert, err := tls.LoadX509KeyPair(clientCert.Cert, clientCert.Key)
		if err != nil {
			return nil, fmt.Errorf("invalid client cert/key: %s", err)
		}

		tr.TLSClientConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}

	if !followRedirects {
		return &http.Client{
			Timeout:   time.Duration(timeout) * time.Second,
			Transport: tr,
			CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}, nil
	}

	return &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: tr,
	}, nil
}
