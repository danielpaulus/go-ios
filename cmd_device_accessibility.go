package main

import (
	"github.com/danielpaulus/go-ios/ios"
)

func runAssistiveTouchCommand(ctx commandContext) {
	runAccessibilityToggle(ctx, assistiveTouch)
}

func runVoiceOverCommand(ctx commandContext) {
	runAccessibilityToggle(ctx, voiceOver)
}

func runZoomCommand(ctx commandContext) {
	runAccessibilityToggle(ctx, zoomTouch)
}

func runAccessibilityToggle(ctx commandContext, run func(ios.DeviceEntry, string, bool)) {
	force, _ := ctx.Args.Bool("--force")
	for _, operation := range []string{"enable", "disable", "toggle", "get"} {
		if enabled, _ := ctx.Args.Bool(operation); enabled {
			run(ctx.Device, operation, force)
		}
	}
}

func runTimeFormatCommand(ctx commandContext) {
	force, _ := ctx.Args.Bool("--force")
	for _, operation := range []string{"24h", "12h", "toggle", "get"} {
		if enabled, _ := ctx.Args.Bool(operation); enabled {
			timeFormat(ctx.Device, operation, force)
		}
	}
}

func runAXCommand(ctx commandContext) {
	startAx(ctx.Device, ctx.Args)
}

func runResetAXCommand(ctx commandContext) {
	resetAx(ctx.Device)
}
