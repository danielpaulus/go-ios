package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/remote"
	"github.com/danielpaulus/go-ios/ios/tunnel"
	"github.com/docopt/docopt-go"
)

// runRemoteCommand serves the `ios remote` browser remote-control. The live
// screen is WDA-free (instruments screenshot service). Input is routed through a
// UI automation driver — DeviceKit by default (WDA is broken on iOS 26; DeviceKit
// works), selectable with --driver=<wda|devicekit>. The driver's base URL comes
// from --devicekit-url (default remote.DefaultDeviceKitURL / GO_IOS_DEVICEKIT_URL)
// or, for --driver=wda, --wda-url (default remote.DefaultWDAURL / GO_IOS_WDA_URL).
//
// By default (DeviceKit driver) `ios remote` also spawns and supervises the input
// runner (`ios ui run devicekit`), auto-respawning it when the on-device XCTest
// automation channel drops — so input self-heals instead of silently breaking.
// Pass --no-manage-runner to connect to an externally-started runner instead.
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

	// Manage the runner by default; --no-manage-runner opts out. Supervision only
	// applies to the DeviceKit runner (the one we can spawn), matched to driver.
	manageRunner := !boolArg(ctx.Args, "--no-manage-runner") && runnerDriver == remote.DriverDeviceKit

	// resolver re-fetches fresh tunnel/RSD info so the screen stream self-heals
	// when the device's tunnel address changes. Non-fatal: on failure the screen
	// service keeps using the last known address and retries.
	tunnelInfo := tunnelInfoConfigFromArgs(ctx.Args)
	udid := ctx.Device.Properties.SerialNumber
	resolver := func() (ios.DeviceEntry, error) {
		return refreshDeviceTunnelInfo(ctx.Device, udid, tunnelInfo)
	}

	server, err := remote.NewServer(ctx.Device, remote.Config{
		Driver:       driver,
		DriverURL:    driverURL,
		Resolver:     resolver,
		ManageRunner: manageRunner,
	})
	exitIfError("failed starting remote server (developer disk image mounted?)", err)
	defer server.Close()

	// Cancel on SIGINT/SIGTERM so the supervised runner is terminated cleanly
	// (no orphaned `ios ui run devicekit`) before we exit.
	runCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	exitIfError("remote server stopped", server.Run(runCtx, port))
}

// refreshDeviceTunnelInfo re-reads the device's tunnel info from the tunnel
// daemon and attaches a fresh RSD provider, mirroring resolveDevice's automatic
// path but WITHOUT exiting the process on failure — a transient tunnel gap must
// not kill a long-running `ios remote`. On any error the previous device entry
// is returned unchanged so the caller keeps using the last known address.
func refreshDeviceTunnelInfo(device ios.DeviceEntry, udid string, tunnelInfo tunnelInfoConfig) (ios.DeviceEntry, error) {
	info, err := tunnel.TunnelInfoForDevice(udid, tunnelInfo.Host, tunnelInfo.Port)
	if err != nil {
		return device, err
	}
	fresh, err := deviceWithTunnelInfo(udid, info)
	if err != nil {
		return device, err
	}
	return fresh, nil
}

// deviceWithTunnelInfo builds a DeviceEntry with an RSD provider for the given
// tunnel, the non-fatal counterpart to deviceWithRsdProvider (which exits on
// error).
func deviceWithTunnelInfo(udid string, info tunnel.Tunnel) (ios.DeviceEntry, error) {
	device, err := ios.GetDevice(udid)
	if err != nil {
		return device, err
	}
	rsdService, err := ios.NewWithAddrPortDevice(info.Address, info.RsdPort, device)
	if err != nil {
		return device, fmt.Errorf("connecting to RSD at %s:%d: %w", info.Address, info.RsdPort, err)
	}
	defer rsdService.Close()
	rsdProvider, err := rsdService.Handshake()
	if err != nil {
		return device, fmt.Errorf("RSD handshake at %s:%d: %w", info.Address, info.RsdPort, err)
	}
	device1, err := ios.GetDeviceWithAddress(udid, info.Address, rsdProvider)
	if err != nil {
		return device, err
	}
	device1.UserspaceTUN = info.UserspaceTUN
	device1.UserspaceTUNHost = ios.HttpApiHost()
	device1.UserspaceTUNPort = info.UserspaceTUNPort
	return device1, nil
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
