// Package hid injects HID events - touch, hardware buttons and virtual keyboard -
// into iOS 17+ devices through the CoreDevice HID services published by the
// Developer Disk Image's dtuhidd daemon.
//
// Two RemoteXPC services are wrapped:
//
//   - Indigo (com.apple.coredevice.hid.indigo) delivers hardware button events.
//   - Universal (com.apple.coredevice.hid.universalhidservice) enumerates the
//     HID surfaces already registered on the device and posts raw HID reports to
//     them. Touch and keyboard input are delivered this way.
//
// Both services use the plain {messageType, payload, featureIdentifier} envelope
// that dtuhidd recognises - not the CoreDevice.* envelope built by package
// coredevice.
//
// # Touch requires an active media stream
//
// Touch reports are gated on a running media stream. Without one, dtuhidd
// publishes the HID surfaces as eventSource: externalAccessory and backboardd
// discards every digitizer event ("ignoring digitizer event for display <main>
// from unsupported service"). Reports are still accepted without error, so a
// caller sees success while nothing happens on screen. Starting a media stream
// flips the surfaces to authenticated/builtIn and routes reports through to UIKit
// as real UIEventTypeTouches; the stream's RTP payload can be discarded, it only
// has to exist.
//
// Session handles this: it starts a stream through package ios/display the first
// time a gesture needs one and holds it open for the session's lifetime. Prefer
// Session over driving UniversalConnection directly, and reach for the raw
// SendReport/SendTouchscreen/SendKeyboard primitives only when managing the
// stream yourself.
//
// The gate is per client - a stream started by Xcode does not authenticate our
// reports - and it covers the touchscreen, gesture and keyboard surfaces.
// Hardware buttons are exempt: their surface is authenticated out of the box, so
// Session.PressButton never starts a stream.
package hid

import (
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/xpc"
)

const (
	indigoServiceName    = "com.apple.coredevice.hid.indigo"
	universalServiceName = "com.apple.coredevice.hid.universalhidservice"

	buttonFeatureIdentifier    = "com.apple.coredevice.feature.remote.hid.button"
	universalFeatureIdentifier = "com.apple.coredevice.feature.remote.universalhidservice"
)

// _ServiceIDs of the HID surfaces statically registered on the device. Use
// ListConnectedServices to discover others - the gesture surface in particular is
// session-specific in some captures, so prefer the enumerated value when present.
const (
	// SurfaceMainTouchscreen is the true digitizer, driven by 58-byte reports.
	SurfaceMainTouchscreen uint64 = 257 // 0x101
	// SurfaceTouchscreenGesture is the trackpad-style pointer surface, driven by
	// 19-byte reports. It moves the visual cursor in a mirror window but does not
	// by itself produce an on-screen touch.
	SurfaceTouchscreenGesture uint64 = 1281 // 0x501
	// SurfaceKeyboardDefault is the default _ServiceID for a host-registered
	// virtual keyboard. Bit 32 marks it session-specific, matching the convention
	// macOS Universal Control uses for mirrored peripherals.
	SurfaceKeyboardDefault uint64 = 0x100002001
)

// ButtonState is the state transition reported for a hardware button.
type ButtonState uint64

const (
	// ButtonDown reports the button being pressed.
	ButtonDown ButtonState = 1
	// ButtonUp reports the button being released.
	ButtonUp ButtonState = 2
	// ButtonCanceled abandons an in-progress press.
	ButtonCanceled ButtonState = 3
)

// IndigoConnection is a connection to the Indigo HID service, which delivers
// hardware button events.
//
// Only the button path is implemented. Indigo's other event kinds (keyboard,
// scroll, digitizer, vendor-defined) use Apple's Mercury peer-event envelope,
// whose on-wire form is not known; dtuhidd accepts a dispatch in that shape but
// logs "Resetting gesture state then canceling" without invoking a handler. Touch
// is delivered through UniversalConnection instead.
type IndigoConnection struct {
	conn *xpc.Connection
}

// NewIndigo connects to the Indigo HID service on the device. iOS 17+ only, and
// the Developer Disk Image must be mounted so dtuhidd is running.
func NewIndigo(device ios.DeviceEntry) (*IndigoConnection, error) {
	conn, err := ios.ConnectToXpcServiceTunnelIface(device, indigoServiceName)
	if err != nil {
		return nil, fmt.Errorf("NewIndigo: %w", err)
	}
	return &IndigoConnection{conn: conn}, nil
}

// SendButton reports a single hardware-button state change.
//
// usagePage and usageCode identify the button in HID terms - page 0x0C
// (Consumer) covers the media buttons, page 0x09 (Button) the generic ones.
func (c *IndigoConnection) SendButton(usagePage, usageCode uint64, state ButtonState) error {
	if err := c.conn.Send(buildButtonPayload(usagePage, usageCode, state), xpc.HeartbeatRequestFlag); err != nil {
		return fmt.Errorf("SendButton: failed to send button event: %w", err)
	}
	return nil
}

// Close closes the connection to the Indigo HID service.
func (c *IndigoConnection) Close() error {
	return c.conn.Close()
}

// UniversalConnection is a connection to the universalhidservice, which exposes
// the device's registered HID surfaces and accepts raw HID reports for them.
type UniversalConnection struct {
	conn *xpc.Connection
}

// NewUniversal connects to the universalhidservice on the device. iOS 17+ only,
// and the Developer Disk Image must be mounted so dtuhidd is running.
func NewUniversal(device ios.DeviceEntry) (*UniversalConnection, error) {
	conn, err := ios.ConnectToXpcServiceTunnelIface(device, universalServiceName)
	if err != nil {
		return nil, fmt.Errorf("NewUniversal: %w", err)
	}
	return &UniversalConnection{conn: conn}, nil
}

// ListConnectedServices enumerates the HID surfaces currently registered on the
// device. Each entry carries a _ServiceID usable as the target of SendReport.
func (c *UniversalConnection) ListConnectedServices() (map[string]interface{}, error) {
	if err := c.conn.Send(buildListServicesPayload(), xpc.HeartbeatRequestFlag); err != nil {
		return nil, fmt.Errorf("ListConnectedServices: failed to send request: %w", err)
	}
	res, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		return nil, fmt.Errorf("ListConnectedServices: failed to read response: %w", err)
	}
	return res, nil
}

// SendReport delivers a raw HID report to one of the device's HID surfaces.
//
// The report layout is surface-specific and its first byte is the HID report ID;
// use BuildTouchscreenReport, BuildDigitizerReport or BuildKeyboardReport to
// produce a well-formed one.
//
// dtuhidd does not acknowledge reports, so a nil error means the report was
// written to the channel, not that the device acted on it. See the package
// documentation on the media-stream gate.
func (c *UniversalConnection) SendReport(serviceID uint64, report []byte) error {
	if len(report) == 0 {
		return fmt.Errorf("SendReport: report is empty")
	}
	if err := c.conn.Send(buildSendReportPayload(serviceID, report), xpc.HeartbeatRequestFlag); err != nil {
		return fmt.Errorf("SendReport: failed to send report to surface %d: %w", serviceID, err)
	}
	return nil
}

// SendTouchscreen posts a single mainTouchscreen report at (x, y).
//
// Pass TouchContact for a touch sample and TouchRelease to lift. A tap is one
// TouchContact followed by one TouchRelease at the same coordinates; a drag is a
// run of TouchContact reports with advancing coordinates terminated by a single
// TouchRelease. There is no separate begin/end opcode - every TouchContact means
// "in contact at this position".
//
// Coordinates are in the digitizer's own units, not points; see ListConnectedServices
// for the surface's logical bounds.
func (c *UniversalConnection) SendTouchscreen(state TouchState, x, y uint16, serviceID uint64) error {
	if err := c.SendReport(serviceID, BuildTouchscreenReport(state, x, y, Timestamp())); err != nil {
		return fmt.Errorf("SendTouchscreen: %w", err)
	}
	return nil
}

// SendDigitizer posts a single gesture/pointer report at (x, y). This moves the
// visual cursor on the gesture surface; an actual on-screen touch also needs
// SendTouchscreen.
func (c *UniversalConnection) SendDigitizer(x, y int32, serviceID uint64) error {
	if err := c.SendReport(serviceID, BuildDigitizerReport(x, y, Timestamp())); err != nil {
		return fmt.Errorf("SendDigitizer: %w", err)
	}
	return nil
}

// SendKeyboard posts a single virtual-keyboard report to a surface registered by
// CreateKeyboardService.
//
// usages is the full set of HID keyboard usages currently held down, not a delta:
// every report replaces whatever the device believed was pressed, so releasing a
// key means resending without it. Pass no usages to release everything.
func (c *UniversalConnection) SendKeyboard(serviceID uint64, usages ...uint8) error {
	if err := c.SendReport(serviceID, BuildKeyboardReport(usages, Timestamp())); err != nil {
		return fmt.Errorf("SendKeyboard: %w", err)
	}
	return nil
}

// CreateKeyboardService registers a host-side virtual HID keyboard with dtuhidd
// and returns the _ServiceID it assigned, which is normally the requested one.
// Address subsequent SendKeyboard calls to the returned value.
//
// Unlike the touch surfaces this one is not pre-registered on the device, so it
// must be created before any keyboard report will be accepted. The same
// media-stream gate applies: without a running stream the new surface is
// published as externalAccessory and its reports are dropped.
func (c *UniversalConnection) CreateKeyboardService(serviceID uint64, product, manufacturer string, vendorID, productID int64) (uint64, error) {
	if err := c.conn.Send(buildCreateKeyboardPayload(serviceID, product, manufacturer, vendorID, productID), xpc.HeartbeatRequestFlag); err != nil {
		return 0, fmt.Errorf("CreateKeyboardService: failed to send request: %w", err)
	}
	res, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		return 0, fmt.Errorf("CreateKeyboardService: failed to read response: %w", err)
	}
	// dtuhidd echoes the ID it settled on; fall back to the requested one when the
	// response omits it, which is what the reference implementation does.
	if assigned, ok := res["serviceID"].(uint64); ok {
		return assigned, nil
	}
	if assigned, ok := res["serviceID"].(int64); ok {
		return uint64(assigned), nil
	}
	return serviceID, nil
}

// Close closes the connection to the universalhidservice.
func (c *UniversalConnection) Close() error {
	return c.conn.Close()
}
