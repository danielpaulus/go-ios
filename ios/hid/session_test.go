package hid

import (
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These cover the session logic that needs no device. Stream negotiation is
// covered by ios/display's encoder tests.

// A closed session must fail fast rather than dereference its released
// connections.
func TestClosedSessionRejectsEverything(t *testing.T) {
	s := &Session{closed: true}

	assert.Error(t, s.TouchDown(Point{}))
	assert.Error(t, s.TouchUp(Point{}))
}

// Close runs on the teardown path of a failed open, so it must tolerate a
// session that never opened anything, and must be idempotent.
func TestCloseOnUnopenedSessionIsSafe(t *testing.T) {
	s := &Session{}

	require.NoError(t, s.Close())
	require.NoError(t, s.Close())
}

func TestNewSessionRejectsDeviceWithoutTunnel(t *testing.T) {
	_, err := NewSession(ios.DeviceEntry{})
	assert.Error(t, err)
}
