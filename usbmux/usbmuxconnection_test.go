package usbmux_test

import (
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/danielpaulus/go-ios/usbmux"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	mock "github.com/stretchr/testify/mock"
)

func TestCodec(t *testing.T) {
	//mc := new(manualClientMock)
	//	mc.On("GetSessionInfo", mock.Anything, mock.Anything).Return(manualclient.Session{}, errors.New("Fail"))

	deviceConnMock := new(DeviceConnectionMock)
	deviceConnMock.On("Connect", mock.Anything)
	muxConn := usbmux.NewUsbMuxConnectionWithDeviceConnection(deviceConnMock)

	deviceConnMock.activeCodec = muxConn
	muxConn.Close()
	muxConn.Send(usbmux.NewReadDevices())
	actual, err := muxConn.Encode(usbmux.NewReadDevices())
	if assert.NoError(t, err) {
		golden := filepath.Join("test-fixture", "readdevices.bin")
		if *update {
			err := ioutil.WriteFile(golden, []byte(actual), 0644)
			if err != nil {
				log.Fatal(err)
			}
		}
		expected, _ := ioutil.ReadFile(golden)
		assert.ElementsMatch(t, actual, expected)

		f, err := os.Open(golden)
		if assert.NoError(t, err) {
			go func() {
				err := muxConn.Decode(f)
				if err != nil {
					log.Fatal("USBMux decoder failed in unit test")
				}
			}()
			decoded := <-muxConn.ResponseChannel
			log.Info(decoded)
			assert.ElementsMatch(t, decoded, []byte(usbmux.ToPlist(usbmux.NewReadDevices())))
		}

	}

}

type DeviceConnectionMock struct {
	mock.Mock
	activeCodec usbmux.Codec
}

func (mock *DeviceConnectionMock) ResumeReading()
func (mock *DeviceConnectionMock) Listen(activeCodec usbmux.Codec, c net.Conn) {}

func (conn *DeviceConnectionMock) EnableSessionSsl(pairRecord usbmux.PairRecord)
func (conn *DeviceConnectionMock) EnableSessionSslServerMode(pairRecord usbmux.PairRecord)
func (mock *DeviceConnectionMock) StopReadingAfterNextMessage()
func (mock *DeviceConnectionMock) ResumeReadingWithNewCodec(codec usbmux.Codec)
func (mock *DeviceConnectionMock) SetCodec(codec usbmux.Codec)

func (mock *DeviceConnectionMock) Connect(activeCodec usbmux.Codec) {}
func (mock *DeviceConnectionMock) ConnectToSocketAddress(activeCodec usbmux.Codec, socketAddress string) {
}
func (mock *DeviceConnectionMock) Close() {}
func (mock *DeviceConnectionMock) SendForProtocolUpgrade(muxConnection *usbmux.MuxConnection, message interface{}, newCodec usbmux.Codec) []byte {
	return nil
}
func (mock *DeviceConnectionMock) SendForProtocolUpgradeSSL(muxConnection *usbmux.MuxConnection, message interface{}, newCodec usbmux.Codec, pairRecord usbmux.PairRecord) []byte {
	return nil
}
func (mock *DeviceConnectionMock) Send(message interface{}) {}
func (mock *DeviceConnectionMock) SendForSslUpgrade(lockDownConn *usbmux.LockDownConnection, pairRecord usbmux.PairRecord) usbmux.StartSessionResponse {
	return usbmux.StartSessionResponse{}
}
