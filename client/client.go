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

// PEM header prefix constant
const pemHeaderPrefix = "-----BEGIN"

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
	transport, err := createHTTPTransport(clientCert)
	if err != nil {
		return nil, err
	}

	return buildHTTPClient(followRedirects, timeout, transport), nil
}

// createHTTPTransport creates an HTTP transport with TLS configuration
func createHTTPTransport(clientCert config.ClientAuth) (*http.Transport, error) {
	transport := &http.Transport{
		DisableCompression: false,
	}

	tlsConfig, err := buildTLSConfig(clientCert)
	if err != nil {
		return nil, err
	}

	// Only set TLS config if we have either custom CA or client certs
	if tlsConfig != nil {
		transport.TLSClientConfig = tlsConfig
	}

	return transport, nil
}

// buildTLSConfig creates TLS configuration for custom CA and client certificates
func buildTLSConfig(clientCert config.ClientAuth) (*tls.Config, error) {
	hasCA := strings.TrimSpace(clientCert.CA) != ""
	hasClientCert := clientCert.UseAuth()

	if !hasCA && !hasClientCert {
		return nil, nil // No TLS config needed
	}

	tlsConfig := &tls.Config{}

	if hasCA {
		if err := configureCACertificate(tlsConfig, clientCert.CA); err != nil {
			return nil, err
		}
	}

	if hasClientCert {
		if err := configureClientCertificate(tlsConfig, clientCert); err != nil {
			return nil, err
		}
	}

	return tlsConfig, nil
}

// configureCACertificate sets up custom CA certificate in TLS config
func configureCACertificate(tlsConfig *tls.Config, caData string) error {
	caCert, err := loadCertificateData(caData)
	if err != nil {
		return fmt.Errorf("could not read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return fmt.Errorf("failed to parse CA certificate")
	}

	tlsConfig.RootCAs = caCertPool
	return nil
}

// configureClientCertificate sets up client certificate in TLS config
func configureClientCertificate(tlsConfig *tls.Config, clientCert config.ClientAuth) error {
	cert, err := loadClientCertificatePair(clientCert.Cert, clientCert.Key)
	if err != nil {
		return fmt.Errorf("invalid client cert/key: %w", err)
	}

	tlsConfig.Certificates = []tls.Certificate{cert}
	return nil
}

// loadCertificateData loads certificate data from either inline PEM or file path
func loadCertificateData(data string) ([]byte, error) {
	trimmedData := strings.TrimSpace(data)

	// Check if it's direct PEM content (starts with -----BEGIN)
	if strings.HasPrefix(trimmedData, pemHeaderPrefix) {
		return []byte(trimmedData), nil
	}

	// It's a file path
	return os.ReadFile(trimmedData)
}

// loadClientCertificatePair loads a client certificate pair from cert and key data
func loadClientCertificatePair(certData, keyData string) (tls.Certificate, error) {
	certTrimmed := strings.TrimSpace(certData)
	keyTrimmed := strings.TrimSpace(keyData)

	certIsPEM := strings.HasPrefix(certTrimmed, pemHeaderPrefix)
	keyIsPEM := strings.HasPrefix(keyTrimmed, pemHeaderPrefix)

	// Both must be the same format (both PEM or both file paths)
	if certIsPEM != keyIsPEM {
		return tls.Certificate{}, fmt.Errorf("client certificate and key must both be either file paths or inline PEM content, not mixed")
	}

	if certIsPEM {
		// Both are direct PEM content
		return tls.X509KeyPair([]byte(certTrimmed), []byte(keyTrimmed))
	}

	// Both are file paths
	return tls.LoadX509KeyPair(certTrimmed, keyTrimmed)
}

// buildHTTPClient creates the final HTTP client with redirect and timeout configuration
func buildHTTPClient(followRedirects bool, timeout int, transport *http.Transport) *http.Client {
	client := &http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}

	if !followRedirects {
		client.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}

	return client
}

// ClientAuthBuilder helps build ClientAuth configurations
type ClientAuthBuilder struct {
	auth config.ClientAuth
}

// NewClientAuthBuilder creates a new builder with base configuration
func NewClientAuthBuilder(baseConfig config.ClientAuth) *ClientAuthBuilder {
	return &ClientAuthBuilder{auth: baseConfig}
}

// WithCert sets the client certificate
func (b *ClientAuthBuilder) WithCert(cert string) *ClientAuthBuilder {
	if strings.TrimSpace(cert) != "" {
		b.auth.Cert = cert
	}
	return b
}

// WithKey sets the client certificate key
func (b *ClientAuthBuilder) WithKey(key string) *ClientAuthBuilder {
	if strings.TrimSpace(key) != "" {
		b.auth.Key = key
	}
	return b
}

// WithCA sets the custom CA certificate
func (b *ClientAuthBuilder) WithCA(ca string) *ClientAuthBuilder {
	if strings.TrimSpace(ca) != "" {
		b.auth.CA = ca
	}
	return b
}

// WithBasicAuth sets basic authentication credentials
func (b *ClientAuthBuilder) WithBasicAuth(username, password string) *ClientAuthBuilder {
	if strings.TrimSpace(username) != "" {
		b.auth.Username = username
	}
	if strings.TrimSpace(password) != "" {
		b.auth.Password = password
	}
	return b
}

// Build returns the final ClientAuth configuration
func (b *ClientAuthBuilder) Build() config.ClientAuth {
	return b.auth
}

// BuildClientAuth creates a ClientAuth object (legacy compatibility)
// Deprecated: Use NewClientAuthBuilder for better API
func BuildClientAuth(clientCert, clientCertKey, clientCA, basicAuthUser, basicAuthPass string, clientCertConfig config.ClientAuth) config.ClientAuth {
	return NewClientAuthBuilder(clientCertConfig).
		WithCert(clientCert).
		WithKey(clientCertKey).
		WithCA(clientCA).
		WithBasicAuth(basicAuthUser, basicAuthPass).
		Build()
}
