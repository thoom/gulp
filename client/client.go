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

// DisableTLSVerification is used as the flag to determine whether to verify the TLS certs of the request
var DisableTLSVerification bool

// CreateRequest Creates the http request and applies any authentication
func CreateRequest(method, url string, body []byte, headers map[string]string, auth config.AuthConfig) (*http.Request, error) {
	var reader io.Reader

	// Don't build the reader if using a GET/HEAD request
	if method != "GET" && method != "HEAD" && body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		return nil, fmt.Errorf("could not build request: %s", err)
	}

	// Apply basic auth if configured
	if auth.UseBasicAuth() {
		authStr := base64.StdEncoding.EncodeToString([]byte(auth.Basic.Username + ":" + auth.Basic.Password))
		req.Header.Set("Authorization", "Basic "+authStr)
	}

	// Apply headers
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

// CreateClient Creates the HTTP client for the requests
func CreateClient(followRedirects bool, timeout int, clientCert config.AuthConfig) (*http.Client, error) {
	transport, err := createHTTPTransport(clientCert)
	if err != nil {
		return nil, err
	}

	client := buildHTTPClient(transport, followRedirects, timeout)
	return client, nil
}

func createHTTPTransport(clientCert config.AuthConfig) (*http.Transport, error) {
	// Create the default transport
	transport := &http.Transport{}

	// Build TLS configuration
	tlsConfig, err := buildTLSConfig(clientCert)
	if err != nil {
		return nil, err
	}

	transport.TLSClientConfig = tlsConfig

	return transport, nil
}

func buildTLSConfig(clientCert config.AuthConfig) (*tls.Config, error) {
	tlsConfig := &tls.Config{}

	// Skip TLS verification if disabled
	if DisableTLSVerification {
		tlsConfig.InsecureSkipVerify = true
	}

	// Configure CA certificate if provided
	if err := configureCACertificate(tlsConfig, clientCert); err != nil {
		return nil, err
	}

	// Configure client certificate if provided
	if err := configureClientCertificate(tlsConfig, clientCert); err != nil {
		return nil, err
	}

	return tlsConfig, nil
}

func configureCACertificate(tlsConfig *tls.Config, clientCert config.AuthConfig) error {
	ca := clientCert.Certificate.CA
	if ca == "" {
		return nil
	}

	caCertData, err := loadCertificateData(ca)
	if err != nil {
		return fmt.Errorf("failed to load CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCertData)
	tlsConfig.RootCAs = caCertPool

	return nil
}

func configureClientCertificate(tlsConfig *tls.Config, clientCert config.AuthConfig) error {
	cert := clientCert.Certificate.Cert
	key := clientCert.Certificate.Key
	if cert == "" || key == "" {
		return nil
	}

	certPair, err := loadClientCertificatePair(cert, key)
	if err != nil {
		return fmt.Errorf("failed to load client certificate: %w", err)
	}

	tlsConfig.Certificates = []tls.Certificate{certPair}
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
func buildHTTPClient(transport *http.Transport, followRedirects bool, timeout int) *http.Client {
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

// AuthConfigBuilder helps build AuthConfig configurations
type AuthConfigBuilder struct {
	auth config.AuthConfig
}

// NewAuthConfigBuilder creates a new builder with base configuration
func NewAuthConfigBuilder(baseConfig config.AuthConfig) *AuthConfigBuilder {
	return &AuthConfigBuilder{auth: baseConfig}
}

// WithCert sets the client certificate file
func (b *AuthConfigBuilder) WithCert(cert string) *AuthConfigBuilder {
	b.auth.Certificate.Cert = cert
	return b
}

// WithKey sets the client key file
func (b *AuthConfigBuilder) WithKey(key string) *AuthConfigBuilder {
	b.auth.Certificate.Key = key
	return b
}

// WithCA sets the CA certificate file
func (b *AuthConfigBuilder) WithCA(ca string) *AuthConfigBuilder {
	b.auth.Certificate.CA = ca
	return b
}

// WithBasicAuth sets basic authentication credentials
func (b *AuthConfigBuilder) WithBasicAuth(username, password string) *AuthConfigBuilder {
	b.auth.Basic.Username = username
	b.auth.Basic.Password = password
	return b
}

// Build returns the final AuthConfig configuration
func (b *AuthConfigBuilder) Build() config.AuthConfig {
	return b.auth
}

// BuildAuthConfig creates an AuthConfig object from individual parameters
func BuildAuthConfig(clientCert, clientCertKey, clientCA, basicAuthUser, basicAuthPass string, baseConfig config.AuthConfig) config.AuthConfig {
	// Trim whitespace from all inputs
	clientCert = strings.TrimSpace(clientCert)
	clientCertKey = strings.TrimSpace(clientCertKey)
	clientCA = strings.TrimSpace(clientCA)
	basicAuthUser = strings.TrimSpace(basicAuthUser)
	basicAuthPass = strings.TrimSpace(basicAuthPass)

	builder := NewAuthConfigBuilder(baseConfig)

	// Only set non-empty values
	if clientCert != "" {
		builder.WithCert(clientCert)
	}
	if clientCertKey != "" {
		builder.WithKey(clientCertKey)
	}
	if clientCA != "" {
		builder.WithCA(clientCA)
	}
	if basicAuthUser != "" || basicAuthPass != "" {
		builder.WithBasicAuth(basicAuthUser, basicAuthPass)
	}

	return builder.Build()
}
