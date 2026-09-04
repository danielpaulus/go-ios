package main

import "github.com/docopt/docopt-go"

var deviceCommands = []command{
	commandByBool("activate", runActivateCommand),
	commandByBool("ip", runIPCommand),
	commandByBool("pcap", runPCAPCommand),
	commandByBool("ps", runPSCommand),
	{
		name: "install",
		match: func(args docopt.Opts) bool {
			return boolArg(args, "install") && !boolArg(args, "ui")
		},
		run: runInstallCommand,
	},
	commandByBool("sign", runSignCommand),
	{
		name: "ui install",
		match: func(args docopt.Opts) bool {
			return boolArg(args, "ui") && boolArg(args, "install")
		},
		run: runUIInstallCommand,
	},
	{
		name: "ui run",
		match: func(args docopt.Opts) bool {
			return boolArg(args, "ui") && boolArg(args, "run")
		},
		run: runUIRunCommand,
	},
	commandByBool("uninstall", runUninstallCommand),
	commandByBool("lang", runLangCommand),
	commandByBool("dproxy", runDproxyCommand),
	commandByBool("info", runInfoCommand),
	commandByBool("syslog", runSyslogCommand),
	commandByBool("ostrace", runOSTraceCommand),
	{
		// "screenshot" is also a subcommand literal of `ios ui screenshot`
		// (dispatched as a global ui command), so only match the top-level
		// `ios screenshot`.
		name: "screenshot",
		match: func(args docopt.Opts) bool {
			return boolArg(args, "screenshot") && !boolArg(args, "ui")
		},
		run: runScreenshotCommand,
	},
	commandByBool("resetlocation", runResetLocationCommand),
	commandByBool("devicename", runDeviceNameCommand),
	commandByBool("apps", runAppsCommand),
	commandByBool("date", runDateCommand),
	commandByBool("diagnostics", runDiagnosticsCommand),
	commandByBool("pair", runPairCommand),
	commandByBool("readpair", runReadPairCommand),
	commandByBool("batteryregistry", runBatteryRegistryCommand),
	commandByBool("reboot", runRebootCommand),
	commandByBool("shutdown", runShutdownCommand),
	commandByBool("diskspace", runDiskspaceCommand),
	commandByBool("batterycheck", runBatteryCheckCommand),
	commandByBool("erase", runEraseCommand),
	commandByBool("rsd", runRSDCommand),
	commandByBool("mobilegestalt", runMobileGestaltCommand),
	commandByBool("devicestate", runDeviceStateCommand),
	commandByBool("wifi", runWifiCommand),
	commandByBool("prepare", runPrepareCommand),
	commandByBool("set-wallpaper", runSetWallpaperCommand),
	commandByBool("get-wallpaper", runGetWallpaperCommand),
	commandByBool("get-icon-layout", runGetIconLayoutCommand),
	commandByBool("set-icon-layout", runSetIconLayoutCommand),
	commandByBool("crash", runCrashCommand),
	commandByBool("instruments", runInstrumentsCommand),
	commandByBool("image", runImageCommand),
	commandByBool("assistivetouch", runAssistiveTouchCommand),
	commandByBool("voiceover", runVoiceOverCommand),
	commandByBool("zoom", runZoomCommand),
	{
		// "lockdown" is also a subcommand literal of `ios info lockdown`, so only
		// match the top-level `ios lockdown get [<key>]`.
		name: "lockdown",
		match: func(args docopt.Opts) bool {
			return boolArg(args, "lockdown") && !boolArg(args, "info")
		},
		run: runLockdownCommand,
	},
	commandByBool("setlocation", runSetLocationCommand),
	commandByBool("setlocationgpx", runSetLocationGPXCommand),
	commandByBool("timeformat", runTimeFormatCommand),
	commandByBool("httpproxy", runHTTPProxyCommand),
	commandByBool("mdm", runMdmCommand),
	commandByBool("profile", runProfileCommand),
	commandByBool("forward", runForwardCommand),
	{
		// "launch" is also a subcommand literal of `ios webinspector launch <url>`
		// and `ios ui app launch <bundleID>` (the latter is dispatched as a global
		// ui command), so only match the top-level `ios launch <bundleID>`.
		name: "launch",
		match: func(args docopt.Opts) bool {
			return boolArg(args, "launch") && !boolArg(args, "webinspector") && !boolArg(args, "ui")
		},
		run: runLaunchCommand,
	},
	commandByBool("sysmontap", runSysmontapCommand),
	commandByBool("memlimitoff", runMemlimitOffCommand),
	commandByBool("kill", runKillCommand),
	commandByBool("runtest", runTestCommand),
	commandByBool("runxctest", runXCTestCommand),
	commandByBool("runwda", runWDACommand),
	commandByBool("ax", runAXCommand),
	commandByBool("resetax", runResetAXCommand),
	commandByBool("debug", runDebugCommand),
	commandByBool("file", runFileCommand),
	commandByBool("pasteboard", runPasteboardCommand),
	commandByBool("fsync", runFsyncCommand),
	commandByBool("devmode", runDevModeCommand),
	commandByBool("webinspector", runWebInspectorCommand),
	commandByBool("remote", runRemoteCommand),
}
