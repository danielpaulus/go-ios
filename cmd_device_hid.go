package main

import (
	"fmt"
	"strconv"
	"time"

	"github.com/danielpaulus/go-ios/ios/hid"
	"github.com/docopt/docopt-go"
)

func runHIDCommand(ctx commandContext) {
	controls, err := hid.NewControls(ctx.Device)
	exitIfError("failed connecting to the hid service (iOS 27+ with the developer disk image mounted)", err)
	defer controls.Close()

	switch {
	case boolArg(ctx.Args, "button"):
		hold, err := buttonHold(ctx.Args)
		exitIfError("invalid --hold", err)
		exitIfError("failed pressing the button", controls.PressButton(pressedButton(ctx.Args), hold))
	case boolArg(ctx.Args, "hinge"):
		degrees, err := hingeAngle(ctx.Args)
		exitIfError("invalid hinge angle", err)
		exitIfError("failed setting the hinge angle", controls.SetHingeAngle(degrees))
	}
}

// pressedButton is the button named on the command line. docopt only lets one
// of them through.
func pressedButton(args docopt.Opts) hid.Button {
	for _, b := range hid.Buttons() {
		if boolArg(args, b.String()) {
			return b
		}
	}
	return hid.ButtonHome
}

// buttonHold is --hold in milliseconds, or zero for the button's own hold.
func buttonHold(args docopt.Opts) (time.Duration, error) {
	value, _ := args.String("--hold")
	if value == "" {
		return 0, nil
	}
	ms, err := strconv.Atoi(value)
	if err != nil || ms <= 0 {
		return 0, fmt.Errorf("%q is not a positive number of milliseconds", value)
	}
	return time.Duration(ms) * time.Millisecond, nil
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
