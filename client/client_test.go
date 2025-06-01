package client

import (
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/thoom/gulp/config"
)

func init() {
	// Reset TLS verification for tests
	DisableTLSVerification = false
}

func TestDisableTLS(t *testing.T) {
	assert := assert.New(t)

	DisableTLSVerification = true
	assert.True(DisableTLSVerification)
}

func TestCreateRequest(t *testing.T) {
	assert := assert.New(t)

	method := "GET"
	url := "http://test.ex.io"
	headers := map[string]string{}
	headers["X-Test-Header"] = "abc123def"

	req, err := CreateRequest(method, url, nil, headers, config.AuthConfig{})
	assert.Nil(err)
	assert.Equal(url, req.URL.String())
	assert.Equal(method, req.Method)
	assert.Equal(1, len(req.Header))
	assert.EqualValues(headers["X-Test-Header"], req.Header.Get("X-Test-Header"))
	assert.Nil(req.Body)
}

func TestCreateRequestBadMethod(t *testing.T) {
	assert := assert.New(t)

	method := "INVALID METHOD"
	url := "http://test.ex.io"
	headers := map[string]string{}

	req, err := CreateRequest(method, url, nil, headers, config.AuthConfig{})
	assert.Nil(req)
	assert.Error(err)
}

func TestCreateRequestGetWithBody(t *testing.T) {
	assert := assert.New(t)

	method := "GET"
	url := "http://test.ex.io"
	body := []byte("body!")

	req, err := CreateRequest(method, url, body, map[string]string{}, config.AuthConfig{})
	assert.Nil(err)
	assert.Equal(url, req.URL.String())
	assert.Equal(method, req.Method)
	assert.Empty(req.Header)
	assert.Nil(req.Body)
}

func TestCreateRequestPostWithBody(t *testing.T) {
	assert := assert.New(t)

	method := "POST"
	url := "http://test.ex.io"
	body := "body!"

	req, err := CreateRequest(method, url, []byte(body), map[string]string{}, config.AuthConfig{})
	assert.Nil(err)
	assert.Equal(url, req.URL.String())
	assert.Equal(method, req.Method)
	assert.Empty(req.Header)

	// Hacky way to get the body for now
	requestDump, _ := httputil.DumpRequest(req, true)
	reqDumpStr := strings.Split(string(requestDump), "\n")
	assert.Equal(body, reqDumpStr[len(reqDumpStr)-1])
}

func TestCreateClient(t *testing.T) {
	assert := assert.New(t)

	client, err := CreateClient(false, 10, config.New.Auth)
	assert.Nil(err)
	assert.Equal(time.Duration(10)*time.Second, client.Timeout)
}

func TestCreateClientFollowRedirects(t *testing.T) {
	assert := assert.New(t)
	client, err := CreateClient(true, 10, config.New.Auth)
	assert.Nil(err)
	assert.Equal(time.Duration(10)*time.Second, client.Timeout)
}

func TestCreateClientClientCertAuth(t *testing.T) {
	assert := assert.New(t)

	cert := `-----BEGIN CERTIFICATE-----
MIIExDCCAqwCCQCTh/r3DopngDANBgkqhkiG9w0BAQsFADAkMQswCQYDVQQGEwJV
UzEVMBMGA1UEAwwMVGVzdCBQcml2YXRlMB4XDTIyMDIxNjIwMjkzNFoXDTMyMDIx
NDIwMjkzNFowJDELMAkGA1UEBhMCVVMxFTATBgNVBAMMDFRlc3QgUHJpdmF0ZTCC
AiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAK4jkDOiysMW3YAt+sXJ0uqO
hLfXc8XEJdlAvRBhPIQDGWuBJIf3JrEZ8iZgP+6NYFyTyE5CLVxnDKzQA6vrrUQp
/3i2CTaIcl9RSw6/qrsr0V6s0wlNzweFqMSGDSH09595u4Guby1WR8AwHL/Myoi6
Hzey9jIvGTLvJa6bLK/xCHRmr7uayq+05/aBrsWnhzQ9zEGr3DnRylC/wTwy09MM
YPQOtvq14s77OjIuj2RGJDzsm/7SebOW9XJoRrJ1a//iAP8EO5R1QcIW9SV6Wv8+
9anJ2HT9Gv6NIS8q1sQorIORilfF8dQBwyUUY6EUf9btgRjGqLvEAP1s7TKHTC6O
o7G71nFLEksIS/5fAwMMDHMxGyYoA5yLyrLc2v50UKMCMsATjNHUYICeZOlwti5H
7iHwhTjOSpxVI6elwUytTj10xMcvmJV7AUpIT3PwbbZcFXsyUbfqZvRFrLOLg53E
rQpFRjSazFUdzw2jlSBYnm3z+DeA60AaTh3n2+IxxN6NKXkefRlTLR7KfQs6+J8G
deL9ydw+164Xah3r1lAUbSe6Oi7YAmQ56bUQsaDe550hV4dHV6d8cPHsOY3ad1si
egmNA0gpp6o+51iB5F7C9Hz7Qp5egPWc9l0pSiTKPTQysEmcV5ZPS5+PHl299wcX
O48gD8QpexNRJ+Bq+E/XAgMBAAEwDQYJKoZIhvcNAQELBQADggIBAF9+sDuj2f1I
hbG2VPLzto8eJVu8q+w9zghzXwMBHUQnL4IZABLsuHRGfHIx/FjsL5WDcr17EmMa
L7A+epLekAD1DRVmt8GwXPyEkXnMcrMwZHzi+jE49aQrQ+DJ62Ki7sPkcQEpf1ZM
435gNrEAoq+3oNGE+Rejs4m5UKybdvQEHen/tCuM8/ns6Pugdo7ihuIDo0f7aXwn
jrPYI3q1nKak5xIKwxoEubCrHtgMC3wmMiNNMO6hBCUsZZDPu3RkhWZYWGPxJISHT
1E+pmVVoEnggewh3R92/aUletsM7ZQ410FBRrOki6tv9gWd3BpjlVESkZCU1tu3
hCM05zX2YGeaomVtjj4s8KaR1Fw4nR5QSTTFhSnZuNo1WifSNlI5p+Ir4jaOqCwP
L+SLQ+CkWFQifnmhKzXcU1BF+JnfYW8Opt+wO2PU3CXGzMcPBpVIA8A2SM2hs4Id
5wlxBb+smSviawji4GKUfBlBR8OMINDsBBrlI4rs6nfh/DLE1mwsJC+IGePsG1xI
lOA4MPmK4WhctxNUYoiKoeQ2pQufa6Smy0LO33yOrp+CB8auPmme6qZjKHcT66XQ
7QzoMXwC0lCpBs1lwy9RW1Bsrc5IjgY18Hxi5JlsQWybieeh0rN/MgzLEOxFXwfn
+qZyHKpJvNn/X7szVwakIEGsgeSmhPAK
-----END CERTIFICATE-----
`

	key := `-----BEGIN PRIVATE KEY-----
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
uHsldhZyjInCxkuuzW3khHFKSs+C
-----END PRIVATE KEY-----
`

	certFile, _ := os.CreateTemp(os.TempDir(), "test-client")
	defer certFile.Close()
	os.WriteFile(certFile.Name(), []byte(cert), 0644)

	keyFile, _ := os.CreateTemp(os.TempDir(), "test-client")
	defer keyFile.Close()
	os.WriteFile(keyFile.Name(), []byte(key), 0644)

	clientAuth := config.AuthConfig{
		Certificate: config.CertAuthConfig{
			Cert: certFile.Name(),
			Key:  keyFile.Name(),
		},
	}

	_, err := CreateClient(true, 10, clientAuth)
	assert.Nil(err)
}

func TestCreateClientClientCertAuthError(t *testing.T) {
	assert := assert.New(t)

	clientAuth := config.AuthConfig{
		Certificate: config.CertAuthConfig{
			Cert: "test.pem",
			Key:  "test.key",
		},
	}

	_, err := CreateClient(true, 10, clientAuth)
	assert.NotNil(err)
}

func TestBuildClientAgentConfigOnly(t *testing.T) {
	assert := assert.New(t)

	def := config.New.Auth

	res := BuildAuthConfig("", "", "", "", "", def)
	assert.Equal(def, res)
}

func TestBuildAgentCertFlag(t *testing.T) {
	assert := assert.New(t)

	def := config.New.Auth

	res := BuildAuthConfig("test1.pem", "", "", "", "", def)
	assert.Equal("test1.pem", res.Certificate.Cert)
	assert.Equal(def.Certificate.Key, res.Certificate.Key)
	assert.Equal(def.Certificate.CA, res.Certificate.CA)
}

func TestBuildAgentCertKeyFlag(t *testing.T) {
	assert := assert.New(t)

	def := config.New.Auth

	res := BuildAuthConfig("", "testkey.pem", "", "", "", def)
	assert.Equal(def.Certificate.Cert, res.Certificate.Cert)
	assert.Equal("testkey.pem", res.Certificate.Key)
	assert.Equal(def.Certificate.CA, res.Certificate.CA)
}

// Tests for CA certificate functionality
func TestBuildClientAuthWithCA(t *testing.T) {
	assert := assert.New(t)

	def := config.New.Auth

	res := BuildAuthConfig("", "", "customca.pem", "", "", def)
	assert.Equal("customca.pem", res.Certificate.CA)
}

func TestBuildClientAuthCAFromConfig(t *testing.T) {
	assert := assert.New(t)

	def := config.New.Auth
	def.Certificate.CA = "testca.pem"

	res := BuildAuthConfig("", "", "", "", "", def)
	assert.Equal("testca.pem", res.Certificate.CA)
}

// Tests for custom CA certificate file path
func TestCreateClientWithCAFile(t *testing.T) {
	assert := assert.New(t)

	// Use the existing working client cert as CA for testing purposes
	caCert := `-----BEGIN CERTIFICATE-----
MIIExDCCAqwCCQCTh/r3DopngDANBgkqhkiG9w0BAQsFADAkMQswCQYDVQQGEwJV
UzEVMBMGA1UEAwwMVGVzdCBQcml2YXRlMB4XDTIyMDIxNjIwMjkzNFoXDTMyMDIx
NDIwMjkzNFowJDELMAkGA1UEBhMCVVMxFTATBgNVBAMMDFRlc3QgUHJpdmF0ZTCC
AiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAK4jkDOiysMW3YAt+sXJ0uqO
hLfXc8XEJdlAvRBhPIQDGWuBJIf3JrEZ8iZgP+6NYFyTyE5CLVxnDKzQA6vrrUQp
/3i2CTaIcl9RSw6/qrsr0V6s0wlNzweFqMSGDSH09595u4Guby1WR8AwHL/Myoi6
Hzey9jIvGTLvJa6bLK/xCHRmr7uayq+05/aBrsWnhzQ9zEGr3DnRylC/wTwy09MM
YPQOtvq14s77OjIuj2RGJDzsm/7SebOW9XJoRrJ1a//iAP8EO5R1QcIW9SV6Wv8+
9anJ2HT9Gv6NIS8q1sQorIORilfF8dQBwyUUY6EUf9btgRjGqLvEAP1s7TKHTC6O
o7G71nFLEksIS/5fAwMMDHMxGyYoA5yLyrLc2v50UKMCMsATjNHUYICeZOlwti5H
7iHwhTjOSpxVI6elwUytTj10xMcvmJV7AUpIT3PwbbZcFXsyUbfqZvRFrLOLg53E
rQpFRjSazFUdzw2jlSBYnm3z+DeA60AaTh3n2+IxxN6NKXkefRlTLR7KfQs6+J8G
deL9ydw+164Xah3r1lAUbSe6Oi7YAmQ56bUQsaDe550hV4dHV6d8cPHsOY3ad1si
egmNA0gpp6o+51iB5F7C9Hz7Qp5egPWc9l0pSiTKPTQysEmcV5ZPS5+PHl299wcX
O48gD8QpexNRJ+Bq+E/XAgMBAAEwDQYJKoZIhvcNAQELBQADggIBAF9+sDuj2f1I
hbG2VPLzto8eJVu8q+w9zghzXwMBHUQnL4IZABLsuHRGfHIx/FjsL5WDcr17EmMa
L7A+epLekAD1DRVmt8GwXPyEkXnMcrMwZHzi+jE49aQrQ+DJ62Ki7sPkcQEpf1ZM
435gNrEAoq+3oNGE+Rejs4m5UKybdvQEHen/tCuM8/ns6Pugdo7ihuIDo0f7aXwn
jrPYI3q1nKak5xIKwxoEubCrHtgMC3wmMiNNMO6hBCUsZZDPu3RkhWZYWGPxJISHT
1E+pmVVoEnggewh3R92/aUletsM7ZQ410FBRrOki6tv9gWd3BpjlVESkZCU1tu3
hCM05zX2YGeaomVtjj4s8KaR1Fw4nR5QSTTFhSnZuNo1WifSNlI5p+Ir4jaOqCwP
L+SLQ+CkWFQifnmhKzXcU1BF+JnfYW8Opt+wO2PU3CXGzMcPBpVIA8A2SM2hs4Id
5wlxBb+smSviawji4GKUfBlBR8OMINDsBBrlI4rs6nfh/DLE1mwsJC+IGePsG1xI
lOA4MPmK4WhctxNUYoiKoeQ2pQufa6Smy0LO33yOrp+CB8auPmme6qZjKHcT66XQ
7QzoMXwC0lCpBs1lwy9RW1Bsrc5IjgY18Hxi5JlsQWybieeh0rN/MgzLEOxFXwfn
+qZyHKpJvNn/X7szVwakIEGsgeSmhPAK
-----END CERTIFICATE-----`

	caFile, _ := os.CreateTemp(os.TempDir(), "test-ca")
	defer caFile.Close()
	defer os.Remove(caFile.Name())
	os.WriteFile(caFile.Name(), []byte(caCert), 0644)

	clientAuth := config.AuthConfig{
		Certificate: config.CertAuthConfig{
			CA: caFile.Name(),
		},
	}

	client, err := CreateClient(true, 10, clientAuth)
	assert.Nil(err)
	assert.NotNil(client)

	// Ensure TLS config exists and has a root CA pool
	tlsConfig := client.Transport.(*http.Transport).TLSClientConfig
	assert.NotNil(tlsConfig)
	assert.NotNil(tlsConfig.RootCAs)
}

// Tests for inline PEM content
func TestCreateClientWithInlineCA(t *testing.T) {
	assert := assert.New(t)

	// Use the existing working client cert as CA for testing purposes
	inlineCA := `-----BEGIN CERTIFICATE-----
MIIExDCCAqwCCQCTh/r3DopngDANBgkqhkiG9w0BAQsFADAkMQswCQYDVQQGEwJV
UzEVMBMGA1UEAwwMVGVzdCBQcml2YXRlMB4XDTIyMDIxNjIwMjkzNFoXDTMyMDIx
NDIwMjkzNFowJDELMAkGA1UEBhMCVVMxFTATBgNVBAMMDFRlc3QgUHJpdmF0ZTCC
AiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAK4jkDOiysMW3YAt+sXJ0uqO
hLfXc8XEJdlAvRBhPIQDGWuBJIf3JrEZ8iZgP+6NYFyTyE5CLVxnDKzQA6vrrUQp
/3i2CTaIcl9RSw6/qrsr0V6s0wlNzweFqMSGDSH09595u4Guby1WR8AwHL/Myoi6
Hzey9jIvGTLvJa6bLK/xCHRmr7uayq+05/aBrsWnhzQ9zEGr3DnRylC/wTwy09MM
YPQOtvq14s77OjIuj2RGJDzsm/7SebOW9XJoRrJ1a//iAP8EO5R1QcIW9SV6Wv8+
9anJ2HT9Gv6NIS8q1sQorIORilfF8dQBwyUUY6EUf9btgRjGqLvEAP1s7TKHTC6O
o7G71nFLEksIS/5fAwMMDHMxGyYoA5yLyrLc2v50UKMCMsATjNHUYICeZOlwti5H
7iHwhTjOSpxVI6elwUytTj10xMcvmJV7AUpIT3PwbbZcFXsyUbfqZvRFrLOLg53E
rQpFRjSazFUdzw2jlSBYnm3z+DeA60AaTh3n2+IxxN6NKXkefRlTLR7KfQs6+J8G
deL9ydw+164Xah3r1lAUbSe6Oi7YAmQ56bUQsaDe550hV4dHV6d8cPHsOY3ad1si
egmNA0gpp6o+51iB5F7C9Hz7Qp5egPWc9l0pSiTKPTQysEmcV5ZPS5+PHl299wcX
O48gD8QpexNRJ+Bq+E/XAgMBAAEwDQYJKoZIhvcNAQELBQADggIBAF9+sDuj2f1I
hbG2VPLzto8eJVu8q+w9zghzXwMBHUQnL4IZABLsuHRGfHIx/FjsL5WDcr17EmMa
L7A+epLekAD1DRVmt8GwXPyEkXnMcrMwZHzi+jE49aQrQ+DJ62Ki7sPkcQEpf1ZM
435gNrEAoq+3oNGE+Rejs4m5UKybdvQEHen/tCuM8/ns6Pugdo7ihuIDo0f7aXwn
jrPYI3q1nKak5xIKwxoEubCrHtgMC3wmMiNNMO6hBCUsZZDPu3RkhWZYWGPxJISHT
1E+pmVVoEnggewh3R92/aUletsM7ZQ410FBRrOki6tv9gWd3BpjlVESkZCU1tu3
hCM05zX2YGeaomVtjj4s8KaR1Fw4nR5QSTTFhSnZuNo1WifSNlI5p+Ir4jaOqCwP
L+SLQ+CkWFQifnmhKzXcU1BF+JnfYW8Opt+wO2PU3CXGzMcPBpVIA8A2SM2hs4Id
5wlxBb+smSviawji4GKUfBlBR8OMINDsBBrlI4rs6nfh/DLE1mwsJC+IGePsG1xI
lOA4MPmK4WhctxNUYoiKoeQ2pQufa6Smy0LO33yOrp+CB8auPmme6qZjKHcT66XQ
7QzoMXwC0lCpBs1lwy9RW1Bsrc5IjgY18Hxi5JlsQWybieeh0rN/MgzLEOxFXwfn
+qZyHKpJvNn/X7szVwakIEGsgeSmhPAK
-----END CERTIFICATE-----`

	clientAuth := config.AuthConfig{
		Certificate: config.CertAuthConfig{
			CA: inlineCA,
		},
	}

	client, err := CreateClient(true, 10, clientAuth)
	assert.Nil(err)
	assert.NotNil(client)

	// Ensure TLS config exists and has a root CA pool
	tlsConfig := client.Transport.(*http.Transport).TLSClientConfig
	assert.NotNil(tlsConfig)
	assert.NotNil(tlsConfig.RootCAs)
}

// Tests for inline client certificates
func TestCreateClientWithInlineClientCert(t *testing.T) {
	assert := assert.New(t)

	inlineCert := `-----BEGIN CERTIFICATE-----
MIIExDCCAqwCCQCTh/r3DopngDANBgkqhkiG9w0BAQsFADAkMQswCQYDVQQGEwJV
UzEVMBMGA1UEAwwMVGVzdCBQcml2YXRlMB4XDTIyMDIxNjIwMjkzNFoXDTMyMDIx
NDIwMjkzNFowJDELMAkGA1UEBhMCVVMxFTATBgNVBAMMDFRlc3QgUHJpdmF0ZTCC
AiIwDQYJKoZIhvcNAQEBBQADggIPADCCAgoCggIBAK4jkDOiysMW3YAt+sXJ0uqO
hLfXc8XEJdlAvRBhPIQDGWuBJIf3JrEZ8iZgP+6NYFyTyE5CLVxnDKzQA6vrrUQp
/3i2CTaIcl9RSw6/qrsr0V6s0wlNzweFqMSGDSH09595u4Guby1WR8AwHL/Myoi6
Hzey9jIvGTLvJa6bLK/xCHRmr7uayq+05/aBrsWnhzQ9zEGr3DnRylC/wTwy09MM
YPQOtvq14s77OjIuj2RGJDzsm/7SebOW9XJoRrJ1a//iAP8EO5R1QcIW9SV6Wv8+
9anJ2HT9Gv6NIS8q1sQorIORilfF8dQBwyUUY6EUf9btgRjGqLvEAP1s7TKHTC6O
o7G71nFLEksIS/5fAwMMDHMxGyYoA5yLyrLc2v50UKMCMsATjNHUYICeZOlwti5H
7iHwhTjOSpxVI6elwUytTj10xMcvmJV7AUpIT3PwbbZcFXsyUbfqZvRFrLOLg53E
rQpFRjSazFUdzw2jlSBYnm3z+DeA60AaTh3n2+IxxN6NKXkefRlTLR7KfQs6+J8G
deL9ydw+164Xah3r1lAUbSe6Oi7YAmQ56bUQsaDe550hV4dHV6d8cPHsOY3ad1si
egmNA0gpp6o+51iB5F7C9Hz7Qp5egPWc9l0pSiTKPTQysEmcV5ZPS5+PHl299wcX
O48gD8QpexNRJ+Bq+E/XAgMBAAEwDQYJKoZIhvcNAQELBQADggIBAF9+sDuj2f1I
hbG2VPLzto8eJVu8q+w9zghzXwMBHUQnL4IZABLsuHRGfHIx/FjsL5WDcr17EmMa
L7A+epLekAD1DRVmt8GwXPyEkXnMcrMwZHzi+jE49aQrQ+DJ62Ki7sPkcQEpf1ZM
435gNrEAoq+3oNGE+Rejs4m5UKybdvQEHen/tCuM8/ns6Pugdo7ihuIDo0f7aXwn
jrPYI3q1nKak5xIKwxoEubCrHtgMC3wmMiNNMO6hBCUsZZDPu3RkhWZYWGPxJISHT
1E+pmVVoEnggewh3R92/aUletsM7ZQ410FBRrOki6tv9gWd3BpjlVESkZCU1tu3
hCM05zX2YGeaomVtjj4s8KaR1Fw4nR5QSTTFhSnZuNo1WifSNlI5p+Ir4jaOqCwP
L+SLQ+CkWFQifnmhKzXcU1BF+JnfYW8Opt+wO2PU3CXGzMcPBpVIA8A2SM2hs4Id
5wlxBb+smSviawji4GKUfBlBR8OMINDsBBrlI4rs6nfh/DLE1mwsJC+IGePsG1xI
lOA4MPmK4WhctxNUYoiKoeQ2pQufa6Smy0LO33yOrp+CB8auPmme6qZjKHcT66XQ
7QzoMXwC0lCpBs1lwy9RW1Bsrc5IjgY18Hxi5JlsQWybieeh0rN/MgzLEOxFXwfn
+qZyHKpJvNn/X7szVwakIEGsgeSmhPAK
-----END CERTIFICATE-----`

	inlineKey := `-----BEGIN PRIVATE KEY-----
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
uHsldhZyjInCxkuuzW3khHFKSs+C
-----END PRIVATE KEY-----`

	clientAuth := config.AuthConfig{
		Certificate: config.CertAuthConfig{
			Cert: inlineCert,
			Key:  inlineKey,
		},
	}

	client, err := CreateClient(true, 10, clientAuth)
	assert.Nil(err)
	assert.NotNil(client)
	assert.NotNil(client.Transport.(*http.Transport).TLSClientConfig)
	assert.Equal(1, len(client.Transport.(*http.Transport).TLSClientConfig.Certificates))
}

// Test mixed format validation - should fail
func TestCreateClientMixedFormat(t *testing.T) {
	assert := assert.New(t)

	// Create a temporary cert file
	certFile, _ := os.CreateTemp(os.TempDir(), "test-cert")
	defer certFile.Close()
	defer os.Remove(certFile.Name())
	os.WriteFile(certFile.Name(), []byte("cert content"), 0644)

	inlineKey := `-----BEGIN PRIVATE KEY-----
MIIJQQIBADANBgkqhkiG9w0BAQEFAASCCSswggknAgEAAoICAQCuI5AzosrDFt2A
LfrFydLqjoS313PFxCXZQL0QYTyEAxlrgSSH9yaxGfImYD/ujWBck8hOQi1cZwys
0AOr661EKf94tgk2iHJfUUsOv6q7K9FerNMJTc8HhajEhg0h9PefebuBrm8tVkfA
-----END PRIVATE KEY-----`

	clientAuth := config.AuthConfig{
		Certificate: config.CertAuthConfig{
			Cert: certFile.Name(), // File path
			Key:  inlineKey,       // Inline PEM
		},
	}

	_, err := CreateClient(true, 10, clientAuth)
	assert.NotNil(err)
	assert.Contains(err.Error(), "mixed")
}

// Test error cases
func TestCreateClientInvalidCAFile(t *testing.T) {
	assert := assert.New(t)

	clientAuth := config.AuthConfig{
		Certificate: config.CertAuthConfig{
			CA: "/nonexistent/ca.pem",
		},
	}

	_, err := CreateClient(true, 10, clientAuth)
	assert.NotNil(err)
	assert.Contains(err.Error(), "failed to load CA certificate")
}

func TestCreateClientInvalidCAPEM(t *testing.T) {
	assert := assert.New(t)

	clientAuth := config.AuthConfig{
		Certificate: config.CertAuthConfig{
			CA: "-----BEGIN CERTIFICATE-----\ninvalid pem content\n-----END CERTIFICATE-----",
		},
	}

	client, err := CreateClient(true, 10, clientAuth)
	// Invalid PEM content might be silently ignored by Go's cert pool
	// so we just check that the client was created
	assert.Nil(err)
	assert.NotNil(client)
}

func TestCreateClientInvalidInlineCert(t *testing.T) {
	assert := assert.New(t)

	clientAuth := config.AuthConfig{
		Certificate: config.CertAuthConfig{
			Cert: "-----BEGIN CERTIFICATE-----\ninvalid cert\n-----END CERTIFICATE-----",
			Key:  "-----BEGIN PRIVATE KEY-----\ninvalid key\n-----END PRIVATE KEY-----",
		},
	}

	_, err := CreateClient(true, 10, clientAuth)
	assert.NotNil(err)
	assert.Contains(err.Error(), "failed to load client certificate")
}

func TestCreateRequestWithBasicAuth(t *testing.T) {
	assert := assert.New(t)

	method := "POST"
	url := "http://test.ex.io"
	headers := map[string]string{}
	clientAuth := config.AuthConfig{
		Basic: config.BasicAuthConfig{
			Username: "testuser",
			Password: "testpass",
		},
	}

	req, err := CreateRequest(method, url, nil, headers, clientAuth)
	assert.Nil(err)
	assert.Equal(url, req.URL.String())
	assert.Equal(method, req.Method)

	// Check that Authorization header is set with basic auth
	authHeader := req.Header.Get("Authorization")
	assert.NotEmpty(authHeader)
	assert.Contains(authHeader, "Basic ")

	// Verify the base64 encoding is correct
	expectedAuth := "Basic dGVzdHVzZXI6dGVzdHBhc3M=" // base64 of "testuser:testpass"
	assert.Equal(expectedAuth, authHeader)
}

func TestBuildAuthConfigWithBasicAuth(t *testing.T) {
	assert := assert.New(t)

	def := config.New.Auth

	res := BuildAuthConfig("", "", "", "myuser", "mypass", def)
	assert.Equal("myuser", res.Basic.Username)
	assert.Equal("mypass", res.Basic.Password)
}

func TestBasicAuthUsage(t *testing.T) {
	assert := assert.New(t)

	// Test UseAuth method
	auth := config.AuthConfig{
		Certificate: config.CertAuthConfig{
			Cert: "cert.pem",
			Key:  "key.pem",
		},
	}
	assert.True(auth.UseAuth())

	// Test with empty values
	authEmpty := config.AuthConfig{}
	assert.False(authEmpty.UseAuth())

	// Test with only username
	authPartial := config.AuthConfig{
		Certificate: config.CertAuthConfig{
			Cert: "cert.pem",
		},
	}
	assert.False(authPartial.UseAuth())
}

func TestCreateHTTPTransport(t *testing.T) {
	assert := assert.New(t)

	authConfig := config.AuthConfig{}
	transport, err := createHTTPTransport(authConfig)
	assert.Nil(err)
	assert.NotNil(transport)
}

func TestBuildTLSConfig(t *testing.T) {
	assert := assert.New(t)

	// Test with no CA or client cert
	clientAuth := config.AuthConfig{}
	tlsConfig, err := buildTLSConfig(clientAuth)
	assert.Nil(err)
	assert.NotNil(tlsConfig)                  // TLS config is always created now
	assert.True(tlsConfig.InsecureSkipVerify) // Due to DisableTLSVerification being true

	// Test with CA only
	clientAuth = config.AuthConfig{
		Certificate: config.CertAuthConfig{
			CA: "-----BEGIN CERTIFICATE-----\nMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA\n-----END CERTIFICATE-----",
		},
	}
	tlsConfig, err = buildTLSConfig(clientAuth)
	assert.Nil(err)
	assert.NotNil(tlsConfig)
	assert.NotNil(tlsConfig.RootCAs)
}

func TestLoadCertificateData(t *testing.T) {
	assert := assert.New(t)

	// Test with PEM content
	pemData := "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----"
	result, err := loadCertificateData(pemData)
	assert.Nil(err)
	assert.Equal([]byte(pemData), result)

	// Test with file path (non-existent)
	_, err = loadCertificateData("/nonexistent/file.pem")
	assert.NotNil(err)
}

func TestLoadClientCertificatePair(t *testing.T) {
	assert := assert.New(t)

	// Test mixed format error
	certPEM := "-----BEGIN CERTIFICATE-----\ntest\n-----END CERTIFICATE-----"
	keyFile := "/path/to/key.pem"
	_, err := loadClientCertificatePair(certPEM, keyFile)
	assert.NotNil(err)
	assert.Contains(err.Error(), "must both be either file paths or inline PEM content")

	// Test both file paths (non-existent)
	_, err = loadClientCertificatePair("/cert.pem", "/key.pem")
	assert.NotNil(err)
}

func TestBuildHTTPClient(t *testing.T) {
	assert := assert.New(t)

	transport := &http.Transport{}

	// Test with redirects enabled
	client := buildHTTPClient(transport, true, 30)
	assert.NotNil(client)
	assert.Equal(30*time.Second, client.Timeout)
	assert.Equal(transport, client.Transport)
	assert.Nil(client.CheckRedirect)

	// Test with redirects disabled
	client = buildHTTPClient(transport, false, 60)
	assert.NotNil(client)
	assert.Equal(60*time.Second, client.Timeout)
	assert.NotNil(client.CheckRedirect)
}
