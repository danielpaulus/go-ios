package usbmux_test

import (
	"io"
	"testing"

	"github.com/danielpaulus/go-ios/usbmux"
	mock "github.com/stretchr/testify/mock"
)

func TestCodec(t *testing.T) {
	//mc := new(manualClientMock)
	//	mc.On("GetSessionInfo", mock.Anything, mock.Anything).Return(manualclient.Session{}, errors.New("Fail"))

	//muxConn.Send(usbmux.NewReadDevices())
	/*	golden := filepath.Join("test-fixture", "readdevices.bin")
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
			}()*/

	//assert.ElementsMatch(t, decoded, []byte(usbmux.ToPlist(usbmux.NewReadDevices())))
}

type DeviceConnectionMock struct {
	mock.Mock
}

func (mock *DeviceConnectionMock) Connect() {}
func (mock *DeviceConnectionMock) ConnectToSocketAddress(socketAddress string) {

}
func (mock *DeviceConnectionMock) Close() {}
func (mock *DeviceConnectionMock) Send(message []byte) error {

	return nil
}
func (mock *DeviceConnectionMock) Reader() io.Reader {
	return nil
}
func (mock *DeviceConnectionMock) Writer() io.Writer {
	return nil
}
func (mock *DeviceConnectionMock) EnableSessionSsl(pairRecord usbmux.PairRecord) error {
	return nil
}
func (mock *DeviceConnectionMock) EnableSessionSslServerMode(pairRecord usbmux.PairRecord) {}
func (mock *DeviceConnectionMock) EnableSessionSslHandshakeOnly(pairRecord usbmux.PairRecord) error {
	return nil
}
func (mock *DeviceConnectionMock) EnableSessionSslServerModeHandshakeOnly(pairRecord usbmux.PairRecord) {
}
func (mock *DeviceConnectionMock) DisableSessionSSL() {}
