package hid

import (
	"fmt"
	"sync"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
)

const logModule = "go-ios/hid"

// Point is a position on the touchscreen, 0 to 65535 on each axis, independent
// of pixel size and orientation.
type Point struct {
	X uint16
	Y uint16
}

// hidConn is the part of the HID connection this package uses, extracted so
// gestures can be exercised without a device.
type hidConn interface {
	sendTouch(state touchState, p Point) error
	Close() error
}

// Session delivers touch input to one device.
//
// A media stream has to be running for any of this to reach the screen. Without
// one the device accepts every report and discards it, returning no error, so a
// caller that forgets sees silence rather than a failure. Starting that stream
// and keeping it up for the session's lifetime is the caller's job: ios/display
// does it, and this package deliberately does not.
type Session struct {
	device ios.DeviceEntry

	mutex sync.Mutex
	hid   hidConn

	// Tracks a contact held by the Touch* methods, so closing mid-gesture lifts it
	// rather than leaving the device believing a finger is down.
	contactDown bool
	lastContact Point

	closed bool
}

// NewSession connects to the device's HID service. iOS 27+ and a kernel tunnel.
//
// Touch does nothing until a media stream is running: the device accepts the
// reports and throws them away, returning no error. Starting that stream and
// keeping it up is the caller's job.
func NewSession(device ios.DeviceEntry) (*Session, error) {
	session := &Session{device: device}

	conn, err := newUniversal(device)
	if err != nil {
		return nil, fmt.Errorf("NewSession: %w", err)
	}
	session.hid = conn
	return session, nil
}

// TouchDown reports a contact at point and leaves it there. Call it again at a
// new point to move: the device tracks where the contact is, not what changed,
// so there is no separate move. TouchUp lifts it.
func (s *Session) TouchDown(point Point) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.checkOpen(); err != nil {
		return err
	}
	if err := s.hid.sendTouch(touchContact, point); err != nil {
		return fmt.Errorf("TouchDown: %w", err)
	}
	s.contactDown = true
	s.lastContact = point
	return nil
}

// TouchUp lifts the contact, and is a no-op when nothing is down. It takes no
// context deliberately: a caller tearing down is exactly when the finger has to
// come off the screen, so this must not be skippable.
func (s *Session) TouchUp(point Point) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.checkOpen(); err != nil {
		return err
	}
	if !s.contactDown {
		return nil
	}
	if err := s.hid.sendTouch(touchRelease, point); err != nil {
		return fmt.Errorf("TouchUp: %w", err)
	}
	s.contactDown = false
	return nil
}

// Close lifts any held contact and closes the HID connection. It is idempotent
// and waits for a gesture in flight.
func (s *Session) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true

	// Lift a contact left down by an interrupted input stream first: once the
	// stream is gone the device would keep believing a finger is on the screen.
	if s.contactDown {
		if err := s.hid.sendTouch(touchRelease, s.lastContact); err != nil {
			golog.Warn("failed to lift a held contact while closing, the device may still consider the screen touched",
				"module", logModule, "error", err)
		}
		s.contactDown = false
	}

	if s.hid != nil {
		if err := s.hid.Close(); err != nil {
			return fmt.Errorf("Session.Close: %w", err)
		}
		s.hid = nil
	}
	return nil
}

func (s *Session) checkOpen() error {
	if s.closed || s.hid == nil {
		return fmt.Errorf("session is closed")
	}
	return nil
}
