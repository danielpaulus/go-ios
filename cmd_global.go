package main

import "github.com/docopt/docopt-go"

var preProxyCommands = []command{
	{
		name: "version",
		match: func(args docopt.Opts) bool {
			return boolArg(args, "version") || boolArg(args, "--version")
		},
		run: func(ctx commandContext) {
			printVersion()
		},
	},
}

var globalCommands = []command{
	{
		name: "ui",
		match: func(args docopt.Opts) bool {
			return boolArg(args, "ui") && !boolArg(args, "install")
		},
		run: runUICommand,
	},
	commandByBool("listen", runListenCommand),
	{
		name:  "list",
		match: isDeviceListCommand,
		run:   runDeviceListCommand,
	},
	{
		// `sign certificate appstoreconnect` mints an account-wide certificate and
		// needs no device, so it is global (dispatched before device resolution).
		name: "sign certificate",
		match: func(args docopt.Opts) bool {
			return boolArg(args, "sign") && boolArg(args, "certificate")
		},
		run: runSignCertificateAppStoreConnectCommand,
	},
}

func runListenCommand(ctx commandContext) {
	startListening()
}

// isDeviceListCommand matches the bare global `ios list`. globalCommands are
// dispatched before deviceCommands, so every device subcommand that also has a
// `list` literal (`ios <cmd> list`) sets the same "list" arg and would otherwise
// be swallowed here. Each such command must be excluded below;
// TestDeviceListCommandOnlyMatchesTopLevelList guards that every `<cmd> list`
// subcommand is excluded rather than falling through to the device list.
func isDeviceListCommand(args docopt.Opts) bool {
	listCommand := boolArg(args, "list")
	diagnosticsCommand := boolArg(args, "diagnostics")
	imageCommand := boolArg(args, "image")
	deviceStateCommand := boolArg(args, "devicestate")
	profileCommand := boolArg(args, "profile")
	webInspectorCommand := boolArg(args, "webinspector")
	return listCommand && !diagnosticsCommand && !imageCommand && !deviceStateCommand && !profileCommand && !webInspectorCommand
}

func runDeviceListCommand(ctx commandContext) {
	details, _ := ctx.Args.Bool("--details")
	printDeviceList(details)
}
