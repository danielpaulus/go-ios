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

	exitIfError("Device not found: "+udid, err)
	if addressErr == nil && rsdErr == nil {
		if userspaceTunnelErr == nil {
			device.UserspaceTUN = true
			device.UserspaceTUNHost = userspaceTunnelHost
			device.UserspaceTUNPort = userspaceTunnelPort
		}
		return deviceWithRsdProvider(device, udid, address, rsdPort)
	}

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
