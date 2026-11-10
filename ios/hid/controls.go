package hid

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/deviceinfo"
	"github.com/danielpaulus/go-ios/ios/xpc"
)

// ErrNotFoldable is returned for a hinge angle sent to a device without one.
var ErrNotFoldable = errors.New("the device does not fold")

// closeGrace is how long Close waits after the last event. An event the
// connection is closed on straight away is lost; 50 ms was enough on an iPhone
// Duo with iOS 27.
const closeGrace = 250 * time.Millisecond

// messageConn sends one message and expects no reply. Extracted so Controls
// can be exercised without a device.
type messageConn interface {
	send(msg map[string]interface{}) error
	Close() error
}

type xpcMessageConn struct {
	conn *xpc.Connection
}

func (c *xpcMessageConn) send(msg map[string]interface{}) error {
	return c.conn.Send(msg, xpc.HeartbeatRequestFlag)
}

func (c *xpcMessageConn) Close() error {
	return c.conn.Close()
}

// Controls presses the device's buttons and folds it, the way Xcode's device
// hub does. iOS 27+ and a mounted Developer Disk Image. Buttons work on any
// device; only the hinge needs one that folds, so that is checked when it is
// set rather than here.
//
// The device answers none of these events, so a nil error means an event was
// sent, not that anything happened.
type Controls struct {
	mutex    sync.Mutex
	indigo   messageConn
	closed   bool
	lastSend time.Time
	grace    time.Duration

	// Told apart by the device's displays.
	foldable bool
}

// NewControls connects to the device's hid service, and reads its displays to
// find out whether it folds.
func NewControls(device ios.DeviceEntry) (*Controls, error) {
	foldable, err := isFoldable(device)
	if err != nil {
		return nil, fmt.Errorf("NewControls: %w", err)
	}
	conn, err := ios.ConnectToXpcServiceTunnelIface(device, indigoServiceName)
	if err != nil {
		return nil, fmt.Errorf("NewControls: %w", err)
	}
	return &Controls{
		indigo:   &xpcMessageConn{conn: conn},
		grace:    closeGrace,
		foldable: foldable,
	}, nil
}

// Foldable reports whether the device has a hinge, such as the iPhone Duo.
func (c *Controls) Foldable() bool {
	return c.foldable
}

// Close closes the connection, once the last event has had time to reach the
// device. It is idempotent.
func (c *Controls) Close() error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	if wait := c.grace - time.Since(c.lastSend); wait > 0 {
		time.Sleep(wait)
	}
	if err := c.indigo.Close(); err != nil {
		return fmt.Errorf("Controls.Close: %w", err)
	}
	return nil
}

func (c *Controls) send(msg map[string]interface{}) error {
	if c.closed {
		return fmt.Errorf("controls are closed")
	}
	c.lastSend = time.Now()
	return c.indigo.send(msg)
}

// isFoldable reports whether the device has more than one built-in display, as
// the iPhone Duo does: its inner and cover displays are both listed, folded or
// not.
func isFoldable(device ios.DeviceEntry) (bool, error) {
	info, err := deviceinfo.NewDeviceInfo(device)
	if err != nil {
		return false, fmt.Errorf("isFoldable: %w", err)
	}
	defer info.Close()
	displays, err := info.GetDisplayInfo()
	if err != nil {
		return false, fmt.Errorf("isFoldable: %w", err)
	}
	return countBuiltInDisplays(displays) > 1, nil
}

func countBuiltInDisplays(info map[string]interface{}) int {
	displays, _ := info["displays"].([]interface{})
	n := 0
	for _, d := range displays {
		display, ok := d.(map[string]interface{})
		if !ok {
			continue
		}
		if external, _ := display["external"].(bool); !external {
			n++
		}
	}
	return n
}
