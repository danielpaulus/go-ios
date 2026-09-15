package hid

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// report is one call recorded by fakeHID, so a test can assert what a gesture
// put on the wire rather than what it meant to.
type report struct {
	kind  string
	state touchState
	x, y  uint16
}

type fakeHID struct {
	reports []report
	// failAt makes the nth SendTouch call fail, to exercise the paths that
	// have to clean up after a gesture breaks part way through.
	failAt int
	calls  int
}

func (f *fakeHID) sendTouch(state touchState, p Point) error {
	f.calls++
	if f.failAt > 0 && f.calls == f.failAt {
		return errors.New("send failed")
	}
	f.reports = append(f.reports, report{kind: "touch", state: state, x: p.X, y: p.Y})
	return nil
}

func (f *fakeHID) Close() error { return nil }

// openSession builds a Session on a fake connection. The media stream that
// gates touch is the caller's concern now, so there is nothing else to stand up.
func openSession() (*Session, *fakeHID) {
	f := &fakeHID{}
	return &Session{hid: f}, f
}

func touches(reports []report) []report {
	var out []report
	for _, r := range reports {
		if r.kind == "touch" {
			out = append(out, r)
		}
	}
	return out
}

func TestCloseLiftsAContactLeftDown(t *testing.T) {
	session, fake := openSession()
	require.NoError(t, session.TouchDown(Point{X: 42, Y: 43}))

	before := len(touches(fake.reports))
	require.NoError(t, session.Close())

	got := touches(fake.reports)
	require.Greater(t, len(got), before, "closing has to lift the held contact")
	last := got[len(got)-1]
	assert.Equal(t, touchRelease, last.state)
	assert.Equal(t, [2]uint16{42, 43}, [2]uint16{last.x, last.y}, "lifted where the finger was")
}

func TestTouchUpWithNothingDownSendsNothing(t *testing.T) {
	session, fake := openSession()
	require.NoError(t, session.TouchUp(Point{X: 1, Y: 1}))
	assert.Empty(t, touches(fake.reports), "a stray release must not desynchronise the device")
}

func TestASecondContactKeepsTheFingerDown(t *testing.T) {
	session, fake := openSession()
	require.NoError(t, session.TouchDown(Point{X: 1, Y: 1}))
	require.NoError(t, session.TouchDown(Point{X: 2, Y: 2}))
	require.NoError(t, session.TouchUp(Point{X: 2, Y: 2}))

	got := touches(fake.reports)
	require.Len(t, got, 3)
	assert.Equal(t, touchContact, got[0].state)
	assert.Equal(t, touchContact, got[1].state, "a move is another contact, not a transition")
	assert.Equal(t, touchRelease, got[2].state)
}
