// Package hid drives the device's registered HID surfaces natively over
// RemoteXPC through the Developer Disk Image's dtuhidd daemon — no WDA, no
// DeviceKit, no XCUITest runner. It ports the touch path of pymobiledevice3's
// com.apple.coredevice.hid.universalhidservice.
//
// The service speaks the plain CoreDevice remote-feature envelope
// {featureIdentifier, messageType, payload} directly (the same shape as the
// pasteboard/deviceinfo RemoteXPC services), not the CoreDevice.* invoke
// envelope. send_report is fire-and-forget: the report bytes are handed to a
// specific HID surface identified by its numeric _ServiceID.
//
// A tap is one CONTACT report followed by one RELEASE report at the same
// coordinates, delivered to the mainTouchscreen surface (_ServiceID 257).
//
// Authentication gate: on some iOS versions dtuhidd only routes these reports
// to UIKit while an active CoreDevice media (screen) stream holds the HID
// surfaces authenticated. Without it backboardd logs "ignoring digitizer event
// ... from unsupported service" and drops the touch. Whether iOS 26.x needs
// this gate is determined empirically; see the PR notes.
package hid

import (
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
)

const logModule = "go-ios/hid"

// ServiceName is the RemoteXPC service name for the universal HID service.
const ServiceName = "com.apple.coredevice.hid.universalhidservice"

// featureIdentifier is the CoreDevice remote feature that dtuhidd's
// universalhidservice handler dispatches on.
const featureIdentifier = "com.apple.coredevice.feature.remote.universalhidservice"

// MainTouchscreenServiceID is the _ServiceID of the statically-registered real
// digitizer (the physical touchscreen), which accepts 58-byte report ID 0x09
// touch reports.
const MainTouchscreenServiceID uint64 = 257

// touchscreen report constants (report ID 0x09, 58 bytes).
const (
	touchscreenReportID     = 0x09
	touchscreenStateContact = 0xC2 // "contact in progress at this position"
	touchscreenStateRelease = 0x02 // release contact
	touchscreenReportLen    = 58
)

// Connection is a live connection to the universal HID service. It is not safe
// for concurrent use.
type Connection struct {
	conn   xpcConn
	device ios.DeviceEntry
}

// xpcConn is the subset of *xpc.Connection used here, extracted as an interface
// so the request logic can be exercised with a fake in unit tests.
type xpcConn interface {
	Send(data map[string]interface{}, flags ...uint32) error
	ReceiveOnServerClientStream() (map[string]interface{}, error)
	Close() error
}

// New connects to the universal HID service on the device. Requires a running
// tunnel and a mounted Developer Disk Image (iOS 17+).
func New(device ios.DeviceEntry) (*Connection, error) {
	xpcConn, err := ios.ConnectToXpcServiceTunnelIface(device, ServiceName)
	if err != nil {
		return nil, fmt.Errorf("hid.New: failed to connect to %s: %w", ServiceName, err)
	}
	return &Connection{conn: xpcConn, device: device}, nil
}

// Close closes the connection to the HID service.
func (c *Connection) Close() error {
	return c.conn.Close()
}

// buildTouchscreenReport builds the 58-byte mainTouchscreen HID report (report
// ID 0x09). state is touchscreenStateContact or touchscreenStateRelease; x and
// y are UInt16 (0..65535) normalized screen coordinates (32768 = center).
// timestamp is a 48-bit monotonic value; only its low 48 bits are used.
func buildTouchscreenReport(state byte, x, y uint16, timestamp uint64) []byte {
	report := make([]byte, touchscreenReportLen)
	report[0] = touchscreenReportID
	report[1] = 0x01
	report[2] = 0x05
	report[3] = state
	binary.LittleEndian.PutUint16(report[4:6], x)
	binary.LittleEndian.PutUint16(report[6:8], y)
	// report[8:40] stay zero (32 reserved bytes).
	report[40] = 0x02
	// report[41:44] stay zero.
	// 48-bit little-endian timestamp at report[44:50].
	ts := timestamp & ((1 << 48) - 1)
	for i := 0; i < 6; i++ {
		report[44+i] = byte(ts >> (8 * i))
	}
	// report[50:58] stay zero (8 reserved bytes).
	return report
}

// monotonicTimestamp returns a 48-bit monotonic timestamp for a HID report. The
// gesture recognizer only cares about monotonicity and inter-frame deltas, so a
// truncated nanosecond clock is sufficient.
func monotonicTimestamp() uint64 {
	return uint64(time.Now().UnixNano()) & ((1 << 48) - 1)
}

// sendReport delivers a raw HID report to the given surface. Fire-and-forget:
// the service sends no reply.
func (c *Connection) sendReport(serviceID uint64, report []byte) error {
	request := map[string]interface{}{
		"featureIdentifier": featureIdentifier,
		"messageType":       "Request",
		"payload": map[string]interface{}{
			"send": map[string]interface{}{
				"_0": report,
				"_1": serviceID,
			},
		},
	}
	if err := c.conn.Send(request); err != nil {
		return fmt.Errorf("sendReport: failed to send HID report to surface %d: %w", serviceID, err)
	}
	return nil
}

// Tap performs a single tap at the given normalized coordinates (0..65535 each,
// 32768 = screen center) on the main touchscreen: one CONTACT report followed
// by one RELEASE report at the same position.
func (c *Connection) Tap(x, y uint16) error {
	golog.Info("sending native HID tap", "module", logModule, "udid", c.device.Properties.SerialNumber, "x", x, "y", y, "serviceID", MainTouchscreenServiceID)
	contact := buildTouchscreenReport(touchscreenStateContact, x, y, monotonicTimestamp())
	if err := c.sendReport(MainTouchscreenServiceID, contact); err != nil {
		return fmt.Errorf("Tap: contact: %w", err)
	}
	release := buildTouchscreenReport(touchscreenStateRelease, x, y, monotonicTimestamp())
	if err := c.sendReport(MainTouchscreenServiceID, release); err != nil {
		return fmt.Errorf("Tap: release: %w", err)
	}
	return nil
}

// FractionToNormalized maps a screen fraction in [0,1] to the 0..65535 UInt16
// coordinate space the touchscreen surface expects (0.5 -> 32768). Values
// outside [0,1] are clamped.
func FractionToNormalized(f float64) uint16 {
	if f <= 0 {
		return 0
	}
	if f >= 1 {
		return 65535
	}
	return uint16(math.Round(f * 65535))
}
