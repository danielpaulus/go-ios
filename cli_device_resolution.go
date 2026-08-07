package main

import (
	"log/slog"
	"os"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tunnel"
	"github.com/docopt/docopt-go"
)

type tunnelInfoConfig struct {
	Host string
	Port int
}

func tunnelInfoConfigFromArgs(arguments docopt.Opts) tunnelInfoConfig {
	tunnelInfoHost, err := arguments.String("--tunnel-info-host")
	if err != nil || tunnelInfoHost == "" {
		tunnelInfoHost = ios.HttpApiHost()
	}

	tunnelInfoPort, err := arguments.Int("--tunnel-info-port")
	if err != nil {
		tunnelInfoPort = ios.HttpApiPort()
	}

	return tunnelInfoConfig{
		Host: tunnelInfoHost,
		Port: tunnelInfoPort,
	}
}

func resolveDevice(arguments docopt.Opts, tunnelInfo tunnelInfoConfig) ios.DeviceEntry {
	udid, _ := arguments.String("--udid")
	if udid == "" {
		udid = os.Getenv("GO_IOS_UDID")
	}
	address, addressErr := arguments.String("--address")
	rsdPort, rsdErr := arguments.Int("--rsd-port")
	userspaceTunnelHost, userspaceTunnelHostErr := arguments.String("--userspace-host")
	if userspaceTunnelHostErr != nil {
		userspaceTunnelHost = ios.HttpApiHost()
	}

	userspaceTunnelPort, userspaceTunnelErr := arguments.Int("--userspace-port")

	device, err := ios.GetDevice(udid)
	if isTunnelCommand(arguments) {
		return device
	}

	if addressErr == nil && rsdErr == nil {
		if err != nil {
			// Explicit tunnel coordinates were given, so a device missing from
			// usbmuxd is not fatal: it may only be reachable through a tunnel
			// (e.g. from another host). Build the entry from the coordinates
			// and treat --udid as identity metadata.
			device = directTargetDevice(udid, address, userspaceTunnelErr == nil)
		}
		if userspaceTunnelErr == nil {
			device.UserspaceTUN = true
			device.UserspaceTUNHost = userspaceTunnelHost
			device.UserspaceTUNPort = userspaceTunnelPort
		}
		return deviceWithRsdProvider(device, udid, address, rsdPort)
	}

	exitIfError("Device not found: "+udid, err)

	if !needsAutomaticTunnelInfo(arguments) {
		return device
	}

	info, err := tunnel.TunnelInfoForDevice(device.Properties.SerialNumber, tunnelInfo.Host, tunnelInfo.Port)
	if err == nil {
		device.UserspaceTUNPort = info.UserspaceTUNPort
		device.UserspaceTUNHost = userspaceTunnelHost
		device.UserspaceTUN = info.UserspaceTUN
		return deviceWithRsdProvider(device, udid, info.Address, info.RsdPort)
	}

	slog.Warn("failed to get tunnel info", "udid", device.Properties.SerialNumber)
	return device
}

// Transport markers for devices reachable through a go-ios tunnel instead of
// usbmuxd. They show up as connectionType in `ios list` output, next to the
// usbmuxd-reported "USB"/"Network".
const (
	connectionTypeTunnel          = "tunnel"
	connectionTypeUserspaceTunnel = "userspaceTunnel"
)

// directTargetDevice builds a DeviceEntry from explicit --address/--rsd-port
// coordinates for a device that usbmuxd does not know. The udid (possibly
// empty) is identity metadata; the RSD handshake fills it in later if needed.
func directTargetDevice(udid string, address string, userspaceTunnel bool) ios.DeviceEntry {
	connectionType := connectionTypeTunnel
	if userspaceTunnel {
		connectionType = connectionTypeUserspaceTunnel
	}
	return ios.DeviceEntry{
		Properties: ios.DeviceProperties{
			SerialNumber:   udid,
			ConnectionType: connectionType,
		},
		Address: address,
	}
}

// tunnelBackedDeviceEntry converts a tunnel-agent entry into a DeviceEntry so
// tunnel-backed devices can be listed alongside usbmuxd ones.
func tunnelBackedDeviceEntry(t tunnel.Tunnel, tunnelHost string) ios.DeviceEntry {
	connectionType := connectionTypeTunnel
	if t.UserspaceTUN {
		connectionType = connectionTypeUserspaceTunnel
	}
	return ios.DeviceEntry{
		Properties: ios.DeviceProperties{
			SerialNumber:   t.Udid,
			ConnectionType: connectionType,
		},
		Address:          t.Address,
		UserspaceTUN:     t.UserspaceTUN,
		UserspaceTUNHost: tunnelHost,
		UserspaceTUNPort: t.UserspaceTUNPort,
	}
}

// tunnelBackedDevices queries the tunnel agent's /tunnels HTTP API and returns
// the tunnels as device entries. An error just means no agent is reachable.
func tunnelBackedDevices(tunnelInfo tunnelInfoConfig) ([]ios.DeviceEntry, error) {
	tunnels, err := tunnel.ListRunningTunnels(tunnelInfo.Host, tunnelInfo.Port)
	if err != nil {
		return nil, err
	}
	entries := make([]ios.DeviceEntry, len(tunnels))
	for i, t := range tunnels {
		entries[i] = tunnelBackedDeviceEntry(t, tunnelInfo.Host)
	}
	return entries, nil
}

// mergeTunnelDevices appends tunnel-backed devices that usbmuxd does not
// already report, so `ios list` shows devices that are only reachable via a
// tunnel. usbmuxd entries win for devices known to both sources.
func mergeTunnelDevices(deviceList ios.DeviceList, tunnelDevices []ios.DeviceEntry) ios.DeviceList {
	known := make(map[string]bool, len(deviceList.DeviceList))
	for _, d := range deviceList.DeviceList {
		known[d.Properties.SerialNumber] = true
	}
	for _, d := range tunnelDevices {
		udid := d.Properties.SerialNumber
		// A tunnel entry without a udid cannot be addressed or deduplicated by
		// identity, so skip it rather than let it collapse with other unknowns.
		if udid == "" || known[udid] {
			continue
		}
		known[udid] = true
		deviceList.DeviceList = append(deviceList.DeviceList, d)
	}
	return deviceList
}

// isTunnelOnlyDevice reports whether the entry came from tunnel discovery (or
// explicit tunnel coordinates) rather than usbmuxd, meaning usbmuxd-based
// services like lockdown are not reachable for it on this host.
func isTunnelOnlyDevice(device ios.DeviceEntry) bool {
	connectionType := device.Properties.ConnectionType
	return connectionType == connectionTypeTunnel || connectionType == connectionTypeUserspaceTunnel
}

func needsAutomaticTunnelInfo(args docopt.Opts) bool {
	if boolArg(args, "rsd") || boolArg(args, "file") || boolArg(args, "webinspector") {
		return true
	}
	if boolArg(args, "info") && boolArg(args, "display") {
		return true
	}
	// `ui run` launches an XCUITest runner via testmanagerd, which needs the
	// tunnel on iOS 17+; the other `ui` commands are HTTP-to-a-URL and don't.
	if boolArg(args, "ui") && boolArg(args, "run") {
		return true
	}

	for _, commandName := range []string{
		"debug",
		"devicestate",
		"instruments",
		"kill",
		"launch",
		"memlimitoff",
		"ostrace",
		"pasteboard",
		"ps",
		"resetlocation",
		"runwda",
		"runxctest",
		"runtest",
		"screenshot",
		"setlocation",
		"setlocationgpx",
		"syslog",
		"sysmontap",
	} {
		if boolArg(args, commandName) {
			return true
		}
	}

	return false
}
