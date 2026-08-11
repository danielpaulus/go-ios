package house_arrest

import (
	"bytes"
	"io"
	"net"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"howett.net/plist"
)

func TestVendContainerSuccessDoesNotFallBack(t *testing.T) {
	conn := newFakeHouseArrestConn(t, map[string]interface{}{"Status": "Complete"})
	dials := 0
	dial := func() (ios.DeviceConnectionInterface, error) {
		dials++
		return conn, nil
	}

	client, err := connect(dial, "com.example.app", "udid-1")
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.Equal(t, 1, dials)
	require.Len(t, conn.requests, 1)
	assert.Equal(t, "VendContainer", conn.requests[0]["Command"])
	assert.Equal(t, "com.example.app", conn.requests[0]["Identifier"])
	assert.False(t, conn.closed, "the vended connection must stay open for afc")
}

func TestFallsBackToVendDocumentsWhenVendContainerFails(t *testing.T) {
	containerConn := newFakeHouseArrestConn(t, map[string]interface{}{"Error": "InstallationLookupFailed"})
	documentsConn := newFakeHouseArrestConn(t, map[string]interface{}{"Status": "Complete"})
	conns := []*fakeHouseArrestConn{containerConn, documentsConn}
	dials := 0
	dial := func() (ios.DeviceConnectionInterface, error) {
		conn := conns[dials]
		dials++
		return conn, nil
	}

	client, err := connect(dial, "com.example.app", "udid-1")
	require.NoError(t, err)
	require.NotNil(t, client)

	assert.Equal(t, 2, dials)
	require.Len(t, containerConn.requests, 1)
	assert.Equal(t, "VendContainer", containerConn.requests[0]["Command"])
	assert.True(t, containerConn.closed, "the failed VendContainer connection must be closed")
	require.Len(t, documentsConn.requests, 1)
	assert.Equal(t, "VendDocuments", documentsConn.requests[0]["Command"])
	assert.Equal(t, "com.example.app", documentsConn.requests[0]["Identifier"])
	assert.False(t, documentsConn.closed, "the vended connection must stay open for afc")
}

func TestBothVendsFailingReturnsHelpfulError(t *testing.T) {
	containerConn := newFakeHouseArrestConn(t, map[string]interface{}{"Error": "InstallationLookupFailed"})
	documentsConn := newFakeHouseArrestConn(t, map[string]interface{}{"Error": "ApplicationLookupFailed"})
	conns := []*fakeHouseArrestConn{containerConn, documentsConn}
	dials := 0
	dial := func() (ios.DeviceConnectionInterface, error) {
		conn := conns[dials]
		dials++
		return conn, nil
	}

	client, err := connect(dial, "com.example.app", "udid-1")
	require.Error(t, err)
	assert.Nil(t, client)

	assert.Contains(t, err.Error(), "InstallationLookupFailed")
	assert.Contains(t, err.Error(), "ApplicationLookupFailed")
	assert.Contains(t, err.Error(), "UIFileSharingEnabled")
	assert.True(t, containerConn.closed)
	assert.True(t, documentsConn.closed)
}

func TestTransportErrorDoesNotFallBack(t *testing.T) {
	conn := newFakeHouseArrestConn(t, nil) // connection drops before a response arrives
	dials := 0
	dial := func() (ios.DeviceConnectionInterface, error) {
		dials++
		return conn, nil
	}

	client, err := connect(dial, "com.example.app", "udid-1")
	require.Error(t, err)
	assert.Nil(t, client)

	assert.ErrorIs(t, err, io.EOF)
	assert.NotContains(t, err.Error(), "UIFileSharingEnabled", "transport errors must not be explained as vend denials")
	assert.Equal(t, 1, dials, "a transport error must not trigger the VendDocuments fallback")
	assert.True(t, conn.closed)
}

// fakeHouseArrestConn is a mocked DeviceConnectionInterface that records the
// plist requests sent to it and replays a single canned house_arrest response.
// A nil response simulates a transport failure: reads hit EOF.
type fakeHouseArrestConn struct {
	t        *testing.T
	requests []map[string]interface{}
	readBuf  *bytes.Buffer
	closed   bool
}

func newFakeHouseArrestConn(t *testing.T, response map[string]interface{}) *fakeHouseArrestConn {
	if response == nil {
		return &fakeHouseArrestConn{t: t, readBuf: &bytes.Buffer{}}
	}
	framed, err := ios.NewPlistCodec().Encode(response)
	require.NoError(t, err)
	return &fakeHouseArrestConn{t: t, readBuf: bytes.NewBuffer(framed)}
}

func (f *fakeHouseArrestConn) Send(message []byte) error {
	require.GreaterOrEqual(f.t, len(message), 4, "plist messages carry a 4 byte length header")
	var request map[string]interface{}
	_, err := plist.Unmarshal(message[4:], &request)
	require.NoError(f.t, err)
	f.requests = append(f.requests, request)
	return nil
}

func (f *fakeHouseArrestConn) Reader() io.Reader { return f.readBuf }

func (f *fakeHouseArrestConn) Close() error {
	f.closed = true
	return nil
}

func (f *fakeHouseArrestConn) Read(p []byte) (int, error)  { return f.readBuf.Read(p) }
func (f *fakeHouseArrestConn) Write(p []byte) (int, error) { return len(p), f.Send(p) }
func (f *fakeHouseArrestConn) Writer() io.Writer           { return f }
func (f *fakeHouseArrestConn) Conn() net.Conn              { return nil }
func (f *fakeHouseArrestConn) DisableSessionSSL()          {}

func (f *fakeHouseArrestConn) EnableSessionSsl(ios.PairRecord) error           { return nil }
func (f *fakeHouseArrestConn) EnableSessionSslServerMode(ios.PairRecord) error { return nil }
func (f *fakeHouseArrestConn) EnableSessionSslHandshakeOnly(ios.PairRecord) error {
	return nil
}

func (f *fakeHouseArrestConn) EnableSessionSslServerModeHandshakeOnly(ios.PairRecord) error {
	return nil
}
