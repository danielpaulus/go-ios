package ios_test

import (
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"

	"github.com/stretchr/testify/assert"
)

func TestStringConversion(t *testing.T) {
	entryOne := ios.DeviceEntry{DeviceID: 5, MessageType: "", Properties: ios.DeviceProperties{SerialNumber: "udid0"}}
	entryTwo := ios.DeviceEntry{DeviceID: 5, MessageType: "", Properties: ios.DeviceProperties{SerialNumber: "udid1"}}

	testCases := map[string]struct {
		devices        ios.DeviceList
		expectedOutput string
	}{
		"zero entries":          {ios.DeviceList{DeviceList: make([]ios.DeviceEntry, 0)}, ""},
		"one entry":             {ios.DeviceList{DeviceList: []ios.DeviceEntry{entryOne}}, "udid0\n"},
		"more than one entries": {ios.DeviceList{DeviceList: []ios.DeviceEntry{entryOne, entryTwo}}, "udid0\nudid1\n"},
	}

	for _, tc := range testCases {
		actual := tc.devices.String()
		assert.Equal(t, tc.expectedOutput, actual)
	}
}
