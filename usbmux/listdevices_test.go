package usbmux_test

/*
import (
	"github.com/danielpaulus/go-ios/usbmux"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringConversion(t *testing.T) {
	entryOne := usbmux.DeviceEntry{DeviceID: 5, MessageType: "", Properties: usbmux.DeviceProperties{SerialNumber: "udid0"}}
	entryTwo := usbmux.DeviceEntry{DeviceID: 5, MessageType: "", Properties: usbmux.DeviceProperties{SerialNumber: "udid1"}}

	testCases := map[string]struct {
		devices        usbmux.DeviceList
		expectedOutput string
	}{
		"zero entries":          {usbmux.DeviceList{DeviceList: make([]usbmux.DeviceEntry, 0)}, ""},
		"one entry":             {usbmux.DeviceList{DeviceList: []usbmux.DeviceEntry{entryOne}}, "udid0\n"},
		"more than one entries": {usbmux.DeviceList{DeviceList: []usbmux.DeviceEntry{entryOne, entryTwo}}, "udid0\nudid1\n"},
	}

	for _, tc := range testCases {
		actual := tc.devices.String()
		assert.Equal(t, tc.expectedOutput, actual)
	}

}

func TestListDevicesCommand(t *testing.T) {
	generified := func() interface{} { return usbmux.ListDevices() }
	entryOne := usbmux.DeviceEntry{DeviceID: 5, MessageType: "", Properties: usbmux.DeviceProperties{SerialNumber: "udid0"}}
	list := usbmux.DeviceList{DeviceList: []usbmux.DeviceEntry{entryOne}}
	receivedList := GenericMockUsbmuxdIntegrationTest(t, generified, usbmux.NewReadDevices(), list).(usbmux.DeviceList)
	assert.Equal(t, entryOne.Properties.SerialNumber, receivedList.DeviceList[0].Properties.SerialNumber)
}
*/
