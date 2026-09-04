package display

import (
	"context"
	"errors"
	"slices"
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
	// invoke closes the connection from another goroutine to interrupt a read, so
	// every field is guarded.
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
		// Closing is what ends a blocked read on the real transport, so the fake
		// waits on the same signal rather than on a deadline.
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

func (f *fakeConn) requests() []map[string]interface{} {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return slices.Clone(f.sent)
}

func newService(c xpcConn) *Service {
	return &Service{conn: c, deviceID: uuid.New().String()}
}

func TestStopMediaStreamSendsStopAll(t *testing.T) {
	// The device rejects the request without stopAll, and rejects it as invalid
	// when false, so the request must carry it as true.
	f := newFakeConn()
	f.reply = map[string]interface{}{"CoreDevice.output": map[string]interface{}{}}
	s := newService(f)

	require.NoError(t, s.StopMediaStream(context.Background(), uuid.New()))
	sent := f.requests()
	require.Len(t, sent, 1)

	input, ok := sent[0]["CoreDevice.input"].(map[string]interface{})
	require.True(t, ok, "request carries an input dictionary")
	assert.Equal(t, true, input["stopAll"])
	assert.Contains(t, input, "avcMediaStreamOptionClientSessionID")
	assert.Equal(t, actionMediaStreamStop, sent[0]["CoreDevice.actionIdentifier"])
}

func TestStartVideoStreamRequiresAReceiver(t *testing.T) {
	s := newService(newFakeConn())

	_, err := s.StartVideoStream(context.Background(), VideoStreamRequest{SenderIP: "fd00::1"})
	assert.Error(t, err, "a stream with nowhere to send to must be refused")

	_, err = s.StartVideoStream(context.Background(), VideoStreamRequest{
		ReceiverIP: "fd00::2", ReceiverPort: 5000,
	})
	assert.Error(t, err, "a stream with no sender address must be refused")
}

func TestTimedOutRequestClosesTheConnection(t *testing.T) {
	// The read may have consumed part of a message, and nothing says whether it
	// did, so the connection cannot be reused.
	f := newFakeConn()
	f.replyWait = 500 * time.Millisecond
	f.reply = map[string]interface{}{}
	s := newService(f)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	_, err := s.invoke(ctx, featureStopMediaStream, actionMediaStreamStop, map[string]interface{}{})
	require.Error(t, err)
	assert.True(t, f.isClosed(), "an interrupted read leaves the stream position unknown, so the connection must go")

	// Nothing is left running: invoke returned because the read itself was
	// interrupted, so a later call simply fails on the closed connection.
	_, err = s.invoke(context.Background(), featureStopMediaStream, actionMediaStreamStop, map[string]interface{}{})
	assert.Error(t, err, "a later call must fail rather than decode the remains of the last one")
}

func TestInvokeReportsSendFailures(t *testing.T) {
	f := newFakeConn()
	f.sendErr = errors.New("channel gone")
	s := newService(f)

	_, err := s.invoke(context.Background(), featureStopMediaStream, actionMediaStreamStop, map[string]interface{}{})
	assert.ErrorContains(t, err, "channel gone")
	assert.True(t, f.isClosed(), "a failed write may be half a frame, so the connection must go")
}
