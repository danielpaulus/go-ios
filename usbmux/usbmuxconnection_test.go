package usbmux_test

import (
	"testing"
	"usbmuxd/usbmux"

	mock "github.com/stretchr/testify/mock"
)

func TestCodec(t *testing.T) {
	//mc := new(manualClientMock)
	//	mc.On("GetSessionInfo", mock.Anything, mock.Anything).Return(manualclient.Session{}, errors.New("Fail"))

	deviceConnMock := new(DeviceConnectionMock)
	deviceConnMock.On("Connect", mock.Anything)
	muxConn := usbmux.NewUsbMuxConnectionWithDeviceConnection(deviceConnMock)
	muxConn.Close()
}

type DeviceConnectionMock struct {
	mock.Mock
}

func (mock *DeviceConnectionMock) Connect(activeCodec usbmux.Codec) {

}

func (mock *DeviceConnectionMock) ConnectToSocketAddress(activeCodec usbmux.Codec, socketAddress string) {
}
func (mock *DeviceConnectionMock) Close() {}
func (mock *DeviceConnectionMock) SendForProtocolUpgrade(muxConnection *usbmux.MuxConnection, message interface{}, newCodec usbmux.Codec) []byte {
	return nil
}
func (mock *DeviceConnectionMock) Send(message interface{}) {}
