// Package hid injects touch, button and keyboard events over CoreDevice. Touch
// needs a media stream running to be accepted, which Session owns. iOS 27+.
package hid

import (
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/xpc"
)

const (
	universalServiceName = "com.apple.coredevice.hid.universalhidservice"

	universalFeatureIdentifier = "com.apple.coredevice.feature.remote.universalhidservice"
)

// _ServiceIDs of the surfaces the device registers. ListConnectedServices
// enumerates them, and its values are the ones to prefer.
const (
	SurfaceMainTouchscreen uint64 = 257 // 0x101
	// SurfaceTouchscreenGesture is the trackpad-style pointer surface. It moves a
	// mirroring host's cursor without putting a contact on the screen.
	SurfaceTouchscreenGesture uint64 = 1281 // 0x501
	// SurfaceKeyboardDefault is where a host-side virtual keyboard is registered.
	// Unlike the touch surfaces it does not pre-exist, so the value is ours.
	SurfaceKeyboardDefault uint64 = 0x100002001
)

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

// SendReport delivers a raw HID report, built by one of the Build* functions.
// Reports are not acknowledged, so a nil error means written, not acted on.
func (c *UniversalConnection) SendReport(serviceID uint64, report []byte) error {
	if len(report) == 0 {
		return fmt.Errorf("SendReport: report is empty")
	}
	if err := c.conn.Send(buildSendReportPayload(serviceID, report), xpc.HeartbeatRequestFlag); err != nil {
		return fmt.Errorf("SendReport: failed to send report to surface %d: %w", serviceID, err)
	}
	return nil
}

// SendTouchscreen posts one touchscreen report at (x, y). Every TouchContact
// means "in contact here", so a drag is a run of them ending in TouchRelease.
func (c *UniversalConnection) SendTouchscreen(state TouchState, x, y uint16, serviceID uint64) error {
	if err := c.SendReport(serviceID, BuildTouchscreenReport(state, x, y, Timestamp())); err != nil {
		return fmt.Errorf("SendTouchscreen: %w", err)
	}
	return nil
}

// SendDigitizer posts one pointer report at (x, y). It moves the cursor on the
// gesture surface; an on-screen touch needs SendTouchscreen.
func (c *UniversalConnection) SendDigitizer(x, y int32, serviceID uint64) error {
	if err := c.SendReport(serviceID, BuildDigitizerReport(x, y, Timestamp())); err != nil {
		return fmt.Errorf("SendDigitizer: %w", err)
	}
	return nil
}

// SendKeyboard posts one keyboard report. usages is the full set of keys held
// down, not a delta, so releasing one means resending without it.
func (c *UniversalConnection) SendKeyboard(serviceID uint64, usages ...uint8) error {
	if err := c.SendReport(serviceID, BuildKeyboardReport(usages, Timestamp())); err != nil {
		return fmt.Errorf("SendKeyboard: %w", err)
	}
	return nil
}

// CreateKeyboardService registers a virtual keyboard and returns the _ServiceID
// assigned to it. It must be created before any keyboard report is accepted.
func (c *UniversalConnection) CreateKeyboardService(serviceID uint64, product, manufacturer string, vendorID, productID int64) (uint64, error) {
	if err := c.conn.Send(buildCreateKeyboardPayload(serviceID, product, manufacturer, vendorID, productID), xpc.HeartbeatRequestFlag); err != nil {
		return 0, fmt.Errorf("CreateKeyboardService: failed to send request: %w", err)
	}
	res, err := c.conn.ReceiveOnServerClientStream()
	if err != nil {
		return 0, fmt.Errorf("CreateKeyboardService: failed to read response: %w", err)
	}
	// dtuhidd echoes the ID it settled on, and has been seen to omit it; the
	// requested one is what it assigns in that case.
	if assigned, ok := res["serviceID"].(uint64); ok {
		return assigned, nil
	}
	if assigned, ok := res["serviceID"].(int64); ok {
		return uint64(assigned), nil
	}
	return serviceID, nil
}

func (c *UniversalConnection) Close() error {
	return c.conn.Close()
}
