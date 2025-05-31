package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/thoom/gulp/config"
)

var buildVersion string

// DisableTLSVerification disables TLS verification
func DisableTLSVerification() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			// This function will be called instead of the default verification
			// Here you can implement your own logic or simply return nil to bypass all checks
			return nil
		},
	}
}

// CreateRequest will create a request object
func CreateRequest(method, url string, body []byte, headers map[string]string, clientAuth config.ClientAuth) (*http.Request, error) {
	var reader io.Reader

	// Don't build the reader if using a GET/HEAD request
	if method != "GET" && method != "HEAD" && body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("could not build request: %s", err)
	}

	// Add basic auth header if credentials are provided
	if clientAuth.UseBasicAuth() {
		auth := base64.StdEncoding.EncodeToString([]byte(clientAuth.Username + ":" + clientAuth.Password))
		req.Header.Set("Authorization", "Basic "+auth)
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

	// Initialize TLS config
	tlsConfig := &tls.Config{}

	// Handle custom CA certificate
	if strings.TrimSpace(clientCert.CA) != "" {
		var caCert []byte
		var err error

		caData := strings.TrimSpace(clientCert.CA)

		// Check if it's direct PEM content (starts with -----BEGIN)
		if strings.HasPrefix(caData, "-----BEGIN") {
			// It's direct PEM content
			caCert = []byte(caData)
		} else {
			// It's a file path
			caCert, err = os.ReadFile(caData)
			if err != nil {
				return nil, fmt.Errorf("could not read CA certificate file: %s", err)
			}
		}

		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		tlsConfig.RootCAs = caCertPool
	}

	// Handle client certificate authentication
	if clientCert.UseAuth() {
		var cert tls.Certificate
		var err error

		certData := strings.TrimSpace(clientCert.Cert)
		keyData := strings.TrimSpace(clientCert.Key)

		// Check if both cert and key are direct PEM content
		if strings.HasPrefix(certData, "-----BEGIN") && strings.HasPrefix(keyData, "-----BEGIN") {
			// Both are direct PEM content
			cert, err = tls.X509KeyPair([]byte(certData), []byte(keyData))
		} else if !strings.HasPrefix(certData, "-----BEGIN") && !strings.HasPrefix(keyData, "-----BEGIN") {
			// Both are file paths
			cert, err = tls.LoadX509KeyPair(certData, keyData)
		} else {
			// Mixed format - one is inline, one is file path - this is an error
			return nil, fmt.Errorf("client certificate and key must both be either file paths or inline PEM content, not mixed")
		}

		if err != nil {
			return nil, fmt.Errorf("invalid client cert/key: %s", err)
		}

		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Only set TLS config if we have either custom CA or client certs
	if strings.TrimSpace(clientCert.CA) != "" || clientCert.UseAuth() {
		tr.TLSClientConfig = tlsConfig
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

// Creates a ClientAuth object
func BuildClientAuth(clientCert, clientCertKey, clientCA, basicAuthUser, basicAuthPass string, clientCertConfig config.ClientAuth) config.ClientAuth {
	clientAuth := clientCertConfig
	if strings.TrimSpace(clientCert) != "" {
		clientAuth.Cert = clientCert
	}

	if strings.TrimSpace(clientCertKey) != "" {
		clientAuth.Key = clientCertKey
	}

	if strings.TrimSpace(clientCA) != "" {
		clientAuth.CA = clientCA
	}

	if strings.TrimSpace(basicAuthUser) != "" {
		clientAuth.Username = basicAuthUser
	}

	if strings.TrimSpace(basicAuthPass) != "" {
		clientAuth.Password = basicAuthPass
	}

	return clientAuth
}
