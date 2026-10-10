package main

import (
	"strconv"

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

// intArgDefault reads a numeric docopt flag, returning def when the flag is
// absent, empty, or not a valid integer.
func intArgDefault(args docopt.Opts, name string, def int) int {
	s, _ := args.String(name)
	if s == "" {
		return def
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return v
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
