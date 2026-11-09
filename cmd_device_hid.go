package main

import (
	"fmt"
	"strconv"

	"github.com/danielpaulus/go-ios/ios/hid"
	"github.com/docopt/docopt-go"
)

func runHIDCommand(ctx commandContext) {
	controls, err := hid.NewControls(ctx.Device)
	exitIfError("failed connecting to the hid service (iOS 27+ with the developer disk image mounted)", err)
	defer controls.Close()

	if boolArg(ctx.Args, "hinge") {
		degrees, err := hingeAngle(ctx.Args)
		exitIfError("invalid hinge angle", err)
		exitIfError("failed setting the hinge angle", controls.SetHingeAngle(degrees))
	}
}

// hingeAngle is the angle a preset stands for, or the number of degrees given.
func hingeAngle(args docopt.Opts) (float64, error) {
	switch {
	case boolArg(args, "close"):
		return hid.HingeClosed, nil
	case boolArg(args, "halfopen"):
		return hid.HingeHalfOpen, nil
	case boolArg(args, "open"):
		return hid.HingeOpenFlat, nil
	}
	value, _ := args.String("<degrees>")
	degrees, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("%q is not close, halfopen, open or a number of degrees", value)
	}
	return degrees, nil
}
