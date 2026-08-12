package imagemounter

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTssClientVerifiesTLS guards against re-introducing InsecureSkipVerify:
// getSignature reuses the client from newTssClient to POST to gs.apple.com,
// which serves a valid public certificate, so TLS verification must stay on.
func TestTssClientVerifiesTLS(t *testing.T) {
	c := newTssClient()
	transport, ok := c.h.Transport.(*http.Transport)
	require.True(t, ok, "expected *http.Transport")
	if transport.TLSClientConfig != nil {
		assert.False(t, transport.TLSClientConfig.InsecureSkipVerify,
			"TLS certificate verification must not be disabled")
	}
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
