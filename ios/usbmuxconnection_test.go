package ios_test

import (
	"io"
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"
	mock "github.com/stretchr/testify/mock"
)

func TestCodec(t *testing.T) {
	//mc := new(DeviceConnectionMock)

	//	mc.On("GetSessionInfo", mock.Anything, mock.Anything).Return(manualclient.Session{}, errors.New("Fail"))

	//muxConn.Send(ios.NewReadDevices())
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

	//assert.ElementsMatch(t, decoded, []byte(ios.ToPlist(ios.NewReadDevices())))
}

type DeviceConnectionMock struct {
	mock.Mock
}

func (mock *DeviceConnectionMock) ConnectToSocketAddress(socketAddress string) {
	mock.Called(socketAddress)

}
func (mock *DeviceConnectionMock) Close() {
	mock.Called()
}
func (mock *DeviceConnectionMock) Send(message []byte) error {
	args := mock.Called(message)
	return args.Error(0)
}
func (mock *DeviceConnectionMock) Reader() io.Reader {
	args := mock.Called()
	return args.Get(0).(io.Reader)
}
func (mock *DeviceConnectionMock) Writer() io.Writer {
	args := mock.Called()
	return args.Get(0).(io.Writer)
}
func (mock *DeviceConnectionMock) EnableSessionSsl(pairRecord ios.PairRecord) error {
	args := mock.Called(pairRecord)
	return args.Error(0)
}
func (mock *DeviceConnectionMock) EnableSessionSslServerMode(pairRecord ios.PairRecord) {
	mock.Called(pairRecord)
}
func (mock *DeviceConnectionMock) EnableSessionSslHandshakeOnly(pairRecord ios.PairRecord) error {
	args := mock.Called(pairRecord)
	return args.Error(0)
}
func (mock *DeviceConnectionMock) EnableSessionSslServerModeHandshakeOnly(pairRecord ios.PairRecord) {
	mock.Called(pairRecord)
}
func (mock *DeviceConnectionMock) DisableSessionSSL() {
	mock.Called()
}
