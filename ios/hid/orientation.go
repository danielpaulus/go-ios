package hid

import (
	"fmt"
	"log/slog"
	"math"
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
// that, nor tell what it was last set to, and after a restart it ignores the
// picker until UnstickOrientation has been called. Other devices are turned
// through devicecontrol, which answers with the orientation the device is in
// now.
func (c *Controls) SetOrientation(o Orientation) error {
	n, ok := o.names()
	if !ok {
		return fmt.Errorf("SetOrientation: unknown orientation %d", int(o))
	}
	if c.foldable {
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

// UnstickOrientation makes a foldable follow SetOrientation again. After a
// restart it ignores its orientation picker until the virtualization service
// has been sent an empty report and the hinge has been set, as the device hub
// does when it connects and folds; folding it by hand does not count. Once
// done, it holds until the next restart. Whether it is needed cannot be read,
// so call it once after the device restarts, before turning it. It takes about
// a second, mostly to connect to the universal HID service.
//
// It sets the hinge to the angle the device goes by (see HingeAngleInUse), so
// the hinge and the display in use stay as they are. That is the angle read,
// unless it does not fit the display in use, as when the hinge is set right
// after a fold by hand, or cannot be read: then it is the end of the range the display in use goes
// with, open flat for the larger, inner display and closed for the cover
// display. The display in use stays then, but a hinge in between, such as half
// open, does not.
//
// A device that does not fold gets ErrNotFoldable.
func (c *Controls) UnstickOrientation() error {
	if !c.foldable {
		return fmt.Errorf("UnstickOrientation: %w", ErrNotFoldable)
	}
	degrees, err := c.unstickAngle()
	if err != nil {
		return fmt.Errorf("UnstickOrientation: %w", err)
	}
	if err := c.unstickOrientation(degrees); err != nil {
		return fmt.Errorf("UnstickOrientation: %w", err)
	}
	return nil
}

// UnstickOrientationAt does what UnstickOrientation does, but sets the hinge
// to degrees. An angle past where the displays switch, such as 130 on a
// closed device, switches them, as SetHingeAngle does.
//
// Deprecated: UnstickOrientation keeps the hinge where it is, as it reads the
// angle.
func (c *Controls) UnstickOrientationAt(degrees float64) error {
	if !c.foldable {
		return fmt.Errorf("UnstickOrientationAt: %w", ErrNotFoldable)
	}
	if math.IsNaN(degrees) || degrees < HingeClosed || degrees > HingeOpenFlat {
		return fmt.Errorf("UnstickOrientationAt: %v degrees is outside 0 to 180", degrees)
	}
	if err := c.unstickOrientation(degrees); err != nil {
		return fmt.Errorf("UnstickOrientationAt: %w", err)
	}
	return nil
}

// The hinge angles the displays are in use at, as measured on an iPhone Duo
// with iOS 27: the cover display at 90 degrees and below, the inner one at 130
// and above. In between it depends on which way the hinge moved.
const (
	coverDisplayMaxAngle = 90.0
	innerDisplayMinAngle = 130.0
)

// HingeAngleInUse is the hinge angle the device goes by, in degrees from 0
// (closed) to 180 (open flat), from the angle read (deviceinfo.GetHingeAngle)
// and the device's display info (deviceinfo.GetDisplayInfo).
//
// The displays follow whichever came last, the hinge being set or the device
// being folded by hand. The angle read normally does too, but after a fold by
// hand the hinge sensor keeps sending small new samples while it settles, and
// those win over the hinge being set meanwhile: set open after it was closed
// by hand, an iPhone Duo kept reading about 0 degrees while its inner display
// was in use. So the angle read is kept only when it fits the display in use: more than 90 degrees for the inner display, less than 130
// for the cover. Otherwise it is the end of the range that display goes with:
// 180 for the inner display, 0 for the cover. fits reports whether it is the
// angle read.
func HingeAngleInUse(read float64, displays map[string]interface{}) (degrees float64, fits bool) {
	read = clampHingeAngle(read)
	inner := innerDisplayInUse(displays)
	if inner && read > coverDisplayMaxAngle || !inner && read < innerDisplayMinAngle {
		return read, true
	}
	if inner {
		return HingeOpenFlat, false
	}
	return HingeClosed, false
}

// unstickAngle is the angle to set the hinge to while unsticking, so it stays
// as the device goes by it: the one in use (see HingeAngleInUse), or the end of
// the range the display in use goes with when the angle cannot be read.
func (c *Controls) unstickAngle() (float64, error) {
	c.mutex.Lock()
	closed := c.closed
	c.mutex.Unlock()
	if closed {
		return 0, fmt.Errorf("controls are closed")
	}
	read, readErr := c.readHingeAngle()
	displays, displaysErr := c.readDisplays()
	switch {
	case readErr != nil && displaysErr != nil:
		return 0, fmt.Errorf("read the hinge angle: %w; read displays: %w", readErr, displaysErr)
	case displaysErr != nil:
		return clampHingeAngle(read), nil
	case readErr != nil:
		end := HingeClosed
		if innerDisplayInUse(displays) {
			end = HingeOpenFlat
		}
		slog.Warn("cannot read the hinge angle; setting it to the end the display in use goes with", "degrees", end, "error", readErr)
		return end, nil
	}
	degrees, fits := HingeAngleInUse(read, displays)
	if !fits {
		slog.Info("the hinge angle read does not fit the display in use; setting it to the end that display goes with", "read", read, "degrees", degrees)
	}
	return degrees, nil
}

// unstickOrientation sends the virtualization service an empty report, then
// sets the hinge to degrees.
func (c *Controls) unstickOrientation(degrees float64) error {
	c.mutex.Lock()
	closed := c.closed
	c.mutex.Unlock()
	if closed {
		return fmt.Errorf("controls are closed")
	}
	conn, err := c.dialUniversalHID()
	if err != nil {
		return fmt.Errorf("connect to the universal HID service: %w", err)
	}
	sendErr := conn.send(map[string]interface{}{
		"featureIdentifier": universalHIDFeatureIdentifier,
		"messageType":       "Request",
		"payload": map[string]interface{}{
			"send": map[string]interface{}{"_0": emptyVirtualizationReport(), "_1": virtualizationServiceID},
		},
	})
	// Like the hid service, this one drops a message the connection is closed
	// on straight away.
	time.Sleep(c.grace)
	closeErr := conn.Close()
	if sendErr != nil {
		return fmt.Errorf("send the empty report: %w", sendErr)
	}
	if closeErr != nil {
		return fmt.Errorf("close the universal HID service: %w", closeErr)
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	if err := c.send(buildVirtualizationPayload(hingeControl, "range", degrees)); err != nil {
		return fmt.Errorf("set the hinge to %v degrees: %w", degrees, err)
	}
	return nil
}

// emptyVirtualizationReport is the report the device hub sends the
// virtualization service when it connects: the vendor defined usage page
// (0xff61, little endian) and usage (0x5b), with no control in it.
func emptyVirtualizationReport() []byte {
	return []byte{0x81, 0x61, 0xff, 0x5b, 0x00, 0x00}
}

// innerDisplayInUse reports whether the built-in display in use is the larger
// of a foldable's two, the one used open. The display in use is the one with
// its backlight on, and active only breaks a tie: the device can keep a dark
// display active, as when it is open while its hinge is set closed.
func innerDisplayInUse(info map[string]interface{}) bool {
	displays, _ := info["displays"].([]interface{})
	inUseArea, largest, best := 0.0, 0.0, 0
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
		rank := 0
		if backlight, _ := display["backlightState"].(string); backlight == "activeOn" {
			rank += 2
		}
		if active, _ := display["active"].(bool); active {
			rank++
		}
		if rank > best {
			inUseArea, best = area, rank
		}
	}
	return inUseArea > 0 && inUseArea == largest
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
