package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios/hid"
	log "github.com/sirupsen/logrus"
)

// Eight samples over 100ms reads as a flick, which is what VoiceOver navigation
// needs. Pass a longer duration for a slow drag.
const (
	defaultDragSteps    = 8
	defaultDragDuration = 100 * time.Millisecond
)

// runHIDCommand dispatches the `ios hid` subcommands. Each single-gesture command
// negotiates its own stream; `ios hid session` runs a batch through one.
func runHIDCommand(ctx commandContext) {
	switch {
	case boolArg(ctx.Args, "services"):
		runHIDListServices(ctx)
	case boolArg(ctx.Args, "tap"):
		runHIDTap(ctx)
	case boolArg(ctx.Args, "drag"):
		runHIDDrag(ctx)
	case boolArg(ctx.Args, "button"):
		runHIDButton(ctx)
	case boolArg(ctx.Args, "type"):
		runHIDType(ctx)
	case boolArg(ctx.Args, "session"):
		runHIDSession(ctx)
	}
}

func runHIDListServices(ctx commandContext) {
	withHIDSession(ctx, func(session *hid.Session) error {
		services, err := session.ListServices()
		if err != nil {
			return err
		}
		fmt.Println(convertToJSONString(services))
		return nil
	})
}

func runHIDTap(ctx commandContext) {
	point := hidPoint(ctx, "<x>", "<y>")
	withHIDSession(ctx, func(session *hid.Session) error {
		if err := session.Tap(context.Background(), point); err != nil {
			return err
		}
		log.WithFields(log.Fields{"x": point.X, "y": point.Y}).Info("tap sent")
		return nil
	})
}

func runHIDDrag(ctx commandContext) {
	from := hidPoint(ctx, "<x>", "<y>")
	to := hidPoint(ctx, "<tox>", "<toy>")
	steps := hidOptionalInt(ctx, "--steps", defaultDragSteps)
	exitIfError("invalid --steps", validateDragSteps(steps))
	duration := hidOptionalDuration(ctx, "--duration", defaultDragDuration)

	withHIDSession(ctx, func(session *hid.Session) error {
		if err := session.Drag(context.Background(), from, to, steps, duration); err != nil {
			return err
		}
		log.WithFields(log.Fields{
			"fromX": from.X, "fromY": from.Y, "toX": to.X, "toY": to.Y,
			"steps": steps, "duration": duration,
		}).Info("drag sent")
		return nil
	})
}

func runHIDButton(ctx commandContext) {
	usagePage, err := ctx.Args.Int("<usagepage>")
	exitIfError("failed parsing <usagepage>", err)
	usageCode, err := ctx.Args.Int("<usagecode>")
	exitIfError("failed parsing <usagecode>", err)

	withHIDSession(ctx, func(session *hid.Session) error {
		if err := session.PressButton(context.Background(), uint64(usagePage), uint64(usageCode)); err != nil {
			return err
		}
		log.WithFields(log.Fields{"usagePage": usagePage, "usageCode": usageCode}).Info("button pressed")
		return nil
	})
}

func runHIDType(ctx commandContext) {
	text, err := ctx.Args.String("<text>")
	exitIfError("failed parsing <text>", err)

	withHIDSession(ctx, func(session *hid.Session) error {
		if err := session.Type(context.Background(), text); err != nil {
			return err
		}
		log.WithFields(log.Fields{"length": len(text)}).Info("text typed")
		return nil
	})
}

// runHIDSession runs a batch of gestures inside a single session, so one media
// stream serves all of them.
func runHIDSession(ctx commandContext) {
	source := os.Stdin
	if path, err := ctx.Args.String("--script"); err == nil && path != "" {
		file, err := os.Open(path)
		exitIfError("failed opening the gesture script", err)
		defer file.Close()
		source = file
	}

	gestures, err := parseHIDGestures(source)
	exitIfError("failed parsing gestures", err)
	if len(gestures) == 0 {
		exitIfError("no gestures to run", fmt.Errorf("the script is empty"))
	}

	withHIDSession(ctx, func(session *hid.Session) error {
		for _, gesture := range gestures {
			if err := gesture.run(context.Background(), session); err != nil {
				return fmt.Errorf("line %d (%s): %w", gesture.line, gesture.op, err)
			}
		}
		log.WithFields(log.Fields{"gestures": len(gestures)}).Info("gesture batch finished")
		return nil
	})
}

// withHIDSession opens a session, runs fn and always closes the session, so the
// media stream is stopped even when the gesture fails.
func withHIDSession(ctx commandContext, fn func(*hid.Session) error) {
	session, err := hid.NewSession(ctx.Device)
	exitIfError("failed connecting to the HID service", err)

	runErr := fn(session)
	closeErr := session.Close()

	exitIfError("HID command failed", runErr)
	if closeErr != nil {
		log.WithFields(log.Fields{"error": closeErr}).Debug("closing the HID session reported an error")
	}
}

// hidGesture is one parsed line of a gesture script.
type hidGesture struct {
	line int
	op   string
	run  func(context.Context, *hid.Session) error
}

// parseHIDGestures reads whitespace-separated gesture lines. Blank lines and
// everything after a '#' are ignored.
//
//	tap   X Y
//	drag   X Y TOX TOY [STEPS [DURATION_SECONDS]]
//	stroke SECONDS X1 Y1 X2 Y2 ...   (one contact through every point)
//	move  X Y
//	type  TEXT...   (whitespace is collapsed and '#' cannot appear in the text)
//	button USAGEPAGE USAGECODE
//	sleep SECONDS
func parseHIDGestures(reader io.Reader) ([]hidGesture, error) {
	var gestures []hidGesture

	scanner := bufio.NewScanner(reader)
	for lineNo := 1; scanner.Scan(); lineNo++ {
		line := scanner.Text()
		if idx := strings.Index(line, "#"); idx >= 0 {
			line = line[:idx]
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		gesture, err := parseHIDGesture(lineNo, fields)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNo, err)
		}
		gestures = append(gestures, gesture)
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading gestures: %w", err)
	}
	return gestures, nil
}

func parseHIDGesture(lineNo int, fields []string) (hidGesture, error) {
	op := fields[0]
	args := fields[1:]
	gesture := hidGesture{line: lineNo, op: op}

	switch op {
	case "tap":
		point, err := parsePoint(args, 2)
		if err != nil {
			return gesture, fmt.Errorf("tap wants X Y: %w", err)
		}
		gesture.run = func(ctx context.Context, s *hid.Session) error { return s.Tap(ctx, point) }
	case "drag":
		if len(args) < 4 {
			return gesture, fmt.Errorf("drag wants X Y TOX TOY [STEPS [DURATION_SECONDS]]")
		}
		from, err := parsePoint(args[:2], 2)
		if err != nil {
			return gesture, err
		}
		to, err := parsePoint(args[2:4], 2)
		if err != nil {
			return gesture, err
		}
		steps := defaultDragSteps
		if len(args) > 4 {
			if steps, err = strconv.Atoi(args[4]); err != nil {
				return gesture, fmt.Errorf("invalid STEPS %q: %w", args[4], err)
			}
			if err := validateDragSteps(steps); err != nil {
				return gesture, err
			}
		}
		duration := defaultDragDuration
		if len(args) > 5 {
			seconds, err := strconv.ParseFloat(args[5], 64)
			if err != nil {
				return gesture, fmt.Errorf("invalid DURATION %q: %w", args[5], err)
			}
			duration = time.Duration(seconds * float64(time.Second))
		}
		gesture.run = func(ctx context.Context, s *hid.Session) error {
			return s.Drag(ctx, from, to, steps, duration)
		}
	case "stroke":
		// stroke SECONDS X1 Y1 X2 Y2 ... - the duration comes first so the
		// point list can be any length.
		if len(args) < 3 || len(args)%2 == 0 {
			return gesture, fmt.Errorf("stroke wants SECONDS then at least one X Y pair")
		}
		seconds, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return gesture, fmt.Errorf("invalid SECONDS %q: %w", args[0], err)
		}
		coords := args[1:]
		points := make([]hid.Point, 0, len(coords)/2)
		for i := 0; i < len(coords); i += 2 {
			point, err := parsePoint(coords[i:i+2], 2)
			if err != nil {
				return gesture, err
			}
			points = append(points, point)
		}
		duration := time.Duration(seconds * float64(time.Second))
		gesture.run = func(ctx context.Context, s *hid.Session) error {
			return s.Stroke(ctx, points, duration)
		}
	case "move":
		if len(args) != 2 {
			return gesture, fmt.Errorf("move wants X Y")
		}
		x, err := strconv.ParseInt(args[0], 10, 32)
		if err != nil {
			return gesture, fmt.Errorf("invalid X %q: %w", args[0], err)
		}
		y, err := strconv.ParseInt(args[1], 10, 32)
		if err != nil {
			return gesture, fmt.Errorf("invalid Y %q: %w", args[1], err)
		}
		gesture.run = func(ctx context.Context, s *hid.Session) error {
			return s.MoveDigitizer(ctx, int32(x), int32(y))
		}
	case "type":
		if len(args) == 0 {
			return gesture, fmt.Errorf("type wants TEXT")
		}
		text := strings.Join(args, " ")
		gesture.run = func(ctx context.Context, s *hid.Session) error { return s.Type(ctx, text) }
	case "button":
		if len(args) != 2 {
			return gesture, fmt.Errorf("button wants USAGEPAGE USAGECODE")
		}
		page, err := strconv.ParseUint(args[0], 0, 64)
		if err != nil {
			return gesture, fmt.Errorf("invalid USAGEPAGE %q: %w", args[0], err)
		}
		code, err := strconv.ParseUint(args[1], 0, 64)
		if err != nil {
			return gesture, fmt.Errorf("invalid USAGECODE %q: %w", args[1], err)
		}
		gesture.run = func(ctx context.Context, s *hid.Session) error {
			return s.PressButton(ctx, page, code)
		}
	case "sleep":
		if len(args) != 1 {
			return gesture, fmt.Errorf("sleep wants SECONDS")
		}
		seconds, err := strconv.ParseFloat(args[0], 64)
		if err != nil {
			return gesture, fmt.Errorf("invalid SECONDS %q: %w", args[0], err)
		}
		delay := time.Duration(seconds * float64(time.Second))
		gesture.run = func(ctx context.Context, _ *hid.Session) error {
			select {
			case <-time.After(delay):
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	default:
		return gesture, fmt.Errorf("unknown gesture %q", op)
	}
	return gesture, nil
}

// maxDragSteps caps the samples one drag may send. The limit is arbitrary but
// generous: it exists so a typo cannot flood the device with reports.
const maxDragSteps = 1000

func validateDragSteps(steps int) error {
	if steps < 1 {
		return fmt.Errorf("a drag needs at least 1 step, got %d", steps)
	}
	if steps > maxDragSteps {
		return fmt.Errorf("a drag is limited to %d steps, got %d", maxDragSteps, steps)
	}
	return nil
}

func parsePoint(args []string, want int) (hid.Point, error) {
	if len(args) != want {
		return hid.Point{}, fmt.Errorf("expected %d coordinates, got %d", want, len(args))
	}
	x, err := parseNormalisedCoordinate(args[0])
	if err != nil {
		return hid.Point{}, err
	}
	y, err := parseNormalisedCoordinate(args[1])
	if err != nil {
		return hid.Point{}, err
	}
	return hid.Point{X: x, Y: y}, nil
}

func parseNormalisedCoordinate(arg string) (uint16, error) {
	value, err := strconv.ParseUint(arg, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("invalid coordinate %q: %w", arg, err)
	}
	if value > 0xFFFF {
		return 0, fmt.Errorf("coordinate %d is out of the 0..65535 range", value)
	}
	return uint16(value), nil
}

// hidPoint reads a pair of docopt positional coordinates. Coordinates are
// normalised across the display: 0 is the left/top edge, 65535 the right/bottom.
func hidPoint(ctx commandContext, xArg, yArg string) hid.Point {
	x, err := ctx.Args.Int(xArg)
	exitIfError("failed parsing "+xArg, err)
	y, err := ctx.Args.Int(yArg)
	exitIfError("failed parsing "+yArg, err)
	if x < 0 || x > 0xFFFF || y < 0 || y > 0xFFFF {
		exitIfError("coordinates must be within 0..65535",
			fmt.Errorf("got %s=%d %s=%d", xArg, x, yArg, y))
	}
	return hid.Point{X: uint16(x), Y: uint16(y)}
}

func hidOptionalInt(ctx commandContext, name string, fallback int) int {
	raw, err := ctx.Args.String(name)
	if err != nil || raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	exitIfError("failed parsing "+name, err)
	return value
}

func hidOptionalDuration(ctx commandContext, name string, fallback time.Duration) time.Duration {
	raw, err := ctx.Args.String(name)
	if err != nil || raw == "" {
		return fallback
	}
	seconds, err := strconv.ParseFloat(raw, 64)
	exitIfError("failed parsing "+name, err)
	return time.Duration(seconds * float64(time.Second))
}
