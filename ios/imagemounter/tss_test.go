package imagemounter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

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
