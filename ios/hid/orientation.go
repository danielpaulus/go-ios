package hid

import (
	"fmt"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/xpc"
)

const (
	deviceControlServiceName = "com.apple.coredevice.devicecontrol"

	universalHIDServiceName       = "com.apple.coredevice.hid.universalhidservice"
	universalHIDFeatureIdentifier = "com.apple.coredevice.feature.remote.universalhidservice"
	// The universal HID service the virtualization controls go through.
	virtualizationServiceID uint64 = 0x500

	orientationFeatureIdentifier = "com.apple.coredevice.feature.remote.devicecontrol.orientation"
	orientationMessageType       = "OrientationRequest"
)

// Orientation is how the device is held, as UIDeviceOrientation names it. The
// display only turns with it when the app in front supports that orientation;
// the home screen, for one, stays in portrait.
type Orientation int

const (
	OrientationPortrait Orientation = iota
	// OrientationLandscapeLeft is the device on its side with the front camera
	// on the left.
	OrientationLandscapeLeft
	OrientationPortraitUpsideDown
	// OrientationLandscapeRight is the device on its side with the front camera
	// on the right.
	OrientationLandscapeRight
)

// orientationNames are the names an orientation goes by: in the CLI, to
// devicecontrol, and to a foldable's orientation picker.
type orientationNames struct {
	name          string
	deviceControl string
	picker        string
}

func (o Orientation) names() (orientationNames, bool) {
	switch o {
	case OrientationPortrait:
		return orientationNames{"portrait", "portrait", "portrait"}, true
	case OrientationLandscapeLeft:
		return orientationNames{"landscapeleft", "landscapeLeft", "landscape-left"}, true
	case OrientationPortraitUpsideDown:
		return orientationNames{"portraitupsidedown", "portraitUpsideDown", "pud"}, true
	case OrientationLandscapeRight:
		return orientationNames{"landscaperight", "landscapeRight", "landscape-right"}, true
	}
	return orientationNames{}, false
}

// String is the orientation's name, as the CLI calls it.
func (o Orientation) String() string {
	if n, ok := o.names(); ok {
		return n.name
	}
	return fmt.Sprintf("Orientation(%d)", int(o))
}

// Orientations lists every orientation.
func Orientations() []Orientation {
	return []Orientation{OrientationPortrait, OrientationLandscapeLeft, OrientationPortraitUpsideDown, OrientationLandscapeRight}
}

// ParseOrientation finds an orientation by name, ignoring case.
func ParseOrientation(name string) (Orientation, error) {
	names := make([]string, 0, len(Orientations()))
	for _, o := range Orientations() {
		if strings.EqualFold(o.String(), name) {
			return o, nil
		}
		names = append(names, o.String())
	}
	return 0, fmt.Errorf("unknown orientation %q, expected one of %s", name, strings.Join(names, ", "))
}

// SetOrientation turns the device to o, as if it were held that way, until it
// is set again or the device is turned for real.
//
// A foldable, such as the iPhone Duo, is turned with its orientation picker,
// the way the device hub's rotate button does. The device does not answer
// that, nor tell what it was last set to. It works at any hinge angle, but
// after a restart the device ignores the picker until the hinge has been set,
// so the hinge is set every time (see armOrientationPicker). As the device
// does not report its hinge angle, that moves the hinge to open flat when the
// inner display is in use and to closed otherwise: the display in use does
// not change, but a hinge set in between, such as half open, does not stay
// there. Other devices are turned through devicecontrol, which answers with
// the orientation the device is in now.
func (c *Controls) SetOrientation(o Orientation) error {
	n, ok := o.names()
	if !ok {
		return fmt.Errorf("SetOrientation: unknown orientation %d", int(o))
	}
	if c.foldable {
		if err := c.armOrientationPicker(); err != nil {
			return fmt.Errorf("SetOrientation: %w", err)
		}
		c.mutex.Lock()
		defer c.mutex.Unlock()
		if err := c.send(buildVirtualizationPayload(orientationControl, "enum", n.picker)); err != nil {
			return fmt.Errorf("SetOrientation: %w", err)
		}
		return nil
	}
	now, err := c.requestOrientation(map[string]interface{}{"changeOrientation": map[string]interface{}{"_0": n.deviceControl}})
	if err != nil {
		return fmt.Errorf("SetOrientation: %w", err)
	}
	if now != o {
		return fmt.Errorf("SetOrientation: asked for %s, the device is in %s", o, now)
	}
	return nil
}

// armOrientationPicker makes a foldable follow its orientation picker. After a
// restart it ignores the picker until the virtualization service has been
// sent an empty report and the hinge has been set, as the device hub does when
// it connects and folds. Both are sent every time, since whether they were
// already cannot be read. Nor can the hinge angle, so the hinge is set to the
// end of its range the active display belongs to: open flat when the larger,
// inner display is the active one, closed otherwise. The display stays, but
// an angle in between is lost.
func (c *Controls) armOrientationPicker() error {
	c.mutex.Lock()
	closed := c.closed
	c.mutex.Unlock()
	if closed {
		return fmt.Errorf("controls are closed")
	}
	displays, err := c.readDisplays()
	if err != nil {
		return fmt.Errorf("armOrientationPicker: read displays: %w", err)
	}
	angle := HingeClosed
	if innerDisplayActive(displays) {
		angle = HingeOpenFlat
	}

	conn, err := c.dialUniversalHID()
	if err != nil {
		return fmt.Errorf("armOrientationPicker: %w", err)
	}
	err = conn.send(map[string]interface{}{
		"featureIdentifier": universalHIDFeatureIdentifier,
		"messageType":       "Request",
		"payload": map[string]interface{}{
			"send": map[string]interface{}{"_0": emptyVirtualizationReport(), "_1": virtualizationServiceID},
		},
	})
	// Like the hid service, this one drops a message the connection is closed
	// on straight away.
	time.Sleep(c.grace)
	if closeErr := conn.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return fmt.Errorf("armOrientationPicker: %w", err)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	if err := c.send(buildVirtualizationPayload(hingeControl, "range", angle)); err != nil {
		return fmt.Errorf("armOrientationPicker: %w", err)
	}
	return nil
}

// emptyVirtualizationReport is the report the device hub sends the
// virtualization service when it connects: the vendor defined usage page
// (0xff61, little endian) and usage (0x5b), with no control in it.
func emptyVirtualizationReport() []byte {
	return []byte{0x81, 0x61, 0xff, 0x5b, 0x00, 0x00}
}

// innerDisplayActive reports whether the active built-in display is the larger
// of a foldable's two, the one used open.
func innerDisplayActive(info map[string]interface{}) bool {
	displays, _ := info["displays"].([]interface{})
	activeArea, largest := 0.0, 0.0
	for _, d := range displays {
		display, ok := d.(map[string]interface{})
		if !ok {
			continue
		}
		if external, _ := display["external"].(bool); external {
			continue
		}
		area := displayArea(display)
		largest = max(largest, area)
		if active, _ := display["active"].(bool); active {
			activeArea = area
		}
	}
	return activeArea > 0 && activeArea == largest
}

// displayArea is the display's size in pixels, from bounds [[x, y], [w, h]].
func displayArea(display map[string]interface{}) float64 {
	bounds, _ := display["bounds"].([]interface{})
	if len(bounds) != 2 {
		return 0
	}
	size, _ := bounds[1].([]interface{})
	if len(size) != 2 {
		return 0
	}
	return number(size[0]) * number(size[1])
}

func number(v interface{}) float64 {
	switch n := v.(type) {
	case float64:
		return n
	case int64:
		return float64(n)
	case uint64:
		return float64(n)
	}
	return 0
}

func dialUniversalHID(device ios.DeviceEntry) (messageConn, error) {
	conn, err := ios.ConnectToXpcServiceTunnelIface(device, universalHIDServiceName)
	if err != nil {
		return nil, fmt.Errorf("dialUniversalHID: %w", err)
	}
	return &xpcMessageConn{conn: conn}, nil
}

// requestOrientation sends one orientation request to devicecontrol and
// returns the orientation the device reports.
func (c *Controls) requestOrientation(request map[string]interface{}) (Orientation, error) {
	c.mutex.Lock()
	closed := c.closed
	c.mutex.Unlock()
	if closed {
		return 0, fmt.Errorf("controls are closed")
	}
	conn, err := c.dialDeviceControl()
	if err != nil {
		return 0, err
	}
	defer conn.Close()
	reply, err := conn.request(map[string]interface{}{
		"featureIdentifier": orientationFeatureIdentifier,
		"messageType":       orientationMessageType,
		"payload":           request,
	})
	if err != nil {
		return 0, err
	}
	return parseOrientationReply(reply)
}

// parseOrientationReply reads the orientation devicecontrol reports: the last
// one that was not face up or face down, which is the one the display follows.
func parseOrientationReply(reply map[string]interface{}) (Orientation, error) {
	name, _ := reply["currentDeviceNonFlatOrientation"].(string)
	for _, o := range Orientations() {
		if n, _ := o.names(); n.deviceControl == name {
			return o, nil
		}
	}
	return 0, fmt.Errorf("unexpected reply %v", reply)
}

// requestConn sends a message and waits for the reply. Extracted so Controls
// can be exercised without a device.
type requestConn interface {
	request(msg map[string]interface{}) (map[string]interface{}, error)
	Close() error
}

type xpcRequestConn struct {
	conn *xpc.Connection
}

func dialDeviceControl(device ios.DeviceEntry) (requestConn, error) {
	conn, err := ios.ConnectToXpcServiceTunnelIface(device, deviceControlServiceName)
	if err != nil {
		return nil, fmt.Errorf("dialDeviceControl: %w", err)
	}
	return &xpcRequestConn{conn: conn}, nil
}

func (c *xpcRequestConn) request(msg map[string]interface{}) (map[string]interface{}, error) {
	if err := c.conn.Send(msg, xpc.HeartbeatRequestFlag); err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	reply, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		return nil, fmt.Errorf("request: %w", err)
	}
	return reply, nil
}

func (c *xpcRequestConn) Close() error {
	return c.conn.Close()
}
