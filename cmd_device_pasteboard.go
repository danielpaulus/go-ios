package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/danielpaulus/go-ios/ios/pasteboard"
)

func runPasteboardCommand(ctx commandContext) {
	if !ctx.Device.SupportsRsd() {
		exitIfError("pasteboard command requires iOS 17+ with tunnel", fmt.Errorf("tunnel not running. Start with: ios tunnel start"))
	}

	conn, err := pasteboard.New(ctx.Device)
	exitIfError("pasteboard: failed to connect to pasteboard service", err)
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			slog.Error("Failed to close pasteboard connection", "error", closeErr)
		}
	}()

	if set, _ := ctx.Args.Bool("set"); set {
		var text string
		if ctx.Args["<text>"] != nil {
			text, _ = ctx.Args.String("<text>")
		} else {
			// <text> omitted entirely: read the payload from stdin. An explicit
			// empty argument (`set ""`) must set empty text, not block on stdin.
			data, err := io.ReadAll(os.Stdin)
			exitIfError("pasteboard set: failed to read stdin", err)
			text = string(data)
		}
		err := conn.SetText(text)
		exitIfError("pasteboard set: failed to set pasteboard", err)
		return
	}

	if get, _ := ctx.Args.Bool("get"); get {
		text, ok, err := conn.GetText()
		exitIfError("pasteboard get: failed to read pasteboard", err)
		if !ok {
			return
		}
		fmt.Println(text)
	}
}
