package ios_test

import (
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
)

func TestMuxResponse(t *testing.T) {
	testCases := map[string]struct {
		muxResponse ios.MuxResponse
		successful  bool
	}{
		"successful response":   {ios.MuxResponse{MessageType: "random", Number: 0}, true},
		"unsuccessful response": {ios.MuxResponse{MessageType: "random", Number: 1}, false},
	}

	for _, tc := range testCases {
		bytes := []byte(ios.ToPlist(tc.muxResponse))
		actual := ios.MuxResponsefromBytes(bytes)
		assert.Equal(t, tc.muxResponse, actual)
		assert.Equal(t, tc.successful, actual.IsSuccessFull())
	}
}
