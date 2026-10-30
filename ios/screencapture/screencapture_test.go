package screencapture

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeConn stands in for the XPC connection so the request/reply logic can be
// tested without a device.
type fakeConn struct {
	mutex     sync.Mutex
	sent      []map[string]interface{}
	sendErr   error
	reply     map[string]interface{}
	replyErr  error
	replyWait time.Duration
	closed    bool
	closeCh   chan struct{}
}

func newFakeConn() *fakeConn {
	return &fakeConn{closeCh: make(chan struct{})}
}

func (f *fakeConn) Send(data map[string]interface{}, flags ...uint32) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if f.closed {
		return errors.New("use of closed connection")
	}
	if f.sendErr != nil {
		return f.sendErr
	}
	f.sent = append(f.sent, data)
	return nil
}

func (f *fakeConn) ReceiveOnServerClientStream() (map[string]interface{}, error) {
	f.mutex.Lock()
	wait, reply, err := f.replyWait, f.reply, f.replyErr
	f.mutex.Unlock()

	if wait > 0 {
		select {
		case <-time.After(wait):
		case <-f.closeCh:
			return nil, errors.New("use of closed connection")
		}
	}
	return reply, err
}

func (f *fakeConn) Close() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	if !f.closed {
		f.closed = true
		close(f.closeCh)
	}
	return nil
}

func (f *fakeConn) isClosed() bool {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return f.closed
}

func (f *fakeConn) lastInput(t *testing.T) map[string]interface{} {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	require.NotEmpty(t, f.sent)
	input, ok := f.sent[len(f.sent)-1]["CoreDevice.input"].(map[string]interface{})
	require.True(t, ok, "request carries an input dictionary")
	return input
}

func newService(c xpcConn) *Service {
	return &Service{conn: c, deviceID: uuid.New().String()}
}

func imageReply(png []byte) map[string]interface{} {
	return map[string]interface{}{"CoreDevice.output": map[string]interface{}{
		"displayUniqueID": "123-456-789",
		"image":           png,
		"imageFormat":     "png",
	}}
}

func TestTakeScreenshotOfADisplay(t *testing.T) {
	f := newFakeConn()
	f.reply = imageReply([]byte("png bytes"))
	s := newService(f)

	image, err := s.TakeScreenshot(context.Background(), "123-456-789")

	require.NoError(t, err)
	assert.Equal(t, []byte("png bytes"), image)
	input := f.lastInput(t)
	// The device ignores a "displayUniqueId" key and captures the main display.
	assert.Equal(t, "123-456-789", input["displayUniqueID"])
	// The device rejects a request without a format.
	assert.Equal(t, "png", input["requestedFormat"])
	assert.Equal(t, actionCaptureScreenshot, f.sent[0]["CoreDevice.actionIdentifier"])
	assert.Equal(t, featureCaptureScreenshot, f.sent[0]["CoreDevice.featureIdentifier"])
}

func TestTakeScreenshotOfTheMainDisplay(t *testing.T) {
	f := newFakeConn()
	f.reply = imageReply([]byte("png bytes"))
	s := newService(f)

	_, err := s.TakeScreenshot(context.Background(), "")

	require.NoError(t, err)
	assert.NotContains(t, f.lastInput(t), "displayUniqueID", "no id means the main display")
}

func TestDeviceErrorKeepsTheConnection(t *testing.T) {
	f := newFakeConn()
	f.reply = map[string]interface{}{"CoreDevice.error": map[string]interface{}{
		"code":     int64(4865),
		"domain":   "NSCocoaErrorDomain",
		"userInfo": map[string]interface{}{"NSDebugDescription": "Expected to find key requestedFormat."},
	}}
	s := newService(f)

	_, err := s.TakeScreenshot(context.Background(), "")

	assert.ErrorContains(t, err, "Expected to find key requestedFormat.")
	assert.False(t, f.isClosed(), "a complete reply leaves the stream in a known position")
}

func TestMissingImageIsAnError(t *testing.T) {
	f := newFakeConn()
	f.reply = map[string]interface{}{"CoreDevice.output": map[string]interface{}{"imageFormat": "png"}}
	s := newService(f)

	_, err := s.TakeScreenshot(context.Background(), "")

	assert.ErrorContains(t, err, "no image")
}

func TestClosedConnectionIsReported(t *testing.T) {
	// What a host that did not announce its RemoteXPC version flags sees: the
	// daemon traps while replying and the connection ends.
	f := newFakeConn()
	f.replyErr = errors.New("EOF")
	s := newService(f)

	_, err := s.TakeScreenshot(context.Background(), "123-456-789")

	assert.ErrorContains(t, err, "EOF")
	assert.True(t, f.isClosed())
}

func TestTimedOutRequestClosesTheConnection(t *testing.T) {
	f := newFakeConn()
	f.replyWait = 500 * time.Millisecond
	f.reply = imageReply([]byte("png bytes"))
	s := newService(f)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := s.TakeScreenshot(ctx, "")

	assert.ErrorIs(t, err, context.DeadlineExceeded)
	assert.True(t, f.isClosed(), "an interrupted read leaves the stream position unknown, so the connection must go")
}

func TestSendFailureClosesTheConnection(t *testing.T) {
	f := newFakeConn()
	f.sendErr = errors.New("channel gone")
	s := newService(f)

	_, err := s.TakeScreenshot(context.Background(), "")

	assert.ErrorContains(t, err, "channel gone")
	assert.True(t, f.isClosed(), "a failed write may be half a frame, so the connection must go")
}
