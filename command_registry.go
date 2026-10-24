package main

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/docopt/docopt-go"
)

type commandContext struct {
	Args   docopt.Opts
	Device ios.DeviceEntry
}

type command struct {
	name  string
	match func(docopt.Opts) bool
	run   func(commandContext)
}

func boolArg(args docopt.Opts, name string) bool {
	value, _ := args.Bool(name)
	return value
}

// portArg reads --port for commands that take a single port. `ios forward`
// declares --port as repeatable, so docopt always hands it over as []string
// and Args.String/Args.Int fail. The last value wins.
func portArg(args docopt.Opts) string {
	switch value := args["--port"].(type) {
	case string:
		return value
	case []string:
		if len(value) > 0 {
			return value[len(value)-1]
		}
	}
	return ""
}

func commandByBool(name string, run func(commandContext)) command {
	return command{
		name: name,
		match: func(args docopt.Opts) bool {
			return boolArg(args, name)
		},
		run: run,
	}
}

func dispatchCommand(ctx commandContext, commands []command) bool {
	for _, cmd := range commands {
		if cmd.match(ctx.Args) {
			cmd.run(ctx)
			return true
		}
	}
	return false
}
