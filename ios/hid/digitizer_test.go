package hid

import (
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDigitizer struct {
	events []digitizerEvent
}

func (f *fakeDigitizer) send(ev digitizerEvent) error {
	f.events = append(f.events, ev)
	return nil
}

func (f *fakeDigitizer) Close() error { return nil }

func openDigitizer() (*DigitizerSession, *fakeDigitizer) {
	f := &fakeDigitizer{}
	return &DigitizerSession{display: 2, conn: f}, f
}

// A swipe up from the bottom of a phone held upright: the edge is given when the
// contact is put down, and every later event of the contact repeats it.
func TestDigitizerContactPhasesAndEdge(t *testing.T) {
	s, f := openDigitizer()
	require.NoError(t, s.TouchDown(NormalizedPoint{X: 0.5, Y: 1}, EdgeBottom))
	require.NoError(t, s.TouchMove(NormalizedPoint{X: 0.5, Y: 0.7}))
	require.NoError(t, s.TouchUp(NormalizedPoint{X: 0.5, Y: 0.4}))

	require.Len(t, f.events, 3)
	assert.Equal(t, []digitizerPhase{digitizerBegan, digitizerMoved, digitizerEnded},
		[]digitizerPhase{f.events[0].phase, f.events[1].phase, f.events[2].phase})
	for _, ev := range f.events {
		assert.Equal(t, EdgeBottom, ev.edge, "the edge is where the contact started")
		assert.Equal(t, uint64(2), ev.display)
	}
}

func TestDigitizerTouchUpWithNothingDownSendsNothing(t *testing.T) {
	s, f := openDigitizer()
	require.NoError(t, s.TouchUp(NormalizedPoint{}))
	assert.Empty(t, f.events)
}

func TestDigitizerTouchMoveWithNothingDownFails(t *testing.T) {
	s, f := openDigitizer()
	assert.Error(t, s.TouchMove(NormalizedPoint{X: 0.5, Y: 0.5}))
	assert.Empty(t, f.events)
}

func TestDigitizerTouchDownWhileDownFails(t *testing.T) {
	s, f := openDigitizer()
	require.NoError(t, s.TouchDown(NormalizedPoint{X: 0.5, Y: 0.5}, EdgeNone))
	assert.Error(t, s.TouchDown(NormalizedPoint{X: 0.6, Y: 0.5}, EdgeNone))
	assert.Len(t, f.events, 1)
}

// Close lifts the contact where it was last moved to, not where it went down.
func TestDigitizerCloseLiftsAMovedContactWhereItIs(t *testing.T) {
	s, f := openDigitizer()
	require.NoError(t, s.TouchDown(NormalizedPoint{X: 0.5, Y: 1}, EdgeBottom))
	require.NoError(t, s.TouchMove(NormalizedPoint{X: 0.5, Y: 0.6}))
	require.NoError(t, s.Close())

	last := f.events[len(f.events)-1]
	assert.Equal(t, digitizerEnded, last.phase)
	assert.Equal(t, NormalizedPoint{X: 0.5, Y: 0.6}, last.point)
	assert.Equal(t, EdgeBottom, last.edge)
}

func TestDigitizerCloseLiftsAContactLeftDown(t *testing.T) {
	s, f := openDigitizer()
	require.NoError(t, s.TouchDown(NormalizedPoint{X: 0.25, Y: 0.75}, EdgeNone))
	require.NoError(t, s.Close())

	last := f.events[len(f.events)-1]
	assert.Equal(t, digitizerEnded, last.phase)
	assert.Equal(t, NormalizedPoint{X: 0.25, Y: 0.75}, last.point, "lifted where the finger was")
	require.NoError(t, s.Close(), "idempotent")
}

func TestDigitizerClosedSessionRejectsEverything(t *testing.T) {
	s := &DigitizerSession{closed: true}
	assert.Error(t, s.TouchDown(NormalizedPoint{}, EdgeNone))
	assert.Error(t, s.TouchMove(NormalizedPoint{}))
	assert.Error(t, s.TouchUp(NormalizedPoint{}))
}

func TestNewDigitizerSessionRejectsDeviceWithoutTunnel(t *testing.T) {
	_, err := NewDigitizerSession(ios.DeviceEntry{}, 0)
	assert.Error(t, err)
}

// dtuhidd decodes these with fixed types: the point as doubles and the rest as
// unsigned 64-bit integers.
func TestDigitizerPayload(t *testing.T) {
	body := roundTrip(t, buildDigitizerPayload(digitizerEvent{
		phase: digitizerMoved, point: NormalizedPoint{X: 0.25, Y: 0.75}, edge: EdgeRight, display: 2,
	}))

	assert.Equal(t, digitizerFeatureIdentifier, body["featureIdentifier"])
	assert.Equal(t, digitizerMessageType, body["messageType"])
	payload := dict(t, body, "payload")
	point := dict(t, payload, "pointOne")
	assert.Equal(t, 0.25, point["x"])
	assert.Equal(t, 0.75, point["y"])
	assert.Equal(t, uint64(1), payload["eventType"])
	assert.Equal(t, uint64(4), payload["edge"])
	assert.Equal(t, uint64(2), payload["target"])
}

// Folding a foldable switches the display that is on. The next contact goes to
// the new one; a contact already down is not dragged across to it.
func TestDigitizerSetDisplayAppliesFromTheNextContact(t *testing.T) {
	s, f := openDigitizer()
	require.NoError(t, s.TouchDown(NormalizedPoint{X: 0.5, Y: 0.5}, EdgeNone))
	s.SetDisplay(1)
	require.NoError(t, s.TouchMove(NormalizedPoint{X: 0.6, Y: 0.5}))
	require.NoError(t, s.TouchUp(NormalizedPoint{X: 0.6, Y: 0.5}))
	require.NoError(t, s.TouchDown(NormalizedPoint{X: 0.5, Y: 0.5}, EdgeNone))

	require.Len(t, f.events, 4)
	assert.Equal(t, []uint64{2, 2, 2, 1},
		[]uint64{f.events[0].display, f.events[1].display, f.events[2].display, f.events[3].display})
}
