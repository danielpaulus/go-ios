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
}

func runListenCommand(ctx commandContext) {
	startListening()
}

func isDeviceListCommand(args docopt.Opts) bool {
	listCommand := boolArg(args, "list")
	diagnosticsCommand := boolArg(args, "diagnostics")
	imageCommand := boolArg(args, "image")
	deviceStateCommand := boolArg(args, "devicestate")
	profileCommand := boolArg(args, "profile")
	return listCommand && !diagnosticsCommand && !imageCommand && !deviceStateCommand && !profileCommand
}

func runDeviceListCommand(ctx commandContext) {
	details, _ := ctx.Args.Bool("--details")
	printDeviceList(details)
}
