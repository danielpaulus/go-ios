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

// parseCLIArgs parses a real command line against the CLI's actual docopt
// usage string, exactly as Main does.
func parseCLIArgs(t *testing.T, argv ...string) docopt.Opts {
	t.Helper()
	parser := &docopt.Parser{HelpHandler: docopt.NoHelpHandler}
	args, err := parser.ParseArgs(cliUsage(), argv, version)
	if err != nil {
		t.Fatalf("failed parsing argv %v: %v", argv, err)
	}
	return args
}

// firstMatchingCommand mirrors dispatchCommand's first-match-wins selection
// without running the command.
func firstMatchingCommand(args docopt.Opts, commands []command) string {
	for _, cmd := range commands {
		if cmd.match(args) {
			return cmd.name
		}
	}
	return ""
}

// dispatchedCommand walks the dispatch chain in the same order as Main
// (globalCommands, then deviceCommands) and returns the name of the command
// that would run.
func dispatchedCommand(t *testing.T, argv ...string) string {
	t.Helper()
	args := parseCLIArgs(t, argv...)
	if name := firstMatchingCommand(args, globalCommands); name != "" {
		return "global:" + name
	}
	if name := firstMatchingCommand(args, deviceCommands); name != "" {
		return "device:" + name
	}
	return ""
}

// TestDispatchSubcommandCollisions guards every subcommand literal that is also
// a top-level command name. docopt sets a boolean for each literal it matched,
// so e.g. `ios webinspector launch <url>` sets both "webinspector" and
// "launch"; the registries must not let the wrong top-level command win
// (issue #769).
func TestDispatchSubcommandCollisions(t *testing.T) {
	testCases := []struct {
		name string
		argv []string
		want string
	}{
		// webinspector vs launch (issue #769)
		{name: "webinspector launch dispatches webinspector", argv: []string{"webinspector", "launch", "https://example.com"}, want: "device:webinspector"},
		{name: "webinspector launch with bundle-id dispatches webinspector", argv: []string{"webinspector", "launch", "https://example.com", "--bundle-id=com.apple.mobilesafari"}, want: "device:webinspector"},
		{name: "plain launch dispatches launch", argv: []string{"launch", "com.apple.mobilesafari"}, want: "device:launch"},
		// info vs lockdown
		{name: "info lockdown dispatches info", argv: []string{"info", "lockdown"}, want: "device:info"},
		{name: "plain lockdown dispatches lockdown", argv: []string{"lockdown", "get", "ProductVersion"}, want: "device:lockdown"},
		// ui vs launch/screenshot/install/run
		{name: "ui app launch dispatches ui", argv: []string{"ui", "app", "launch", "com.apple.mobilesafari"}, want: "global:ui"},
		{name: "ui screenshot dispatches ui", argv: []string{"ui", "screenshot"}, want: "global:ui"},
		{name: "plain screenshot dispatches screenshot", argv: []string{"screenshot"}, want: "device:screenshot"},
		{name: "ui install dispatches ui install", argv: []string{"ui", "install", "wda", "--p12file=cert.p12", "--profile=dev.mobileprovision"}, want: "device:ui install"},
		{name: "plain install dispatches install", argv: []string{"install", "--path=app.ipa"}, want: "device:install"},
		{name: "ui run dispatches ui run", argv: []string{"ui", "run", "wda"}, want: "device:ui run"},
		// webinspector vs list (also guarded by TestDeviceListCommandOnlyMatchesTopLevelList)
		{name: "webinspector list dispatches webinspector", argv: []string{"webinspector", "list"}, want: "device:webinspector"},
		{name: "plain list dispatches list", argv: []string{"list"}, want: "global:list"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := dispatchedCommand(t, testCase.argv...); got != testCase.want {
				t.Fatalf("dispatchedCommand(%v) = %q, want %q", testCase.argv, got, testCase.want)
			}
		})
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
