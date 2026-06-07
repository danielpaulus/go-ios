package main

import (
	"github.com/danielpaulus/go-ios/ios/instruments"
)

func runResetLocationCommand(ctx commandContext) {
	resetLocation(ctx.Device)
}

func runSetLocationCommand(ctx commandContext) {
	lat, _ := ctx.Args.String("--lat")
	lon, _ := ctx.Args.String("--lon")

	if ctx.Device.SupportsRsd() {
		server, err := instruments.NewLocationSimulationService(ctx.Device)
		exitIfError("failed to create location simulation service:", err)

		startLocationSimulation(server, lat, lon)
		return
	}

	setLocation(ctx.Device, lat, lon)
}

func runSetLocationGPXCommand(ctx commandContext) {
	gpxFilePath, _ := ctx.Args.String("--gpxfilepath")
	setLocationGPX(ctx.Device, gpxFilePath)
}
