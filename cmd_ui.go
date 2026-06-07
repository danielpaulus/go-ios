package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/docopt/docopt-go"
)

const (
	defaultUIDriver      = "devicekit"
	defaultWDAURL        = "http://127.0.0.1:8100"
	defaultDeviceKitURL  = "http://127.0.0.1:12004"
	uiDriverWDA          = "wda"
	uiDriverDeviceKit    = "devicekit"
	uiDriverAuto         = "auto"
	deviceKitRPCProtocol = "2.0"
)

type uiClient struct {
	driver       string
	wdaURL       string
	deviceKitURL string
	httpClient   *http.Client
	sessionID    string
}

type uiHTTPResponse struct {
	StatusCode int
	Header     http.Header
	Body       []byte
}

func runUICommand(ctx commandContext) {
	if boolArg(ctx.Args, "download") {
		runUIDownloadCommand(ctx)
		return
	}

	client := newUIClient(ctx.Args)

	switch {
	case boolArg(ctx.Args, "status"):
		client.printStatus()
	case boolArg(ctx.Args, "api") || boolArg(ctx.Args, "raw"):
		client.api(ctx)
	case boolArg(ctx.Args, "tap"):
		client.tap(requiredIntArg(ctx.Args, "--x"), requiredIntArg(ctx.Args, "--y"))
	case boolArg(ctx.Args, "swipe"):
		client.swipe(
			requiredIntArg(ctx.Args, "--from-x"),
			requiredIntArg(ctx.Args, "--from-y"),
			requiredIntArg(ctx.Args, "--to-x"),
			requiredIntArg(ctx.Args, "--to-y"),
			optionalFloatArg(ctx.Args, "--duration", 0),
		)
	case boolArg(ctx.Args, "longpress"):
		client.longPress(requiredIntArg(ctx.Args, "--x"), requiredIntArg(ctx.Args, "--y"), optionalFloatArg(ctx.Args, "--duration", 1))
	case boolArg(ctx.Args, "type"):
		client.typeText(requiredStringArg(ctx.Args, "--text"))
	case boolArg(ctx.Args, "button"):
		button, _ := ctx.Args.String("<button>")
		client.button(button)
	case boolArg(ctx.Args, "screenshot"):
		output, _ := ctx.Args.String("--output")
		client.screenshot(output)
	case boolArg(ctx.Args, "source"):
		output, _ := ctx.Args.String("--output")
		client.source(output)
	case boolArg(ctx.Args, "size"):
		client.size()
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
	client := uiClient{
		driver:       driver,
		wdaURL:       strings.TrimRight(wdaURL, "/"),
		deviceKitURL: strings.TrimRight(deviceKitURL, "/"),
		httpClient:   &http.Client{Timeout: 60 * time.Second},
		sessionID:    sessionID,
	}
	client.resolveDriver()
	return client
}

func (c *uiClient) resolveDriver() {
	switch c.driver {
	case uiDriverWDA, uiDriverDeviceKit:
		return
	case uiDriverAuto:
		if c.deviceKitHealthy() {
			c.driver = uiDriverDeviceKit
			return
		}
		if c.wdaHealthy() {
			c.driver = uiDriverWDA
			return
		}
		logFatal("no UI automation backend reachable; start DeviceKit on 127.0.0.1:12004 or WDA on 127.0.0.1:8100, or pass --driver and --*-url")
	default:
		logFatal("unknown --driver: " + c.driver)
	}
}

func (c uiClient) printStatus() {
	switch c.driver {
	case uiDriverDeviceKit:
		printUIResponse(c.deviceKitHTTP(http.MethodGet, "/health", nil))
	case uiDriverWDA:
		printUIResponse(c.wdaHTTP(http.MethodGet, "/status", nil))
	}
}

func (c *uiClient) api(ctx commandContext) {
	switch c.driver {
	case uiDriverDeviceKit:
		methodName := requiredStringArg(ctx.Args, "--rpc-method")
		params := rawParamsFromArgs(ctx)
		printUIResponse(c.deviceKitRPC(methodName, params))
	case uiDriverWDA:
		method, _ := ctx.Args.String("--method")
		if method == "" {
			method = http.MethodGet
		}
		httpPath := requiredStringArg(ctx.Args, "--http-path")
		body := requestBodyFromArgs(ctx)
		printUIResponse(c.wdaHTTP(strings.ToUpper(method), httpPath, body))
	}
}

func (c *uiClient) tap(x, y int) {
	switch c.driver {
	case uiDriverDeviceKit:
		printUIResponse(c.deviceKitRPC("device.io.tap", map[string]interface{}{"x": x, "y": y}))
	case uiDriverWDA:
		printUIResponse(c.wdaHTTP(http.MethodPost, wdaPath("session", c.wdaSession(), "wda", "tap", strconv.Itoa(x), strconv.Itoa(y)), nil))
	}
}

func (c *uiClient) swipe(fromX, fromY, toX, toY int, duration float64) {
	switch c.driver {
	case uiDriverDeviceKit:
		printUIResponse(c.deviceKitRPC("device.io.swipe", map[string]interface{}{"fromX": fromX, "fromY": fromY, "toX": toX, "toY": toY, "duration": duration}))
	case uiDriverWDA:
		body := map[string]interface{}{"fromX": fromX, "fromY": fromY, "toX": toX, "toY": toY, "duration": duration}
		printUIResponse(c.wdaHTTP(http.MethodPost, wdaPath("session", c.wdaSession(), "wda", "dragfromtoforduration"), mustJSON(body)))
	}
}

func (c *uiClient) longPress(x, y int, duration float64) {
	switch c.driver {
	case uiDriverDeviceKit:
		printUIResponse(c.deviceKitRPC("device.io.longpress", map[string]interface{}{"x": x, "y": y, "duration": duration}))
	case uiDriverWDA:
		body := map[string]interface{}{"x": x, "y": y, "duration": duration}
		printUIResponse(c.wdaHTTP(http.MethodPost, wdaPath("session", c.wdaSession(), "wda", "touchAndHold"), mustJSON(body)))
	}
}

func (c *uiClient) typeText(text string) {
	switch c.driver {
	case uiDriverDeviceKit:
		printUIResponse(c.deviceKitRPC("device.io.text", map[string]interface{}{"text": text}))
	case uiDriverWDA:
		printUIResponse(c.wdaHTTP(http.MethodPost, wdaPath("session", c.wdaSession(), "keys"), mustJSON(textBody(text))))
	}
}

func (c *uiClient) button(button string) {
	switch c.driver {
	case uiDriverDeviceKit:
		printUIResponse(c.deviceKitRPC("device.io.button", map[string]interface{}{"button": button}))
	case uiDriverWDA:
		if strings.EqualFold(button, "home") {
			printUIResponse(c.wdaHTTP(http.MethodPost, wdaPath("session", c.wdaSession(), "wda", "homescreen"), nil))
			return
		}
		logFatal("WDA only supports the home button through this command; use --driver=devicekit for lock/volume buttons")
	}
}

func (c *uiClient) screenshot(output string) {
	var image []byte
	switch c.driver {
	case uiDriverDeviceKit:
		resp := c.deviceKitRPC("device.screenshot", map[string]interface{}{})
		image = decodeBase64Response(resp.Body)
	case uiDriverWDA:
		resp := c.wdaHTTP(http.MethodGet, wdaPath("session", c.wdaSession(), "screenshot"), nil)
		image = decodeBase64Response(resp.Body)
	}
	if output == "" || output == "-" {
		_, err := os.Stdout.Write(image)
		exitIfError("failed writing screenshot", err)
		return
	}
	exitIfError("failed writing screenshot", os.WriteFile(output, image, 0644))
}

func (c *uiClient) source(output string) {
	var resp uiHTTPResponse
	switch c.driver {
	case uiDriverDeviceKit:
		resp = c.deviceKitRPC("device.dump.ui", map[string]interface{}{})
	case uiDriverWDA:
		resp = c.wdaHTTP(http.MethodGet, wdaPath("session", c.wdaSession(), "source"), nil)
	}
	writeOrPrintResponse(resp, output)
}

func (c *uiClient) size() {
	switch c.driver {
	case uiDriverDeviceKit:
		printUIResponse(c.deviceKitRPC("device.info", map[string]interface{}{}))
	case uiDriverWDA:
		printUIResponse(c.wdaHTTP(http.MethodGet, wdaPath("session", c.wdaSession(), "window", "size"), nil))
	}
}

func (c *uiClient) orientation(ctx commandContext) {
	switch {
	case boolArg(ctx.Args, "set"):
		orientation, _ := ctx.Args.String("<orientation>")
		if orientation == "" {
			orientation = requiredStringArg(ctx.Args, "--orientation")
		}
		switch c.driver {
		case uiDriverDeviceKit:
			printUIResponse(c.deviceKitRPC("device.io.orientation.set", map[string]interface{}{"orientation": orientation}))
		case uiDriverWDA:
			printUIResponse(c.wdaHTTP(http.MethodPost, wdaPath("session", c.wdaSession(), "orientation"), mustJSON(map[string]string{"orientation": orientation})))
		}
	default:
		switch c.driver {
		case uiDriverDeviceKit:
			printUIResponse(c.deviceKitRPC("device.io.orientation.get", map[string]interface{}{}))
		case uiDriverWDA:
			printUIResponse(c.wdaHTTP(http.MethodGet, wdaPath("session", c.wdaSession(), "orientation"), nil))
		}
	}
}

func (c *uiClient) app(ctx commandContext) {
	switch {
	case boolArg(ctx.Args, "foreground"):
		if c.driver != uiDriverDeviceKit {
			logFatal("app foreground is only available with --driver=devicekit")
		}
		printUIResponse(c.deviceKitRPC("device.apps.foreground", map[string]interface{}{}))
	default:
		bundleID, _ := ctx.Args.String("<bundleID>")
		if bundleID == "" {
			bundleID = requiredStringArg(ctx.Args, "--bundle-id")
		}
		switch {
		case boolArg(ctx.Args, "launch"):
			c.appWithBundleID("launch", bundleID)
		case boolArg(ctx.Args, "terminate"):
			c.appWithBundleID("terminate", bundleID)
		default:
			logFatal("unknown ui app command")
		}
	}
}

func (c *uiClient) appWithBundleID(action string, bundleID string) {
	switch c.driver {
	case uiDriverDeviceKit:
		printUIResponse(c.deviceKitRPC("device.apps."+action, map[string]interface{}{"bundleId": bundleID}))
	case uiDriverWDA:
		printUIResponse(c.wdaHTTP(http.MethodPost, wdaPath("session", c.wdaSession(), "wda", "apps", action), mustJSON(map[string]string{"bundleId": bundleID})))
	}
}

func (c *uiClient) stream(ctx commandContext) {
	streamType := "mjpeg"
	if boolArg(ctx.Args, "h264") {
		streamType = "h264"
	}
	query := url.Values{}
	addQueryArg(ctx.Args, query, "--fps", "fps")
	addQueryArg(ctx.Args, query, "--quality", "quality")
	addQueryArg(ctx.Args, query, "--scale", "scale")
	addQueryArg(ctx.Args, query, "--bitrate", "bitrate")
	endpoint := "/" + streamType
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	switch c.driver {
	case uiDriverDeviceKit:
		c.pipeHTTP(c.deviceKitURL + endpoint)
	case uiDriverWDA:
		if streamType != "mjpeg" {
			logFatal("WDA stream supports mjpeg only; use --driver=devicekit for h264")
		}
		c.pipeHTTP(c.wdaURL + endpoint)
	}
}

func (c *uiClient) wdaSession() string {
	if c.sessionID != "" {
		return c.sessionID
	}
	resp := c.wdaHTTP(http.MethodPost, "/session", mustJSON(map[string]interface{}{
		"capabilities":        map[string]interface{}{},
		"desiredCapabilities": map[string]interface{}{},
	}))
	sessionID := extractSessionID(resp.Body)
	if sessionID == "" {
		logFatal("WDA did not return a session id")
	}
	c.sessionID = sessionID
	return sessionID
}

func (c uiClient) deviceKitHealthy() bool {
	resp, err := c.httpClient.Get(c.deviceKitURL + "/health")
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func (c uiClient) wdaHealthy() bool {
	resp, err := c.httpClient.Get(c.wdaURL + "/status")
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func (c uiClient) deviceKitRPC(method string, params interface{}) uiHTTPResponse {
	body := mustJSON(map[string]interface{}{
		"jsonrpc": deviceKitRPCProtocol,
		"method":  method,
		"params":  params,
		"id":      1,
	})
	return c.deviceKitHTTP(http.MethodPost, "/rpc", body)
}

func (c uiClient) deviceKitHTTP(method string, endpoint string, body []byte) uiHTTPResponse {
	return c.doHTTP(method, c.deviceKitURL, endpoint, body)
}

func (c uiClient) wdaHTTP(method string, endpoint string, body []byte) uiHTTPResponse {
	return c.doHTTP(method, c.wdaURL, endpoint, body)
}

func (c uiClient) doHTTP(method string, baseURL string, endpoint string, body []byte) uiHTTPResponse {
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	req, err := http.NewRequest(method, baseURL+endpoint, bytes.NewReader(body))
	exitIfError("failed creating UI request", err)
	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	exitIfError("UI request failed", err)
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	exitIfError("failed reading UI response", err)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		fmt.Fprintln(os.Stderr, string(respBody))
		os.Exit(1)
	}
	return uiHTTPResponse{StatusCode: resp.StatusCode, Header: resp.Header, Body: respBody}
}

func (c uiClient) pipeHTTP(rawURL string) {
	resp, err := c.httpClient.Get(rawURL)
	exitIfError("stream request failed", err)
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		fmt.Fprintln(os.Stderr, string(body))
		os.Exit(1)
	}
	_, err = io.Copy(os.Stdout, resp.Body)
	exitIfError("stream copy failed", err)
}

func printUIResponse(resp uiHTTPResponse) {
	if len(resp.Body) == 0 {
		return
	}
	var data interface{}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		fmt.Print(string(resp.Body))
		return
	}
	fmt.Println(convertToJSONString(data))
}

func writeOrPrintResponse(resp uiHTTPResponse, output string) {
	if output == "" || output == "-" {
		printUIResponse(resp)
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

func textBody(text string) map[string]interface{} {
	return map[string]interface{}{
		"text":  text,
		"value": []string{text},
	}
}

func wdaPath(parts ...string) string {
	return "/" + path.Join(parts...)
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

func mustJSON(data interface{}) []byte {
	body, err := json.Marshal(data)
	exitIfError("failed encoding UI request body", err)
	return body
}

func decodeBase64Response(body []byte) []byte {
	encoded := findBase64Value(body)
	if encoded == "" {
		logFatal("response did not contain a base64 image")
	}
	image, err := base64.StdEncoding.DecodeString(encoded)
	exitIfError("failed decoding base64 image", err)
	return image
}

func findBase64Value(body []byte) string {
	var decoded interface{}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return ""
	}
	return findBase64String(decoded)
}

func findBase64String(value interface{}) string {
	switch typed := value.(type) {
	case string:
		return typed
	case map[string]interface{}:
		for _, key := range []string{"value", "result", "image", "screenshot", "data"} {
			if nested, ok := typed[key]; ok {
				if result := findBase64String(nested); result != "" {
					return result
				}
			}
		}
	}
	return ""
}

func extractSessionID(body []byte) string {
	var decoded map[string]interface{}
	if err := json.Unmarshal(body, &decoded); err != nil {
		return ""
	}
	if sessionID, ok := decoded["sessionId"].(string); ok {
		return sessionID
	}
	if value, ok := decoded["value"].(map[string]interface{}); ok {
		if sessionID, ok := value["sessionId"].(string); ok {
			return sessionID
		}
		if sessionID, ok := value["session_id"].(string); ok {
			return sessionID
		}
	}
	return ""
}

func addQueryArg(args docopt.Opts, query url.Values, argName string, queryName string) {
	value, _ := args.String(argName)
	if value != "" {
		query.Set(queryName, value)
	}
}
