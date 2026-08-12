package main

import (
	"os"

	"github.com/danielpaulus/go-ios/ios/remote"
)

// runRemoteCommand serves the `ios remote` browser remote-control. The live
// screen is WDA-free (instruments screenshot service); input is routed through
// the WebDriverAgent at --wda-url (default remote.DefaultWDAURL or
// GO_IOS_WDA_URL).
func runRemoteCommand(ctx commandContext) {
	port, _ := ctx.Args.String("--port")
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
