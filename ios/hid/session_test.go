package hid

import (
	"context"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/display"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These cover the session logic that needs no device. Stream negotiation is
// covered by ios/display's encoder tests.

func TestInterpolate(t *testing.T) {
	tests := []struct {
		name        string
		from, to    uint16
		step, steps int
		want        uint16
	}{
		{name: "step 0 is the start point, which UIKit hit-tests", from: 1000, to: 5000, step: 0, steps: 4, want: 1000},
		{name: "single step lands on target", from: 0, to: 1000, step: 1, steps: 1, want: 1000},
		{name: "midpoint", from: 0, to: 1000, step: 1, steps: 2, want: 500},
		{name: "last step lands on target", from: 0, to: 1000, step: 4, steps: 4, want: 1000},
		{name: "descending", from: 1000, to: 0, step: 1, steps: 2, want: 500},
		{name: "no movement", from: 700, to: 700, step: 1, steps: 3, want: 700},
		{name: "full range does not overflow", from: 0, to: 65535, step: 8, steps: 8, want: 65535},
		{name: "full range midpoint", from: 0, to: 65535, step: 4, steps: 8, want: 32767},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, interpolate(tt.from, tt.to, tt.step, tt.steps))
		})
	}

	previous := interpolate(0, 5000, 0, 4)
	for i := 1; i <= 4; i++ {
		current := interpolate(0, 5000, i, 4)
		assert.Greater(t, current, previous, "samples must advance monotonically")
		previous = current
	}
}

func TestSleepCtx(t *testing.T) {
	start := time.Now()
	require.NoError(t, sleepCtx(context.Background(), 10*time.Millisecond))
	assert.GreaterOrEqual(t, time.Since(start), 10*time.Millisecond)

	// A zero or negative interval means "no pacing", not "report the context".
	assert.NoError(t, sleepCtx(context.Background(), 0))
	assert.NoError(t, sleepCtx(context.Background(), -time.Second))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	assert.ErrorIs(t, sleepCtx(ctx, time.Hour), context.Canceled)
}

// A closed session must fail fast rather than dereference its released
// connections.
func TestClosedSessionRejectsEverything(t *testing.T) {
	s := &Session{closed: true}
	ctx := context.Background()

	assert.Error(t, s.Tap(ctx, Point{X: 1, Y: 1}))
	assert.Error(t, s.Drag(ctx, Point{}, Point{X: 1, Y: 1}, 2, 0))
	assert.Error(t, s.MoveDigitizer(ctx, 1, 1))
	assert.Error(t, s.Type(ctx, "hello"))
	assert.Error(t, s.PressButton(ctx, 12, 64))
	assert.Error(t, s.TouchDown(ctx, Point{}))
	assert.Error(t, s.TouchMove(ctx, Point{}))
	assert.Error(t, s.TouchUp(ctx, Point{}))
	assert.Error(t, s.EnsureStream(ctx))
	_, err := s.ListServices()
	assert.Error(t, err)
}

// Close runs on the teardown path of a failed open, so it must tolerate a
// session that never opened anything, and must be idempotent.
func TestCloseOnUnopenedSessionIsSafe(t *testing.T) {
	s := &Session{}

	require.NoError(t, s.Close())
	require.NoError(t, s.Close())
	assert.False(t, s.StreamActive())
}

func TestNewSessionRejectsDeviceWithoutTunnel(t *testing.T) {
	_, err := NewSession(ios.DeviceEntry{})
	assert.Error(t, err)
}

// Touch input cannot work over a userspace tunnel: the device has no route to a
// host UDP socket.
func TestUserspaceTunnelIsRejectedForStreams(t *testing.T) {
	s := &Session{
		device: ios.DeviceEntry{UserspaceTUN: true, Address: "fd00::1"},
		hid:    &UniversalConnection{},
	}

	// Through the exported entry point, so callers warming the stream up front
	// see the same refusal a gesture would hit.
	err := s.EnsureStream(context.Background())
	assert.ErrorIs(t, err, display.ErrUserspaceTunnelUnsupported)
	assert.False(t, s.StreamActive(), "a rejected stream must leave no state behind")
}

// A stray release, which an input stream can produce, must not error or reach
// the device.
func TestTouchUpWithoutDownIsNoop(t *testing.T) {
	s := &Session{hid: &UniversalConnection{}}

	require.NoError(t, s.TouchUp(context.Background(), Point{X: 1, Y: 2}))
	assert.False(t, s.contactDown)
}

func TestWithFrameReaderReplacesTheDrain(t *testing.T) {
	called := false
	session := &Session{}
	WithFrameReader(func(*display.Receiver) error {
		called = true
		return nil
	})(session)

	require.NotNil(t, session.readFrames)
	require.NoError(t, session.readFrames(nil))
	assert.True(t, called, "the frame reader should be used instead of the drain")
}
