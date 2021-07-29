package ios_test

import (
	"bytes"
	"errors"
	"io"
	"net"
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestReleaseDeviceConnection(t *testing.T) {
	mc := new(DeviceConnectionMock)
	muxConn := ios.NewUsbMuxConnection(mc)
	deviceConn := muxConn.ReleaseDeviceConnection()
	assert.Equal(t, mc, deviceConn)
	err := muxConn.Send("")
	assert.Equal(t, io.EOF, err)
	err = muxConn.SendMuxMessage(ios.UsbMuxMessage{})
	assert.Equal(t, io.EOF, err)
	_, err = muxConn.ReadMessage()
	assert.Equal(t, io.EOF, err)
}

func TestCodec(t *testing.T) {
	mc := new(DeviceConnectionMock)
	buf := new(bytes.Buffer)
	mc.On("Writer").Return(buf)
	mc.On("Reader").Return(buf)

	muxConn := ios.NewUsbMuxConnection(mc)
	muxConn.Send(ios.NewReadDevices())

	msg, err := muxConn.ReadMessage()
	assert.NoError(t, err)
	firstTag := msg.Header.Tag
	assert.Equal(t, msg.Header.Length, uint32(len(msg.Payload)+16))

	buf.Reset()
	muxConn.Send(ios.NewReadDevices())
	msg, err = muxConn.ReadMessage()
	assert.NoError(t, err)
	secondTag := msg.Header.Tag

	assert.Equal(t, firstTag+1, secondTag)

	buf.Reset()
	randomTag := uint32(5000)
	msg.Header.Tag = randomTag
	muxConn.SendMuxMessage(msg)
	msg, err = muxConn.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, randomTag, msg.Header.Tag)
}

func TestErrors(t *testing.T) {
	mc := new(DeviceConnectionMock)
	writerMock := new(WriterMock)
	readerMock := new(ReaderMock)

	expectedError := errors.New("error")
	writerMock.On("Write", mock.Anything).Return(0, expectedError)
	readerMock.On("Read", mock.Anything).Return(0, expectedError)
	mc.On("Writer").Return(writerMock)
	mc.On("Reader").Return(readerMock)

	muxConn := ios.NewUsbMuxConnection(mc)
	err := muxConn.Send(ios.NewReadDevices())
	assert.Error(t, err)
	_, err = muxConn.ReadMessage()
	assert.Error(t, err)
}

type ReaderMock struct {
	mock.Mock
}

func (mock *ReaderMock) Read(p []byte) (n int, err error) {
	args := mock.Called(p)
	return args.Int(0), args.Error(1)
}

type WriterMock struct {
	mock.Mock
}

func (mock *WriterMock) Write(p []byte) (n int, err error) {
	args := mock.Called(p)
	return args.Int(0), args.Error(1)
}

type DeviceConnectionMock struct {
	mock.Mock
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
func (mock *DeviceConnectionMock) Conn() net.Conn {
	return nil
}
