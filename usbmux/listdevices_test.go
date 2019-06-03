package usbmux_test

import (
	"testing"
	"usbmuxd/usbmux"

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
	//setup dummy server
	path, cleanup := CreateSocketFilePath("socket")
	defer cleanup()
	serverReceiver := make(chan []byte)
	serverSender := make(chan []byte)
	serverCleanup := StartMuxServer(path, serverReceiver, serverSender)
	defer serverCleanup()

	usbmux.UsbmuxdSocket = path

	returnValue := make(chan usbmux.DeviceList)
	go func() {
		list := usbmux.ListDevices()
		returnValue <- list
	}()
	serverHasReceived := <-serverReceiver

	readDevicesPlist := usbmux.ToPlist(usbmux.NewReadDevices())

	assert.Equal(t, readDevicesPlist, string(serverHasReceived))

	entryOne := usbmux.DeviceEntry{DeviceID: 5, MessageType: "", Properties: usbmux.DeviceProperties{SerialNumber: "udid0"}}
	muxCodec := usbmux.MuxConnection{}
	list := usbmux.DeviceList{DeviceList: []usbmux.DeviceEntry{entryOne}}
	bytes, err := muxCodec.Encode(list)
	if assert.NoError(t, err) {
		serverSender <- bytes
		receivedList := <-returnValue
		assert.Equal(t, entryOne.Properties.SerialNumber, receivedList.DeviceList[0].Properties.SerialNumber)
	}

	/*//check that deviceConnection correctly passes received messages through
	//the active decoder
	serverSender <- message
	decoderShouldDecode := <-dummyCodec.received
	assert.ElementsMatch(t, message, decoderShouldDecode)
	deviceConn.Close()*/
}
