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

// These tests cover the session logic that does not need a device: coordinate
// interpolation, context handling and the closed-session guards. The stream
// negotiation itself is covered by ios/display's encoder tests and by the
// on-device checks documented in the package.

func TestInterpolate(t *testing.T) {
	tests := []struct {
		name        string
		from, to    uint16
		step, steps int
		want        uint16
	}{
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
}

// TestInterpolateCoversTheWholeGesture pins the sampling contract Drag relies
// on: step 0 is the touch-down at the start point, which is what UIKit
// hit-tests, and the last step lands exactly on the target.
func TestInterpolateCoversTheWholeGesture(t *testing.T) {
	from, to := uint16(1000), uint16(5000)
	steps := 4

	assert.Equal(t, from, interpolate(from, to, 0, steps), "step 0 must be the start point")
	assert.Equal(t, to, interpolate(from, to, steps, steps), "the last step must be the target")

	previous := interpolate(from, to, 0, steps)
	for i := 1; i <= steps; i++ {
		current := interpolate(from, to, i, steps)
		assert.Greater(t, current, previous, "samples must advance monotonically")
		previous = current
	}
}

func TestSleepCtxReturnsAfterDuration(t *testing.T) {
	start := time.Now()
	require.NoError(t, sleepCtx(context.Background(), 10*time.Millisecond))
	assert.GreaterOrEqual(t, time.Since(start), 10*time.Millisecond)
}

func TestSleepCtxDoesNotSleepForNonPositiveDurations(t *testing.T) {
	// A zero interval means "no pacing", not "report the context".
	assert.NoError(t, sleepCtx(context.Background(), 0))
	assert.NoError(t, sleepCtx(context.Background(), -time.Second))
}

func TestSleepCtxHonoursCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sleepCtx(ctx, time.Hour)
	assert.ErrorIs(t, err, context.Canceled)
}

// TestClosedSessionRejectsGestures makes sure a closed session fails fast
// instead of dereferencing its released connections.
func TestClosedSessionRejectsGestures(t *testing.T) {
	s := &Session{closed: true}

	ctx := context.Background()
	assert.Error(t, s.Tap(ctx, Point{X: 1, Y: 1}))
	assert.Error(t, s.Drag(ctx, Point{}, Point{X: 1, Y: 1}, 2, 0))
	assert.Error(t, s.MoveDigitizer(ctx, 1, 1))
	assert.Error(t, s.Type(ctx, "hello"))
	assert.Error(t, s.PressButton(ctx, 12, 64))
	_, err := s.ListServices()
	assert.Error(t, err)
}

// TestCloseIsIdempotent covers the common teardown path: Close after a failed
// open, and Close twice, must not panic or error.
func TestCloseIsIdempotent(t *testing.T) {
	s := &Session{}

	require.NoError(t, s.Close())
	require.NoError(t, s.Close())
	assert.False(t, s.StreamActive())
}

// TestTeardownStreamOnFreshSessionIsNoop guards the partially-initialised paths
// ensureStream unwinds through when opening a stream fails.
func TestTeardownStreamOnFreshSessionIsNoop(t *testing.T) {
	s := &Session{}

	s.mutex.Lock()
	s.teardownStream()
	s.mutex.Unlock()

	assert.False(t, s.StreamActive())
}

// TestNewSessionRejectsDeviceWithoutTunnel documents that a session needs a
// tunnelled device: without RSD ports there is nothing to connect to.
func TestNewSessionRejectsDeviceWithoutTunnel(t *testing.T) {
	_, err := NewSession(ios.DeviceEntry{})
	assert.Error(t, err)
}

// TestUserspaceTunnelIsRejectedForStreams pins the documented behaviour: touch
// input cannot work over the userspace tunnel because the device cannot reach a
// host UDP socket.
func TestUserspaceTunnelIsRejectedForStreams(t *testing.T) {
	s := &Session{
		device: ios.DeviceEntry{UserspaceTUN: true, Address: "fd00::1"},
		hid:    &UniversalConnection{},
	}

	err := s.ensureStream(context.Background())
	assert.ErrorIs(t, err, display.ErrUserspaceTunnelUnsupported)
	assert.False(t, s.StreamActive(), "a rejected stream must leave no state behind")
}

// TestTouchUpWithoutDownIsNoop covers the stray-release case an input stream can
// produce: it must not error and must not touch the device.
func TestTouchUpWithoutDownIsNoop(t *testing.T) {
	s := &Session{hid: &UniversalConnection{}}

	require.NoError(t, s.TouchUp(context.Background(), Point{X: 1, Y: 2}))
	assert.False(t, s.contactDown)
}

// TestClosedSessionRejectsStreamingTouches keeps the streaming methods behind
// the same closed-session guard as the gesture methods.
func TestClosedSessionRejectsStreamingTouches(t *testing.T) {
	s := &Session{closed: true}
	ctx := context.Background()

	assert.Error(t, s.TouchDown(ctx, Point{}))
	assert.Error(t, s.TouchMove(ctx, Point{}))
	assert.Error(t, s.TouchUp(ctx, Point{}))
}

// TestCloseWithoutHeldContactDoesNotPanic guards the Close path that lifts a
// held contact: with nothing down it must not dereference the connection.
func TestCloseWithoutHeldContactDoesNotPanic(t *testing.T) {
	s := &Session{}
	require.NoError(t, s.Close())
}

func TestSessionOptions(t *testing.T) {
	s := &Session{displayID: display.DefaultDisplayID, typingInterval: defaultTypingInterval}

	WithDisplayID(7)(s)
	WithTypingInterval(5 * time.Millisecond)(s)

	assert.Equal(t, 7, s.displayID)
	assert.Equal(t, 5*time.Millisecond, s.typingInterval)
}
