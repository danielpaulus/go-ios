package main

import (
	"testing"

	"github.com/docopt/docopt-go"
)

func TestDispatchCommandRunsFirstMatch(t *testing.T) {
	args := docopt.Opts{"alpha": true, "beta": true}
	var ran []string

	handled := dispatchCommand(commandContext{Args: args}, []command{
		commandByBool("alpha", func(commandContext) { ran = append(ran, "alpha") }),
		commandByBool("beta", func(commandContext) { ran = append(ran, "beta") }),
	})

	if !handled {
		t.Fatal("dispatchCommand returned false")
	}
	if len(ran) != 1 || ran[0] != "alpha" {
		t.Fatalf("ran = %#v, want only alpha", ran)
	}
}

func TestDispatchCommandReturnsFalseWithoutMatch(t *testing.T) {
	args := docopt.Opts{"alpha": false}
	handled := dispatchCommand(commandContext{Args: args}, []command{
		commandByBool("alpha", func(commandContext) {
			t.Fatal("handler should not run")
		}),
	})

	if handled {
		t.Fatal("dispatchCommand returned true")
	}
}

func TestDeviceListCommandOnlyMatchesTopLevelList(t *testing.T) {
	if !isDeviceListCommand(docopt.Opts{"list": true}) {
		t.Fatal("top-level list command did not match")
	}

	for _, commandName := range []string{"diagnostics", "image", "devicestate", "profile", "webinspector"} {
		args := docopt.Opts{"list": true, commandName: true}
		if isDeviceListCommand(args) {
			t.Fatalf("list subcommand for %s matched top-level list", commandName)
		}
	}
}

func TestTunnelCommandMatcher(t *testing.T) {
	if !isTunnelCommand(docopt.Opts{"tunnel": true}) {
		t.Fatal("tunnel command did not match")
	}
	if isTunnelCommand(docopt.Opts{"tunnel": false}) {
		t.Fatal("non-tunnel command matched")
	}
}
