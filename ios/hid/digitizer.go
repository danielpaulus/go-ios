package hid

import (
	"fmt"
	"sync"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/xpc"
)

const (
	indigoServiceName = "com.apple.coredevice.hid.indigo"

	digitizerFeatureIdentifier = "com.apple.coredevice.feature.remote.hid.digitizer"
	digitizerMessageType       = "IndigoDigitizerEvent"
)

// Edge is the screen edge a touch starts from, in the display's native
// (unrotated) axes. A touch that starts from an edge can begin a system gesture,
// such as swiping up from the bottom to go home, which a touch reported through
// Session never does.
type Edge uint64

const (
	EdgeNone   Edge = 0
	EdgeTop    Edge = 1
	EdgeLeft   Edge = 2
	EdgeBottom Edge = 3
	EdgeRight  Edge = 4
)

// The phase of a contact on the wire.
type digitizerPhase uint64

const (
	digitizerBegan digitizerPhase = 0
	digitizerMoved digitizerPhase = 1
	digitizerEnded digitizerPhase = 2
)

// NormalizedPoint is a position on a display, 0 to 1 on each axis, in the
// display's native (unrotated) axes.
type NormalizedPoint struct {
	X float64
	Y float64
}

type digitizerEvent struct {
	phase   digitizerPhase
	point   NormalizedPoint
	edge    Edge
	display uint64
}

// digitizerConn is the part of the connection DigitizerSession uses, extracted
// so contacts can be exercised without a device.
type digitizerConn interface {
	send(ev digitizerEvent) error
	Close() error
}

type indigoConnection struct {
	conn *xpc.Connection
}

func (c *indigoConnection) send(ev digitizerEvent) error {
	if err := c.conn.Send(buildDigitizerPayload(ev), xpc.HeartbeatRequestFlag); err != nil {
		return fmt.Errorf("send: failed to send digitizer event to display %d: %w", ev.display, err)
	}
	return nil
}

func (c *indigoConnection) Close() error {
	return c.conn.Close()
}

func buildDigitizerPayload(ev digitizerEvent) map[string]interface{} {
	return map[string]interface{}{
		"featureIdentifier": digitizerFeatureIdentifier,
		"messageType":       digitizerMessageType,
		"payload": map[string]interface{}{
			"pointOne":  map[string]interface{}{"x": ev.point.X, "y": ev.point.Y},
			"eventType": uint64(ev.phase),
			"edge":      uint64(ev.edge),
			"target":    ev.display,
		},
	}
}

// DigitizerSession delivers touch input to one display, the way Xcode's device
// remote control does. Unlike Session it can mark where a touch comes from, so
// edge gestures reach SpringBoard. iOS 27+ and a mounted Developer Disk Image.
type DigitizerSession struct {
	// The display the next contact goes to. A held contact keeps contactDisplay,
	// so changing this mid-gesture does not move a finger to another screen.
	display        uint64
	contactDisplay uint64

	mutex sync.Mutex
	conn  digitizerConn

	// The contact held by TouchDown, so closing mid-gesture lifts it rather than
	// leaving the device believing a finger is down.
	contactDown bool
	lastContact NormalizedPoint
	edge        Edge

	closed bool
}

// NewDigitizerSession connects to the device's digitizer service for the display
// with the given id, as deviceinfo's GetDisplayInfo lists it ("displayId"). Zero
// is the main display.
func NewDigitizerSession(device ios.DeviceEntry, displayID uint64) (*DigitizerSession, error) {
	conn, err := ios.ConnectToXpcServiceTunnelIface(device, indigoServiceName)
	if err != nil {
		return nil, fmt.Errorf("NewDigitizerSession: %w", err)
	}
	return &DigitizerSession{display: displayID, conn: &indigoConnection{conn: conn}}, nil
}

// TouchDown puts a contact at point, or moves the one already down there. The
// edge is taken from the first TouchDown of a contact and kept until TouchUp:
// it says where the contact started, not where it is.
func (s *DigitizerSession) TouchDown(point NormalizedPoint, edge Edge) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.checkOpen(); err != nil {
		return err
	}
	phase := digitizerMoved
	if !s.contactDown {
		phase = digitizerBegan
		s.edge = edge
		s.contactDisplay = s.display
	}
	if err := s.conn.send(digitizerEvent{phase: phase, point: point, edge: s.edge, display: s.contactDisplay}); err != nil {
		return fmt.Errorf("TouchDown: %w", err)
	}
	s.contactDown = true
	s.lastContact = point
	return nil
}

// SetDisplay sends the contacts that begin from now on to another display, such
// as the cover display of a foldable once it is folded. A contact already down
// stays where it began.
func (s *DigitizerSession) SetDisplay(displayID uint64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.display = displayID
}

// TouchUp lifts the contact, and is a no-op when nothing is down.
func (s *DigitizerSession) TouchUp(point NormalizedPoint) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if err := s.checkOpen(); err != nil {
		return err
	}
	if !s.contactDown {
		return nil
	}
	if err := s.conn.send(digitizerEvent{phase: digitizerEnded, point: point, edge: s.edge, display: s.contactDisplay}); err != nil {
		return fmt.Errorf("TouchUp: %w", err)
	}
	s.contactDown = false
	return nil
}

// Close lifts any held contact and closes the connection. It is idempotent.
func (s *DigitizerSession) Close() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true

	if s.contactDown {
		ev := digitizerEvent{phase: digitizerEnded, point: s.lastContact, edge: s.edge, display: s.contactDisplay}
		if err := s.conn.send(ev); err != nil {
			golog.Warn("failed to lift a held contact while closing, the device may still consider the screen touched",
				"module", logModule, "error", err)
		}
		s.contactDown = false
	}

	if s.conn != nil {
		if err := s.conn.Close(); err != nil {
			return fmt.Errorf("DigitizerSession.Close: %w", err)
		}
		s.conn = nil
	}
	return nil
}

func (s *DigitizerSession) checkOpen() error {
	if s.closed || s.conn == nil {
		return fmt.Errorf("session is closed")
	}
	return nil
}
