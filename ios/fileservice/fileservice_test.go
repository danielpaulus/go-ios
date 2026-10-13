package fileservice

import (
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeXpcConn implements controlConnection for tests. Responses are injected
// via channels; a stream with no injected response blocks like a real device
// that never answers.
type fakeXpcConn struct {
	controlResponses chan receiveResult // server->client (control) stream
	listResponses    chan receiveResult // client->server stream
	closed           chan struct{}
	closeOnce        sync.Once
}

func newFakeXpcConn() *fakeXpcConn {
	return &fakeXpcConn{
		controlResponses: make(chan receiveResult, 1),
		listResponses:    make(chan receiveResult, 1),
		closed:           make(chan struct{}),
	}
}

func (f *fakeXpcConn) Send(data map[string]interface{}, flags ...uint32) error {
	return nil
}

func (f *fakeXpcConn) ReceiveOnClientServerStream() (map[string]interface{}, error) {
	select {
	case res := <-f.listResponses:
		return res.response, res.err
	case <-f.closed:
		return nil, errors.New("connection closed")
	}
}

func (f *fakeXpcConn) ReceiveOnServerClientStream() (map[string]interface{}, error) {
	select {
	case res := <-f.controlResponses:
		return res.response, res.err
	case <-f.closed:
		return nil, errors.New("connection closed")
	}
}

func (f *fakeXpcConn) Close() error {
	f.closeOnce.Do(func() { close(f.closed) })
	return nil
}

func testConnection(t *testing.T, fake *fakeXpcConn, timeout time.Duration) *Connection {
	t.Helper()
	t.Cleanup(func() { fake.Close() })
	return &Connection{
		conn:           fake,
		sessionID:      "test-session",
		receiveTimeout: timeout,
	}
}

// listDirectoryWithDeadline fails the test instead of blocking forever if
// ListDirectory hangs (the bug from issue #784).
func listDirectoryWithDeadline(t *testing.T, c *Connection, path string) ([]string, error) {
	t.Helper()
	type result struct {
		files []string
		err   error
	}
	done := make(chan result, 1)
	go func() {
		files, err := c.ListDirectory(path)
		done <- result{files, err}
	}()
	select {
	case res := <-done:
		return res.files, res.err
	case <-time.After(5 * time.Second):
		t.Fatal("ListDirectory did not return within 5s, it hangs")
		return nil, nil
	}
}

func TestListDirectorySurfacesControlStreamError(t *testing.T) {
	fake := newFakeXpcConn()
	// The device reports the failure on the control stream (like every other
	// control response) and never sends anything on the list stream.
	fake.controlResponses <- receiveResult{response: map[string]interface{}{
		"EncodedError": map[string]interface{}{
			"Code":                 uint64(11007),
			"LocalizedDescription": "File paths cannot contain '..'.",
		},
	}}
	c := testConnection(t, fake, 5*time.Second)

	files, err := listDirectoryWithDeadline(t, c, "shared-data")

	require.Error(t, err)
	assert.Nil(t, files)
	var devErr *DeviceError
	require.ErrorAs(t, err, &devErr)
	assert.Equal(t, "File paths cannot contain '..'.", devErr.Description)
	assert.Contains(t, err.Error(), "File paths cannot contain '..'.")
}

func TestListDirectoryTimesOutWhenDeviceNeverResponds(t *testing.T) {
	fake := newFakeXpcConn()
	c := testConnection(t, fake, 50*time.Millisecond)

	start := time.Now()
	files, err := listDirectoryWithDeadline(t, c, ".")

	require.Error(t, err)
	assert.Nil(t, files)
	assert.ErrorIs(t, err, ErrTimeout)
	assert.Less(t, time.Since(start), 5*time.Second, "must fail fast via the receive timeout")
}

func TestListDirectoryReturnsFileList(t *testing.T) {
	fake := newFakeXpcConn()
	fake.listResponses <- receiveResult{response: map[string]interface{}{
		"FileList": []interface{}{"manifest.json", "sub"},
	}}
	c := testConnection(t, fake, time.Second)

	files, err := listDirectoryWithDeadline(t, c, ".")

	require.NoError(t, err)
	assert.Equal(t, []string{"manifest.json", "sub"}, files)
}

func TestPullFileConsumesPendingControlReadAfterListDirectory(t *testing.T) {
	fake := newFakeXpcConn()
	fake.listResponses <- receiveResult{response: map[string]interface{}{
		"FileList": []interface{}{},
	}}
	c := testConnection(t, fake, time.Second)

	// A successful ListDirectory leaves a pending read on the control stream.
	_, err := listDirectoryWithDeadline(t, c, ".")
	require.NoError(t, err)

	// The next control response must reach PullFile through that pending read.
	fake.controlResponses <- receiveResult{response: map[string]interface{}{
		"EncodedError": map[string]interface{}{
			"LocalizedDescription": "No such file or directory",
		},
	}}
	err = c.PullFile("missing.txt", io.Discard)
	require.Error(t, err)
	var devErr *DeviceError
	require.ErrorAs(t, err, &devErr)
	assert.Equal(t, "No such file or directory", devErr.Description)
}
