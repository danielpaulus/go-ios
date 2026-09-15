// Package hid injects touch events over CoreDevice. A media stream has to be
// running for the device to apply them, and starting one is the caller's job:
// without it every report is accepted and discarded with no error. iOS 27+.
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

// surfaceMainTouchscreen is the service id the device gives its built in
// touchscreen. Every touch report goes there.
const surfaceMainTouchscreen uint64 = 257 // 0x101

type universalConnection struct {
	conn *xpc.Connection
}

// newUniversal connects to the universalhidservice on the device. iOS 27+: an
// iOS 18 device answers CoreDeviceError 9021, and below 27 the service is not
// in RSD at all. The Developer Disk Image must be mounted, because dtuhidd
// ships in the image rather than in the OS.
func newUniversal(device ios.DeviceEntry) (*universalConnection, error) {
	conn, err := ios.ConnectToXpcServiceTunnelIface(device, universalServiceName)
	if err != nil {
		return nil, fmt.Errorf("newUniversal: %w", err)
	}
	return &universalConnection{conn: conn}, nil
}

// The device never returns a response, so a nil error means it was sent, not that anything moved.
func (c *universalConnection) sendReport(serviceID uint64, report []byte) error {
	if len(report) == 0 {
		return fmt.Errorf("sendReport: report is empty")
	}
	if err := c.conn.Send(buildSendReportPayload(serviceID, report), xpc.HeartbeatRequestFlag); err != nil {
		return fmt.Errorf("sendReport: failed to send report to surface %d: %w", serviceID, err)
	}
	return nil
}

// sendTouch posts one touch report at p. Every TouchContact means "in contact
// here", so a drag is a run of them ending in TouchRelease.
func (c *universalConnection) sendTouch(state touchState, p Point) error {
	report := buildTouchscreenReport(state, p.X, p.Y, timestamp())
	if err := c.sendReport(surfaceMainTouchscreen, report); err != nil {
		return fmt.Errorf("sendTouch: %w", err)
	}
	return nil
}

func (c *universalConnection) Close() error {
	return c.conn.Close()
}
