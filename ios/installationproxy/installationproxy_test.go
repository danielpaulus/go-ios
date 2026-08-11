package installationproxy

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"howett.net/plist"
)

// framePlist encodes a plist value into the installation_proxy wire format:
// a 4 byte big-endian length header followed by the plist payload. This mirrors
// what PlistCodec.Decode expects to read off the device connection.
func framePlist(t *testing.T, v interface{}) []byte {
	t.Helper()
	payload, err := plist.Marshal(v, plist.XMLFormat)
	require.NoError(t, err)
	buf := new(bytes.Buffer)
	require.NoError(t, binary.Write(buf, binary.BigEndian, uint32(len(payload))))
	buf.Write(payload)
	return buf.Bytes()
}

// browseResponsePlist builds a single BrowseResponse chunk as a plist map.
func browseResponsePlist(status string, currentIndex, currentAmount uint64, apps []map[string]interface{}) map[string]interface{} {
	list := make([]interface{}, len(apps))
	for i, a := range apps {
		list[i] = a
	}
	return map[string]interface{}{
		"Status":        status,
		"CurrentIndex":  currentIndex,
		"CurrentAmount": currentAmount,
		"CurrentList":   list,
	}
}

// fakeConn is a mocked DeviceConnectionInterface backed by an in-memory reader.
// It records the plist requests sent to it and replays framed plist responses.
type fakeConn struct {
	t        *testing.T
	requests []map[string]interface{}
	reader   io.Reader
	closed   bool
}

func newFakeConn(t *testing.T, reader io.Reader) *fakeConn {
	return &fakeConn{t: t, reader: reader}
}

func (f *fakeConn) Send(message []byte) error {
	require.GreaterOrEqual(f.t, len(message), 4, "plist messages carry a 4 byte length header")
	var request map[string]interface{}
	_, err := plist.Unmarshal(message[4:], &request)
	require.NoError(f.t, err)
	f.requests = append(f.requests, request)
	return nil
}

func (f *fakeConn) Reader() io.Reader { return f.reader }

func (f *fakeConn) Close() error {
	f.closed = true
	return nil
}

func (f *fakeConn) Read(p []byte) (int, error)  { return f.reader.Read(p) }
func (f *fakeConn) Write(p []byte) (int, error) { return len(p), f.Send(p) }
func (f *fakeConn) Writer() io.Writer           { return f }
func (f *fakeConn) Conn() net.Conn              { return nil }
func (f *fakeConn) DisableSessionSSL()          {}

func (f *fakeConn) EnableSessionSsl(ios.PairRecord) error                        { return nil }
func (f *fakeConn) EnableSessionSslServerMode(ios.PairRecord) error              { return nil }
func (f *fakeConn) EnableSessionSslHandshakeOnly(ios.PairRecord) error           { return nil }
func (f *fakeConn) EnableSessionSslServerModeHandshakeOnly(ios.PairRecord) error { return nil }

func newTestConnection(t *testing.T, reader io.Reader) *Connection {
	return &Connection{deviceConn: newFakeConn(t, reader), plistCodec: ios.NewPlistCodec()}
}

// endlessBrowseReader yields an unbounded stream of valid, framed BrowseResponse
// chunks whose Status is never "Complete". It generates bytes on demand so the
// test itself does not need to buffer infinite data in memory — it models a
// device that keeps sending "in progress" responses and never terminates.
type endlessBrowseReader struct {
	t   *testing.T
	buf []byte
	// count guards the test against a genuinely unbounded loop: if browseApps
	// ever asks for far more chunks than the cap allows, we stop generating so
	// the test fails loudly instead of hanging.
	count int
}

func (r *endlessBrowseReader) Read(p []byte) (int, error) {
	if len(r.buf) == 0 {
		if r.count > maxBrowseChunks*4 {
			// The fix should have aborted long before this; refuse to feed more.
			return 0, io.ErrUnexpectedEOF
		}
		r.count++
		app := map[string]interface{}{
			"CFBundleIdentifier": "com.example.app",
			"Path":               "/private/var/containers/Bundle/Application/app",
		}
		chunk := browseResponsePlist("in progress", 0, 1, []map[string]interface{}{app})
		r.buf = framePlist(r.t, chunk)
	}
	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

// TestBrowseAppsUnboundedLoopIsCapped reproduces issue #818: a device that never
// sends Status == "Complete" must not loop forever / grow memory without bound.
// With the fix, browseApps aborts with an error after the chunk cap.
func TestBrowseAppsUnboundedLoopIsCapped(t *testing.T) {
	conn := newTestConnection(t, &endlessBrowseReader{t: t})

	apps, err := conn.BrowseAllApps()

	require.Error(t, err, "an endless non-Complete stream must be aborted, not looped forever")
	assert.Contains(t, err.Error(), "chunks")
	assert.Empty(t, apps)
}

// TestBrowseAppsPropagatesDecodeError reproduces the error-shadowing problem: if
// the transport fails mid-read, browseApps must surface that decode error rather
// than swallowing it and spinning on a zero-value status.
func TestBrowseAppsPropagatesDecodeError(t *testing.T) {
	// A length header claiming a 16-byte payload followed by EOF: PlistCodec.Decode
	// fails in io.ReadFull with a payload-size error.
	header := make([]byte, 4)
	binary.BigEndian.PutUint32(header, 16)
	conn := newTestConnection(t, bytes.NewReader(header))

	apps, err := conn.BrowseAllApps()

	require.Error(t, err, "a decode failure must be propagated, not shadowed")
	assert.Contains(t, err.Error(), "incorrect size")
	assert.Empty(t, apps)
}

// TestBrowseAppsPropagatesGarbageError ensures a malformed plist payload (which
// Decode reads fine but plistFromBytes rejects) is also surfaced.
func TestBrowseAppsPropagatesGarbageError(t *testing.T) {
	payload := []byte("not a plist at all")
	buf := new(bytes.Buffer)
	require.NoError(t, binary.Write(buf, binary.BigEndian, uint32(len(payload))))
	buf.Write(payload)
	conn := newTestConnection(t, bytes.NewReader(buf.Bytes()))

	apps, err := conn.BrowseAllApps()

	require.Error(t, err)
	assert.Empty(t, apps)
}

// TestBrowseAppsHappyPath is the no-regression check: a couple of in-progress
// chunks followed by a Complete chunk must return all apps in order.
func TestBrowseAppsHappyPath(t *testing.T) {
	app0 := map[string]interface{}{"CFBundleIdentifier": "com.example.first"}
	app1 := map[string]interface{}{"CFBundleIdentifier": "com.example.second"}
	app2 := map[string]interface{}{"CFBundleIdentifier": "com.example.third"}

	var stream bytes.Buffer
	// Two in-progress chunks and a final Complete chunk. CurrentAmount is the
	// number of entries in that chunk; CurrentIndex is where they land in the
	// aggregated result. Total entries = 3.
	stream.Write(framePlist(t, browseResponsePlist("BrowsingApplications", 0, 1, []map[string]interface{}{app0})))
	stream.Write(framePlist(t, browseResponsePlist("BrowsingApplications", 1, 1, []map[string]interface{}{app1})))
	stream.Write(framePlist(t, browseResponsePlist("Complete", 2, 1, []map[string]interface{}{app2})))

	conn := newTestConnection(t, bytes.NewReader(stream.Bytes()))

	apps, err := conn.BrowseAllApps()

	require.NoError(t, err)
	require.Len(t, apps, 3)
	assert.Equal(t, "com.example.first", apps[0].CFBundleIdentifier())
	assert.Equal(t, "com.example.second", apps[1].CFBundleIdentifier())
	assert.Equal(t, "com.example.third", apps[2].CFBundleIdentifier())

	// The request sent to the device must be a Browse command (behavior preserved).
	fake := conn.deviceConn.(*fakeConn)
	require.Len(t, fake.requests, 1)
	assert.Equal(t, "Browse", fake.requests[0]["Command"])
}
