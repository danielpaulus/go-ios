package main

import (
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/afc"
	"github.com/danielpaulus/go-ios/ios/deviceinfo"
	"github.com/danielpaulus/go-ios/ios/diagnostics"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/pcap"
)

func runIPCommand(ctx commandContext) {
	ip, err := pcap.FindIp(ctx.Device)
	exitIfError("failed", err)
	fmt.Println(convertToJSONString(ip))
}

func runInfoCommand(ctx commandContext) {
	if display, _ := ctx.Args.Bool("display"); display {
		deviceInfo, err := deviceinfo.NewDeviceInfo(ctx.Device)
		exitIfError("Can't connect to deviceinfo service", err)
		defer deviceInfo.Close()

		info, err := deviceInfo.GetDisplayInfo()
		exitIfError("Can't fetch dispaly info", err)

		fmt.Println(convertToJSONString(info))
		return
	}
	printDeviceInfo(ctx.Device)
}

func runScreenshotCommand(ctx commandContext) {
	stream, _ := ctx.Args.Bool("--stream")
	port, _ := ctx.Args.String("--port")
	path, _ := ctx.Args.String("--output")
	if stream {
		if port == "" {
			port = "3333"
		}
		err := instruments.StartMJPEGStreamingServer(ctx.Device, port)
		exitIfError("failed starting mjpeg", err)
		return
	}
	saveScreenshot(ctx.Device, path)
}

func runDateCommand(ctx commandContext) {
	printDeviceDate(ctx.Device)
}

func runDiagnosticsCommand(ctx commandContext) {
	printDiagnostics(ctx.Device)
}

func runBatteryRegistryCommand(ctx commandContext) {
	printBatteryRegistry(ctx.Device)
}

func runDiskspaceCommand(ctx commandContext) {
	afcService, err := afc.New(ctx.Device)
	exitIfError("connect afc service failed", err)
	info, err := afcService.DeviceInfo()
	exitIfError("get device info push failed", err)
	if JSONdisabled {
		fmt.Printf("      Model: %s\n", info.Model)
		fmt.Printf("  BlockSize: %d\n", info.BlockSize)
		fmt.Printf("  FreeSpace: %s\n", diskspaceByteCount(info.FreeBytes))
		fmt.Printf("  UsedSpace: %s\n", diskspaceByteCount(usedDiskBytes(info.TotalBytes, info.FreeBytes)))
		fmt.Printf(" TotalSpace: %s\n", diskspaceByteCount(info.TotalBytes))
	} else {
		fmt.Println(convertToJSONString(info))
	}
}

func diskspaceByteCount(bytes uint64) string {
	const unit = uint64(1000)
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}
	div, exp := unit, 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%cB", float64(bytes)/float64(div), "kMGTPE"[exp])
}

func usedDiskBytes(totalBytes uint64, freeBytes uint64) uint64 {
	if freeBytes > totalBytes {
		exitIfError("diskspace: invalid device info", fmt.Errorf("free bytes %d exceeds total bytes %d", freeBytes, totalBytes))
	}
	return totalBytes - freeBytes
}

func runBatteryCheckCommand(ctx commandContext) {
	printBatteryDiagnostics(ctx.Device)
}

func runRSDCommand(ctx commandContext) {
	listCommand, _ := ctx.Args.Bool("ls")
	if listCommand {
		services := ctx.Device.Rsd.GetServices()
		if JSONdisabled {
			fmt.Println(services)
		} else {
			b, err := marshalJSON(services)
			exitIfError("failed json conversion", err)
			fmt.Println(string(b))
		}
	}
}

func runMobileGestaltCommand(ctx commandContext) {
	conn, _ := diagnostics.New(ctx.Device)
	keys := ctx.Args["<key>"].([]string)
	plist, _ := ctx.Args.Bool("--plist")
	resp, _ := conn.MobileGestaltQuery(keys)
	if plist {
		fmt.Printf("%s\n", ios.ToPlist(resp))
		return
	}
	jb, _ := marshalJSON(resp)
	fmt.Printf("%s\n", jb)
}

func runDeviceStateCommand(ctx commandContext) {
	listCommand, _ := ctx.Args.Bool("list")
	if listCommand {
		deviceState(ctx.Device, true, false, "", "")
		return
	}
	enable, _ := ctx.Args.Bool("enable")
	profileTypeId, _ := ctx.Args.String("<profileTypeId>")
	profileId, _ := ctx.Args.String("<profileId>")
	deviceState(ctx.Device, false, enable, profileTypeId, profileId)
}

func runLockdownCommand(ctx commandContext) {
	if get, _ := ctx.Args.Bool("get"); !get {
		return
	}
	key := ""
	if keyArg := ctx.Args["<key>"]; keyArg != nil {
		if keys, ok := keyArg.([]string); ok && len(keys) > 0 {
			key = keys[0]
		}
	}
	domain, _ := ctx.Args.String("--domain")

	lockdownConnection, err := ios.ConnectLockdownWithSession(ctx.Device)
	exitIfError("failed connecting to lockdown", err)
	defer lockdownConnection.Close()

	if key == "" && domain == "" {
		allValues, err := lockdownConnection.GetValues()
		exitIfError("failed getting lockdown values", err)
		fmt.Println(convertToJSONString(allValues.Value))
	} else if domain != "" {
		value, err := lockdownConnection.GetValueForDomain(key, domain)
		exitIfError(fmt.Sprintf("failed getting value from domain '%s'", domain), err)
		fmt.Println(convertToJSONString(value))
	} else {
		value, err := lockdownConnection.GetValue(key)
		exitIfError(fmt.Sprintf("failed getting lockdown value '%s'", key), err)
		fmt.Println(convertToJSONString(value))
	}
}

func runSysmontapCommand(ctx commandContext) {
	printSysmontapStats(ctx.Device)
}
