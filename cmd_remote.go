package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/danielpaulus/go-ios/ios/remote"
	"github.com/docopt/docopt-go"
)

// runRemoteCommand serves the `ios remote` browser remote-control. Both the live
// screen and input come from a single supervised DeviceKit runner: the screen is
// a passthrough of the runner's hardware video (H.264 at /video.h264, MJPEG at
// /screen as a fallback) and input is routed through a UI automation driver —
// DeviceKit by default (WDA is broken on iOS 26; DeviceKit works), selectable
// with --driver=<wda|devicekit>. The driver's base URL comes from --devicekit-url
// (default remote.DefaultDeviceKitURL / GO_IOS_DEVICEKIT_URL) or, for
// --driver=wda, --wda-url (default remote.DefaultWDAURL / GO_IOS_WDA_URL); the
// video is always proxied from the DeviceKit URL.
//
// By default (DeviceKit driver) `ios remote` also spawns and supervises the input
// runner (`ios ui run devicekit`), auto-respawning it when the on-device XCTest
// automation channel drops — so both screen and input self-heal instead of
// silently breaking. Pass --no-manage-runner to connect to an externally-started
// runner instead.
func runRemoteCommand(ctx commandContext) {
	port := portArg(ctx.Args)
	if port == "" {
		port = "8080"
	}

	driver, _ := ctx.Args.String("--driver")
	switch driver {
	case "", "auto":
		driver = remote.DriverDeviceKit
	case remote.DriverDeviceKit, remote.DriverWDA:
	default:
		exitIfError("invalid --driver for ios remote (use devicekit or wda)", fmt.Errorf("%q", driver))
	}

	// The runner we supervise is the input driver by default; --runner-driver lets
	// it be overridden, but only DeviceKit is a runner we know how to spawn.
	runnerDriver, _ := ctx.Args.String("--runner-driver")
	if runnerDriver == "" {
		runnerDriver = driver
	}

	var driverURL string
	if driver == remote.DriverWDA {
		driverURL, _ = ctx.Args.String("--wda-url")
		if driverURL == "" {
			driverURL = os.Getenv("GO_IOS_WDA_URL")
		}
		if driverURL == "" {
			driverURL = remote.DefaultWDAURL
		}
	} else {
		driverURL, _ = ctx.Args.String("--devicekit-url")
		if driverURL == "" {
			driverURL = os.Getenv("GO_IOS_DEVICEKIT_URL")
		}
		if driverURL == "" {
			driverURL = remote.DefaultDeviceKitURL
		}
	}

	// The video is proxied from the DeviceKit runner regardless of the input
	// driver. When the input driver is DeviceKit it shares driverURL; otherwise
	// (WDA input) resolve the DeviceKit URL independently.
	deviceKitURL := driverURL
	if driver != remote.DriverDeviceKit {
		deviceKitURL, _ = ctx.Args.String("--devicekit-url")
		if deviceKitURL == "" {
			deviceKitURL = os.Getenv("GO_IOS_DEVICEKIT_URL")
		}
		if deviceKitURL == "" {
			deviceKitURL = remote.DefaultDeviceKitURL
		}
	}

	// Manage the runner by default; --no-manage-runner opts out. Supervision only
	// applies to the DeviceKit runner (the one we can spawn), matched to driver.
	manageRunner := !boolArg(ctx.Args, "--no-manage-runner") && runnerDriver == remote.DriverDeviceKit

	// --fps/--bitrate tune the DeviceKit /h264 stream the browser decodes. They
	// default to remote.DefaultFPS / remote.DefaultBitrate (well above the
	// runner's own low defaults, for a smoother mirror).
	fps := intArgDefault(ctx.Args, "--fps", remote.DefaultFPS)
	bitrate := intArgDefault(ctx.Args, "--bitrate", remote.DefaultBitrate)

	server, err := remote.NewServer(ctx.Device, remote.Config{
		Driver:       driver,
		DriverURL:    driverURL,
		DeviceKitURL: deviceKitURL,
		ManageRunner: manageRunner,
		FPS:          fps,
		Bitrate:      bitrate,
	})
	exitIfError("failed starting remote server", err)
	defer server.Close()

	// Cancel on SIGINT/SIGTERM so the supervised runner is terminated cleanly
	// (no orphaned `ios ui run devicekit`) before we exit.
	runCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	exitIfError("remote server stopped", server.Run(runCtx, port))
}

// portArg reads --port. The global usage declares `ios forward … [--port=<mapping>]…`
// as repeatable, so docopt surfaces --port as a []string for every command; a
// plain args.String("--port") therefore returns "". Handle both shapes so
// `ios remote --port=8090` works.
func portArg(args docopt.Opts) string {
	if list, ok := args["--port"].([]string); ok {
		if len(list) > 0 {
			return list[len(list)-1]
		}
		return ""
	}
	s, _ := args.String("--port")
	return s
}
