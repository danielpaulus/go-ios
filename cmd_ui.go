package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios/uidriver"
	"github.com/docopt/docopt-go"
)

const (
	defaultUIDriver     = "devicekit"
	defaultWDAURL       = uidriver.DefaultWDAURL
	defaultDeviceKitURL = uidriver.DefaultDeviceKitURL
	uiDriverWDA         = "wda"
	uiDriverDeviceKit   = "devicekit"
	uiDriverAuto        = "auto"
)

// uiClient wires the CLI arguments to a reusable uidriver.Driver. Argument
// parsing, backend resolution and output formatting live here; all HTTP work
// is delegated to the ios/uidriver package.
type uiClient struct {
	driver       string
	wdaURL       string
	deviceKitURL string
	sessionID    string
}

func runUICommand(ctx commandContext) {
	if boolArg(ctx.Args, "download") {
		runUIDownloadCommand(ctx)
		return
	}

	client := newUIClient(ctx.Args)

	switch {
	case boolArg(ctx.Args, "status"):
		printUIResponse(client.driverOrExit().Status())
	case boolArg(ctx.Args, "api") || boolArg(ctx.Args, "raw"):
		client.api(ctx)
	case boolArg(ctx.Args, "tap"):
		printUIResponse(client.driverOrExit().Tap(requiredIntArg(ctx.Args, "--x"), requiredIntArg(ctx.Args, "--y")))
	case boolArg(ctx.Args, "swipe"):
		printUIResponse(client.driverOrExit().Swipe(
			requiredIntArg(ctx.Args, "--from-x"),
			requiredIntArg(ctx.Args, "--from-y"),
			requiredIntArg(ctx.Args, "--to-x"),
			requiredIntArg(ctx.Args, "--to-y"),
			optionalFloatArg(ctx.Args, "--duration", 0),
		))
	case boolArg(ctx.Args, "longpress"):
		printUIResponse(client.driverOrExit().LongPress(requiredIntArg(ctx.Args, "--x"), requiredIntArg(ctx.Args, "--y"), optionalFloatArg(ctx.Args, "--duration", 1)))
	case boolArg(ctx.Args, "type"):
		printUIResponse(client.driverOrExit().Type(requiredStringArg(ctx.Args, "--text")))
	case boolArg(ctx.Args, "button"):
		button, _ := ctx.Args.String("<button>")
		printUIResponse(client.driverOrExit().PressButton(button))
	case boolArg(ctx.Args, "screenshot"):
		output, _ := ctx.Args.String("--output")
		client.screenshot(output)
	case boolArg(ctx.Args, "source"):
		output, _ := ctx.Args.String("--output")
		client.source(output)
	case boolArg(ctx.Args, "size"):
		printUIResponse(client.driverOrExit().WindowSize())
	case boolArg(ctx.Args, "orientation"):
		client.orientation(ctx)
	case boolArg(ctx.Args, "app"):
		client.app(ctx)
	case boolArg(ctx.Args, "stream"):
		client.stream(ctx)
	default:
		logFatal("unknown ui command")
	}
}

func newUIClient(args docopt.Opts) uiClient {
	driver, _ := args.String("--driver")
	if driver == "" {
		driver = os.Getenv("GO_IOS_UI_DRIVER")
	}
	if driver == "" {
		driver = defaultUIDriver
	}
	wdaURL, _ := args.String("--wda-url")
	if wdaURL == "" {
		wdaURL = os.Getenv("GO_IOS_WDA_URL")
	}
	if wdaURL == "" {
		wdaURL = defaultWDAURL
	}
	deviceKitURL, _ := args.String("--devicekit-url")
	if deviceKitURL == "" {
		deviceKitURL = os.Getenv("GO_IOS_DEVICEKIT_URL")
	}
	if deviceKitURL == "" {
		deviceKitURL = defaultDeviceKitURL
	}
	sessionID, _ := args.String("--session-id")
	return uiClient{
		driver:       driver,
		wdaURL:       strings.TrimRight(wdaURL, "/"),
		deviceKitURL: strings.TrimRight(deviceKitURL, "/"),
		sessionID:    sessionID,
	}
}

// resolveBackend maps the CLI --driver value (wda/devicekit/auto) to a concrete
// uidriver.Backend and its base URL, probing health for the auto mode.
func (c uiClient) resolveBackend() (uidriver.Backend, string) {
	switch c.driver {
	case uiDriverWDA:
		return uidriver.BackendWDA, c.wdaURL
	case uiDriverDeviceKit:
		return uidriver.BackendDeviceKit, c.deviceKitURL
	case uiDriverAuto:
		if d, err := uidriver.New(uidriver.BackendDeviceKit, c.deviceKitURL); err == nil && d.Healthy() {
			return uidriver.BackendDeviceKit, c.deviceKitURL
		}
		if d, err := uidriver.New(uidriver.BackendWDA, c.wdaURL); err == nil && d.Healthy() {
			return uidriver.BackendWDA, c.wdaURL
		}
		logFatal("no UI automation backend reachable; start DeviceKit on 127.0.0.1:12004 or WDA on 127.0.0.1:8100, or pass --driver and --*-url")
		return "", ""
	default:
		logFatal("unknown --driver: " + c.driver)
		return "", ""
	}
}

// driverOrExit builds a uidriver.Driver for the resolved backend or exits on error.
func (c uiClient) driverOrExit() *uidriver.Driver {
	backend, baseURL := c.resolveBackend()
	opts := []uidriver.Option{uidriver.WithTimeout(60 * time.Second)}
	if c.sessionID != "" {
		opts = append(opts, uidriver.WithSessionID(c.sessionID))
	}
	d, err := uidriver.New(backend, baseURL, opts...)
	exitIfError("failed creating ui driver", err)
	return d
}

func (c uiClient) api(ctx commandContext) {
	d := c.driverOrExit()
	switch d.Backend() {
	case uidriver.BackendDeviceKit:
		printUIResponse(d.API(uidriver.APIRequest{
			RPCMethod: requiredStringArg(ctx.Args, "--rpc-method"),
			RPCParams: rawParamsFromArgs(ctx),
		}))
	case uidriver.BackendWDA:
		method, _ := ctx.Args.String("--method")
		printUIResponse(d.API(uidriver.APIRequest{
			Method: method,
			Path:   requiredStringArg(ctx.Args, "--http-path"),
			Body:   requestBodyFromArgs(ctx),
		}))
	}
}

func (c uiClient) screenshot(output string) {
	image, err := c.driverOrExit().Screenshot()
	exitIfError("failed capturing screenshot", err)
	if output == "" || output == "-" {
		_, err := os.Stdout.Write(image)
		exitIfError("failed writing screenshot", err)
		return
	}
	exitIfError("failed writing screenshot", os.WriteFile(output, image, 0644))
}

func (c uiClient) source(output string) {
	resp, err := c.driverOrExit().Source()
	writeOrPrintResponse(resp, err, output)
}

func (c uiClient) orientation(ctx commandContext) {
	d := c.driverOrExit()
	if boolArg(ctx.Args, "set") {
		orientation, _ := ctx.Args.String("<orientation>")
		if orientation == "" {
			orientation = requiredStringArg(ctx.Args, "--orientation")
		}
		printUIResponse(d.SetOrientation(orientation))
		return
	}
	printUIResponse(d.Orientation())
}

func (c uiClient) app(ctx commandContext) {
	d := c.driverOrExit()
	switch {
	case boolArg(ctx.Args, "foreground"):
		resp, err := d.AppForeground()
		if err == uidriver.ErrForegroundUnsupported {
			logFatal("app foreground is only available with --driver=devicekit")
		}
		printUIResponse(resp, err)
	default:
		bundleID, _ := ctx.Args.String("<bundleID>")
		if bundleID == "" {
			bundleID = requiredStringArg(ctx.Args, "--bundle-id")
		}
		switch {
		case boolArg(ctx.Args, "launch"):
			printUIResponse(d.AppLaunch(bundleID))
		case boolArg(ctx.Args, "terminate"):
			printUIResponse(d.AppTerminate(bundleID))
		default:
			logFatal("unknown ui app command")
		}
	}
}

func (c uiClient) stream(ctx commandContext) {
	d := c.driverOrExit()
	opts := uidriver.StreamOptions{H264: boolArg(ctx.Args, "h264")}
	opts.FPS, _ = ctx.Args.String("--fps")
	opts.Quality, _ = ctx.Args.String("--quality")
	opts.Scale, _ = ctx.Args.String("--scale")
	opts.Bitrate, _ = ctx.Args.String("--bitrate")
	body, err := d.Stream(context.Background(), opts)
	if err == uidriver.ErrStreamUnsupported {
		logFatal("WDA stream supports mjpeg only; use --driver=devicekit for h264")
	}
	exitIfError("stream request failed", err)
	defer func() { _ = body.Close() }()
	_, err = io.Copy(os.Stdout, body)
	exitIfError("stream copy failed", err)
}

// printUIResponse pretty-prints a driver response, exiting on a non-nil error.
func printUIResponse(resp uidriver.Response, err error) {
	exitIfError("ui request failed", err)
	if len(resp.Body) == 0 {
		return
	}
	var data interface{}
	if uerr := json.Unmarshal(resp.Body, &data); uerr != nil {
		fmt.Print(string(resp.Body))
		return
	}
	fmt.Println(convertToJSONString(data))
}

func writeOrPrintResponse(resp uidriver.Response, err error, output string) {
	exitIfError("ui request failed", err)
	if output == "" || output == "-" {
		printUIResponse(resp, nil)
		return
	}
	exitIfError("failed writing output", os.WriteFile(output, resp.Body, 0644))
}

func requestBodyFromArgs(ctx commandContext) []byte {
	body, _ := ctx.Args.String("--body")
	bodyFile, _ := ctx.Args.String("--body-file")
	if body != "" && bodyFile != "" {
		logFatal("use only one of --body and --body-file")
	}
	if body != "" {
		return []byte(body)
	}
	if bodyFile != "" {
		data, err := os.ReadFile(bodyFile)
		exitIfError("failed reading body file", err)
		return data
	}
	return nil
}

func rawParamsFromArgs(ctx commandContext) interface{} {
	params, _ := ctx.Args.String("--params")
	paramsFile, _ := ctx.Args.String("--params-file")
	if params != "" && paramsFile != "" {
		logFatal("use only one of --params and --params-file")
	}
	if params == "" && paramsFile != "" {
		data, err := os.ReadFile(paramsFile)
		exitIfError("failed reading params file", err)
		params = string(data)
	}
	if params == "" {
		return map[string]interface{}{}
	}
	var decoded interface{}
	exitIfError("failed parsing params JSON", json.Unmarshal([]byte(params), &decoded))
	return decoded
}

func requiredStringArg(args docopt.Opts, name string) string {
	value, _ := args.String(name)
	if value == "" {
		logFatal(name + " is required")
	}
	return value
}

func requiredIntArg(args docopt.Opts, name string) int {
	value, err := args.Int(name)
	if err != nil {
		logFatal(name + " is required")
	}
	return value
}

func optionalFloatArg(args docopt.Opts, name string, fallback float64) float64 {
	value, err := args.String(name)
	if err != nil || value == "" {
		return fallback
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	exitIfError("failed parsing "+name, err)
	return floatValue
}
