package usbmux_test

import (
	"testing"
	"usbmuxd/usbmux"

	"github.com/stretchr/testify/assert"
)

func TestMuxResponse(t *testing.T) {
	testCases := map[string]struct {
		muxResponse usbmux.MuxResponse
		successful  bool
	}{
		"successful response":   {usbmux.MuxResponse{MessageType: "random", Number: 0}, true},
		"unsuccessful response": {usbmux.MuxResponse{MessageType: "random", Number: 1}, false},
	}

	for _, tc := range testCases {
		bytes := []byte(usbmux.ToPlist(tc.muxResponse))
		actual := usbmux.MuxResponsefromBytes(bytes)
		assert.Equal(t, tc.muxResponse, actual)
		assert.Equal(t, tc.successful, actual.IsSuccessFull())
	}

}
