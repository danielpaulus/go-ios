package imagemounter

import (
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTssClientVerifiesTLS guards against re-introducing InsecureSkipVerify.
func TestTssClientVerifiesTLS(t *testing.T) {
	c := newTssClient()
	transport, ok := c.h.Transport.(*http.Transport)
	require.True(t, ok, "expected *http.Transport")
	require.NotNil(t, transport.TLSClientConfig)
	assert.False(t, transport.TLSClientConfig.InsecureSkipVerify,
		"TLS certificate verification must not be disabled")
}

func TestTssClientTrustsEmbeddedAppleRootCA(t *testing.T) {
	c := newTssClient()
	transport, ok := c.h.Transport.(*http.Transport)
	require.True(t, ok, "expected *http.Transport")
	require.NotNil(t, transport.TLSClientConfig.RootCAs)

	cert := parseAppleRootCA(t)
	assert.True(t, poolContains(transport.TLSClientConfig.RootCAs, cert))
}

func TestAppleRootCAPEMIsValid(t *testing.T) {
	cert := parseAppleRootCA(t)
	assert.Equal(t, "Apple Root CA", cert.Subject.CommonName)
	assert.Equal(t, cert.Subject.String(), cert.Issuer.String(), "expected a self-signed root, not an intermediate")
	assert.True(t, cert.IsCA, "expected the CA basic constraint to be set")
	assert.True(t, cert.NotAfter.After(time.Now()), "embedded Apple Root CA has expired")
}

func parseAppleRootCA(t *testing.T) *x509.Certificate {
	t.Helper()
	block, _ := pem.Decode(appleRootCAPEM)
	require.NotNil(t, block, "appleRootCAPEM is not a valid PEM block")
	cert, err := x509.ParseCertificate(block.Bytes)
	require.NoError(t, err)
	return cert
}

// x509.CertPool has no public membership check, so compare raw subjects.
func poolContains(pool *x509.CertPool, cert *x509.Certificate) bool {
	for _, subject := range pool.Subjects() { //nolint:staticcheck
		if string(subject) == string(cert.RawSubject) {
			return true
		}
	}
	return false
}

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  response
	}{
		{
			name:  "success without request message",
			input: "STATUS=0&MESSAGE=SUCCESS",
			want: response{
				status:  0,
				message: "SUCCESS",
			},
		},
		{
			name:  "response with multiword status",
			input: "STATUS=69&MESSAGE=This device isn't eligible for the requested build.",
			want: response{
				status:  69,
				message: "This device isn't eligible for the requested build.",
			},
		},
		{
			name:  "response with request string",
			input: "STATUS=0&MESSAGE=SUCCESS&REQUEST_STRING=<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>",
			want: response{
				status:        0,
				message:       "SUCCESS",
				requestString: "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			r, err := parseResponse(reader)
			require.NoError(t, err)
			assert.Equal(t, tt.want, r)
		})
	}
}

// TODO: It looks like `REQUEST_STRING` always comes last, but if that's not the case we are not sure what to do
// as it could also contain the '&' separator character
func TestParseResponseRequiresRequestStringLast(t *testing.T) {
	_, err := parseResponse(strings.NewReader("REQUEST_STRING=abc&STATUS=0"))
	assert.Error(t, err)
}
