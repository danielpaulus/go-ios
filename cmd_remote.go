package main

import (
	"os"

	"github.com/danielpaulus/go-ios/ios/remote"
	"github.com/docopt/docopt-go"
)

// runRemoteCommand serves the `ios remote` browser remote-control. The live
// screen is WDA-free (instruments screenshot service); input is routed through
// the WebDriverAgent at --wda-url (default remote.DefaultWDAURL or
// GO_IOS_WDA_URL).
func runRemoteCommand(ctx commandContext) {
	port := portArg(ctx.Args)
	if port == "" {
		port = "8080"
	}

	wdaURL, _ := ctx.Args.String("--wda-url")
	if wdaURL == "" {
		wdaURL = os.Getenv("GO_IOS_WDA_URL")
	}
	if wdaURL == "" {
		wdaURL = remote.DefaultWDAURL
	}

	server, err := remote.NewServer(ctx.Device, wdaURL)
	exitIfError("failed starting remote server (developer disk image mounted?)", err)
	defer server.Close()

	exitIfError("remote server stopped", server.ListenAndServe(port))
}

// portArg reads --port. The global usage declares `ios forward … [--port=<mapping>]…`
// as repeatable, so docopt surfaces --port as a []string for every command; a
// plain args.String("--port") therefore returns "". Handle both shapes so
// `ios remote --port=8090` works.
func portArg(args docopt.Opts) string {
	if list, ok := args["--port"].([]string); ok {
		if len(list) > 0 {
			return list[len(list)-1]
		}
		return ""
	}
	s, _ := args.String("--port")
	return s
}
