package main

import (
	"fmt"
	"log/slog"

	"github.com/danielpaulus/go-ios/ios/hid"
)

// runHIDCommand handles `ios hid ...` — native, agent-free HID input over
// RemoteXPC (iOS 17+), no WDA/DeviceKit/XCUITest. Currently supports
// `hid tap --x=<0-65535> --y=<0-65535>`.
//
// dtuhidd only routes touch reports to UIKit while an active CoreDevice media
// (screen) stream holds the HID surfaces authenticated, so by default a tap
// briefly opens that stream (the auth gate). Pass --raw to send the reports
// without the gate (useful for testing whether a given iOS build needs it).
func runHIDCommand(ctx commandContext) {
	if !ctx.Device.SupportsRsd() {
		exitIfError("hid command requires iOS 17+ with tunnel", fmt.Errorf("tunnel not running. Start with: ios tunnel start"))
	}

	if tap, _ := ctx.Args.Bool("tap"); tap {
		x, err := ctx.Args.Int("--x")
		exitIfError("hid tap: --x must be an integer 0-65535", err)
		y, err := ctx.Args.Int("--y")
		exitIfError("hid tap: --y must be an integer 0-65535", err)
		if x < 0 || x > 65535 || y < 0 || y > 65535 {
			exitIfError("hid tap: coordinates out of range", fmt.Errorf("--x and --y must be in 0-65535 (32768 = screen center), got x=%d y=%d", x, y))
		}
		raw, _ := ctx.Args.Bool("--raw")

		conn, err := hid.New(ctx.Device)
		exitIfError("hid: failed to connect to universal HID service", err)
		defer func() {
			if closeErr := conn.Close(); closeErr != nil {
				slog.Error("Failed to close HID connection", "error", closeErr)
			}
		}()

		if !raw {
			session, err := hid.StartTouchSession(ctx.Device)
			exitIfError("hid tap: failed to open media-stream auth gate (use --raw to skip)", err)
			defer func() {
				if closeErr := session.Close(); closeErr != nil {
					slog.Warn("Failed to close touch session", "error", closeErr)
				}
			}()
		}

		err = conn.Tap(uint16(x), uint16(y))
		exitIfError("hid tap: failed to send tap", err)
		return
	}
}
