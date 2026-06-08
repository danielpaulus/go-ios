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

func TestNeedsAutomaticTunnelInfo(t *testing.T) {
	testCases := []struct {
		name string
		args docopt.Opts
		want bool
	}{
		{name: "zoom stays tunnel-free", args: docopt.Opts{"zoom": true}, want: false},
		{name: "voiceover stays tunnel-free", args: docopt.Opts{"voiceover": true}, want: false},
		{name: "assistivetouch stays tunnel-free", args: docopt.Opts{"assistivetouch": true}, want: false},
		{name: "timeformat stays tunnel-free", args: docopt.Opts{"timeformat": true}, want: false},
		{name: "file needs tunnel", args: docopt.Opts{"file": true}, want: true},
		{name: "rsd needs tunnel", args: docopt.Opts{"rsd": true}, want: true},
		{name: "display info needs tunnel", args: docopt.Opts{"info": true, "display": true}, want: true},
		{name: "plain info stays tunnel-free", args: docopt.Opts{"info": true}, want: false},
		{name: "syslog needs tunnel when available", args: docopt.Opts{"syslog": true}, want: true},
		{name: "runtest needs tunnel on iOS 17", args: docopt.Opts{"runtest": true}, want: true},
		{name: "devicestate needs tunnel (instruments)", args: docopt.Opts{"devicestate": true}, want: true},
		{name: "resetlocation needs tunnel (instruments)", args: docopt.Opts{"resetlocation": true}, want: true},
		{name: "setlocationgpx needs tunnel (instruments)", args: docopt.Opts{"setlocationgpx": true}, want: true},
		{name: "ui run needs tunnel (testmanagerd)", args: docopt.Opts{"ui": true, "run": true}, want: true},
		{name: "ui status stays tunnel-free", args: docopt.Opts{"ui": true, "status": true}, want: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := needsAutomaticTunnelInfo(testCase.args); got != testCase.want {
				t.Fatalf("needsAutomaticTunnelInfo() = %t, want %t", got, testCase.want)
			}
		})
	}
}
