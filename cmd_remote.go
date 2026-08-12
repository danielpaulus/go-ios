package main

import (
	"fmt"
	"os"

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

	// resolver re-fetches fresh tunnel/RSD info so the screen stream self-heals
	// when the device's tunnel address changes. Non-fatal: on failure the screen
	// service keeps using the last known address and retries.
	tunnelInfo := tunnelInfoConfigFromArgs(ctx.Args)
	udid := ctx.Device.Properties.SerialNumber
	resolver := func() (ios.DeviceEntry, error) {
		return refreshDeviceTunnelInfo(ctx.Device, udid, tunnelInfo)
	}

	server, err := remote.NewServer(ctx.Device, driver, driverURL, resolver)
	exitIfError("failed starting remote server (developer disk image mounted?)", err)
	defer server.Close()

	exitIfError("remote server stopped", server.ListenAndServe(port))
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
