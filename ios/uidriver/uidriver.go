// Package uidriver is a reusable client for driving on-device UI automation
// backends over HTTP. It talks to a running WebDriverAgent (WDA, default
// :8100) or DeviceKit (default :12004) instance — typically reached through a
// forwarded local port — and exposes the common automation actions (tap,
// swipe, type, screenshot, source, app control, streaming, …) as a clean,
// exported API.
//
// The client is deliberately decoupled from the go-ios CLI: construct a Driver
// against a backend base URL and call its methods. All methods return values
// and errors rather than terminating the process, so the driver is safe to
// embed in the CLI, the REST API, or any other caller.
package uidriver

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios/golog"
)

const logModule = "go-ios/uidriver"

// Backend identifies which UI automation backend a Driver talks to.
type Backend string

const (
	// BackendWDA drives a WebDriverAgent instance (WebDriver JSON wire protocol).
	BackendWDA Backend = "wda"
	// BackendDeviceKit drives a DeviceKit instance (JSON-RPC 2.0).
	BackendDeviceKit Backend = "devicekit"

	// DefaultWDAURL is the default forwarded WebDriverAgent address.
	DefaultWDAURL = "http://127.0.0.1:8100"
	// DefaultDeviceKitURL is the default forwarded DeviceKit address.
	DefaultDeviceKitURL = "http://127.0.0.1:12004"

	// DefaultTimeout is the default per-request HTTP timeout.
	DefaultTimeout = 60 * time.Second

	deviceKitRPCProtocol = "2.0"
)

// Response is the raw result of a backend HTTP call. Body holds the raw
// response bytes (usually JSON); callers can unmarshal it as needed.
type Response struct {
	StatusCode int         `json:"statusCode"`
	Header     http.Header `json:"-"`
	Body       []byte      `json:"body"`
}

// UnmarshalBody decodes the response body into v.
func (r Response) UnmarshalBody(v interface{}) error {
	return json.Unmarshal(r.Body, v)
}

// Driver is a UI automation client bound to a single backend (WDA or
// DeviceKit) reachable at a base URL. Construct it with New. A Driver is safe
// for sequential use; it lazily establishes a WDA session on first use of a
// session-scoped action.
type Driver struct {
	backend    Backend
	baseURL    string
	httpClient *http.Client
	timeout    time.Duration
	sessionID  string
	// udid is optional, used only to enrich log lines.
	udid string
}

// Option configures a Driver.
type Option func(*Driver)

// WithHTTPClient overrides the HTTP client used for requests.
func WithHTTPClient(c *http.Client) Option {
	return func(d *Driver) {
		if c != nil {
			d.httpClient = c
		}
	}
}

// WithTimeout sets the per-request HTTP timeout (ignored if WithHTTPClient is
// also supplied).
func WithTimeout(timeout time.Duration) Option {
	return func(d *Driver) { d.timeout = timeout }
}

// WithSessionID pre-seeds a WDA session id so the Driver skips session
// creation. Ignored by the DeviceKit backend.
func WithSessionID(sessionID string) Option {
	return func(d *Driver) { d.sessionID = sessionID }
}

// WithUDID attaches a device udid used only to enrich log lines.
func WithUDID(udid string) Option {
	return func(d *Driver) { d.udid = udid }
}

// New constructs a Driver for the given backend at baseURL (for example
// "http://127.0.0.1:8100" for WDA or "http://127.0.0.1:12004" for DeviceKit).
// The base URL is typically the local end of a forwarded connection to the
// device. A trailing slash is trimmed.
func New(backend Backend, baseURL string, opts ...Option) (*Driver, error) {
	switch backend {
	case BackendWDA, BackendDeviceKit:
	default:
		return nil, fmt.Errorf("uidriver: unknown backend %q", backend)
	}
	if strings.TrimSpace(baseURL) == "" {
		return nil, errors.New("uidriver: baseURL must not be empty")
	}
	d := &Driver{
		backend: backend,
		baseURL: strings.TrimRight(baseURL, "/"),
		timeout: DefaultTimeout,
	}
	for _, opt := range opts {
		opt(d)
	}
	if d.httpClient == nil {
		d.httpClient = &http.Client{Timeout: d.timeout}
	}
	return d, nil
}

// Backend returns the backend this Driver targets.
func (d *Driver) Backend() Backend { return d.backend }

// BaseURL returns the (trailing-slash-trimmed) backend base URL.
func (d *Driver) BaseURL() string { return d.baseURL }

// SessionID returns the current WDA session id (empty until one is created).
func (d *Driver) SessionID() string { return d.sessionID }

// Healthy reports whether the backend answers its health/status endpoint with
// a 2xx status.
func (d *Driver) Healthy() bool {
	endpoint := "/status"
	if d.backend == BackendDeviceKit {
		endpoint = "/health"
	}
	resp, err := d.httpClient.Get(d.baseURL + endpoint)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

// Status returns the backend status/health payload.
func (d *Driver) Status() (Response, error) {
	switch d.backend {
	case BackendDeviceKit:
		return d.deviceKitHTTP(http.MethodGet, "/health", nil)
	default:
		return d.wdaHTTP(http.MethodGet, "/status", nil)
	}
}

// Tap taps the screen at absolute coordinates (x, y).
func (d *Driver) Tap(x, y int) (Response, error) {
	switch d.backend {
	case BackendDeviceKit:
		return d.deviceKitRPC("device.io.tap", map[string]interface{}{"x": x, "y": y})
	default:
		session, err := d.wdaSession()
		if err != nil {
			return Response{}, err
		}
		return d.wdaHTTP(http.MethodPost, wdaPath("session", session, "wda", "tap", strconv.Itoa(x), strconv.Itoa(y)), nil)
	}
}

// Swipe drags from (fromX, fromY) to (toX, toY) over the given duration in
// seconds (0 lets the backend choose).
func (d *Driver) Swipe(fromX, fromY, toX, toY int, duration float64) (Response, error) {
	body := map[string]interface{}{"fromX": fromX, "fromY": fromY, "toX": toX, "toY": toY, "duration": duration}
	switch d.backend {
	case BackendDeviceKit:
		return d.deviceKitRPC("device.io.swipe", body)
	default:
		session, err := d.wdaSession()
		if err != nil {
			return Response{}, err
		}
		payload, err := marshalJSON(body)
		if err != nil {
			return Response{}, err
		}
		return d.wdaHTTP(http.MethodPost, wdaPath("session", session, "wda", "dragfromtoforduration"), payload)
	}
}

// LongPress presses and holds at (x, y) for duration seconds.
func (d *Driver) LongPress(x, y int, duration float64) (Response, error) {
	body := map[string]interface{}{"x": x, "y": y, "duration": duration}
	switch d.backend {
	case BackendDeviceKit:
		return d.deviceKitRPC("device.io.longpress", body)
	default:
		session, err := d.wdaSession()
		if err != nil {
			return Response{}, err
		}
		payload, err := marshalJSON(body)
		if err != nil {
			return Response{}, err
		}
		return d.wdaHTTP(http.MethodPost, wdaPath("session", session, "wda", "touchAndHold"), payload)
	}
}

// Type sends text as keyboard input.
func (d *Driver) Type(text string) (Response, error) {
	switch d.backend {
	case BackendDeviceKit:
		return d.deviceKitRPC("device.io.text", map[string]interface{}{"text": text})
	default:
		session, err := d.wdaSession()
		if err != nil {
			return Response{}, err
		}
		payload, err := marshalJSON(textBody(text))
		if err != nil {
			return Response{}, err
		}
		return d.wdaHTTP(http.MethodPost, wdaPath("session", session, "keys"), payload)
	}
}

// ErrButtonUnsupported is returned when a hardware button press is requested
// against a backend that cannot perform it.
var ErrButtonUnsupported = errors.New("uidriver: button not supported by this backend")

// PressButton presses a hardware button by name (for example "home", "lock",
// "volumeup"). WDA only supports "home"; other buttons require DeviceKit and
// yield ErrButtonUnsupported on WDA.
func (d *Driver) PressButton(name string) (Response, error) {
	switch d.backend {
	case BackendDeviceKit:
		return d.deviceKitRPC("device.io.button", map[string]interface{}{"button": name})
	default:
		if strings.EqualFold(name, "home") {
			session, err := d.wdaSession()
			if err != nil {
				return Response{}, err
			}
			return d.wdaHTTP(http.MethodPost, wdaPath("session", session, "wda", "homescreen"), nil)
		}
		return Response{}, fmt.Errorf("%w: WDA only supports the home button; use the DeviceKit backend for lock/volume buttons", ErrButtonUnsupported)
	}
}

// Screenshot captures the screen and returns the decoded image bytes (PNG).
func (d *Driver) Screenshot() ([]byte, error) {
	var resp Response
	var err error
	switch d.backend {
	case BackendDeviceKit:
		resp, err = d.deviceKitRPC("device.screenshot", map[string]interface{}{})
	default:
		var session string
		session, err = d.wdaSession()
		if err != nil {
			return nil, err
		}
		resp, err = d.wdaHTTP(http.MethodGet, wdaPath("session", session, "screenshot"), nil)
	}
	if err != nil {
		return nil, err
	}
	return decodeBase64Response(resp.Body)
}

// Source returns the current UI hierarchy (XML for WDA, backend-defined dump
// for DeviceKit).
func (d *Driver) Source() (Response, error) {
	switch d.backend {
	case BackendDeviceKit:
		return d.deviceKitRPC("device.dump.ui", map[string]interface{}{})
	default:
		session, err := d.wdaSession()
		if err != nil {
			return Response{}, err
		}
		return d.wdaHTTP(http.MethodGet, wdaPath("session", session, "source"), nil)
	}
}

// WindowSize returns the device window/screen size payload.
func (d *Driver) WindowSize() (Response, error) {
	switch d.backend {
	case BackendDeviceKit:
		return d.deviceKitRPC("device.info", map[string]interface{}{})
	default:
		session, err := d.wdaSession()
		if err != nil {
			return Response{}, err
		}
		return d.wdaHTTP(http.MethodGet, wdaPath("session", session, "window", "size"), nil)
	}
}

// Orientation returns the current device orientation payload.
func (d *Driver) Orientation() (Response, error) {
	switch d.backend {
	case BackendDeviceKit:
		return d.deviceKitRPC("device.io.orientation.get", map[string]interface{}{})
	default:
		session, err := d.wdaSession()
		if err != nil {
			return Response{}, err
		}
		return d.wdaHTTP(http.MethodGet, wdaPath("session", session, "orientation"), nil)
	}
}

// SetOrientation sets the device orientation (for example "PORTRAIT",
// "LANDSCAPE").
func (d *Driver) SetOrientation(orientation string) (Response, error) {
	switch d.backend {
	case BackendDeviceKit:
		return d.deviceKitRPC("device.io.orientation.set", map[string]interface{}{"orientation": orientation})
	default:
		session, err := d.wdaSession()
		if err != nil {
			return Response{}, err
		}
		payload, err := marshalJSON(map[string]string{"orientation": orientation})
		if err != nil {
			return Response{}, err
		}
		return d.wdaHTTP(http.MethodPost, wdaPath("session", session, "orientation"), payload)
	}
}

// AppLaunch launches the app identified by bundleID.
func (d *Driver) AppLaunch(bundleID string) (Response, error) {
	return d.appWithBundleID("launch", bundleID)
}

// AppTerminate terminates the app identified by bundleID.
func (d *Driver) AppTerminate(bundleID string) (Response, error) {
	return d.appWithBundleID("terminate", bundleID)
}

// ErrForegroundUnsupported is returned when app foregrounding is requested
// against a backend that does not support it (WDA).
var ErrForegroundUnsupported = errors.New("uidriver: app foreground not supported by this backend")

// AppForeground brings the currently backgrounded app to the foreground. Only
// supported by the DeviceKit backend; returns ErrForegroundUnsupported on WDA.
func (d *Driver) AppForeground() (Response, error) {
	if d.backend != BackendDeviceKit {
		return Response{}, ErrForegroundUnsupported
	}
	return d.deviceKitRPC("device.apps.foreground", map[string]interface{}{})
}

func (d *Driver) appWithBundleID(action, bundleID string) (Response, error) {
	switch d.backend {
	case BackendDeviceKit:
		return d.deviceKitRPC("device.apps."+action, map[string]interface{}{"bundleId": bundleID})
	default:
		session, err := d.wdaSession()
		if err != nil {
			return Response{}, err
		}
		payload, err := marshalJSON(map[string]string{"bundleId": bundleID})
		if err != nil {
			return Response{}, err
		}
		return d.wdaHTTP(http.MethodPost, wdaPath("session", session, "wda", "apps", action), payload)
	}
}

// APIRequest is a raw passthrough to the backend.
//
// For the WDA backend, Method and Path are the HTTP verb and path (Body may be
// nil), and RPCMethod/RPCParams are ignored.
//
// For the DeviceKit backend, RPCMethod is the JSON-RPC method name and
// RPCParams its params (Method/Path/Body are ignored). If RPCParams is nil an
// empty object is sent.
type APIRequest struct {
	// Method is the HTTP method for a WDA passthrough (defaults to GET).
	Method string `json:"method,omitempty"`
	// Path is the HTTP path for a WDA passthrough.
	Path string `json:"path,omitempty"`
	// Body is the raw HTTP request body for a WDA passthrough.
	Body []byte `json:"body,omitempty"`
	// RPCMethod is the JSON-RPC method name for a DeviceKit passthrough.
	RPCMethod string `json:"rpcMethod,omitempty"`
	// RPCParams are the JSON-RPC params for a DeviceKit passthrough.
	RPCParams interface{} `json:"rpcParams,omitempty"`
}

// API performs a raw passthrough request against the backend. See APIRequest
// for which fields apply to which backend.
func (d *Driver) API(req APIRequest) (Response, error) {
	switch d.backend {
	case BackendDeviceKit:
		if req.RPCMethod == "" {
			return Response{}, errors.New("uidriver: RPCMethod is required for the DeviceKit backend")
		}
		params := req.RPCParams
		if params == nil {
			params = map[string]interface{}{}
		}
		return d.deviceKitRPC(req.RPCMethod, params)
	default:
		if req.Path == "" {
			return Response{}, errors.New("uidriver: Path is required for the WDA backend")
		}
		method := req.Method
		if method == "" {
			method = http.MethodGet
		}
		return d.wdaHTTP(strings.ToUpper(method), req.Path, req.Body)
	}
}

// StreamOptions configure a video/mjpeg stream request.
type StreamOptions struct {
	// H264 requests an h264 stream (DeviceKit only); otherwise mjpeg is used.
	H264 bool
	// FPS, Quality, Scale and Bitrate are optional query parameters passed
	// through to the backend when non-empty.
	FPS     string
	Quality string
	Scale   string
	Bitrate string
}

// ErrStreamUnsupported is returned when the requested stream type is not
// supported by the backend (for example h264 over WDA).
var ErrStreamUnsupported = errors.New("uidriver: stream type not supported by this backend")

// Stream opens a video stream against the backend and returns the raw response
// body as an io.ReadCloser for the caller to pipe. The caller owns the returned
// ReadCloser and must Close it. A non-2xx response is returned as an *HTTPError
// with the response body included.
func (d *Driver) Stream(ctx context.Context, opts StreamOptions) (io.ReadCloser, error) {
	streamType := "mjpeg"
	if opts.H264 {
		streamType = "h264"
	}
	if d.backend == BackendWDA && streamType != "mjpeg" {
		return nil, fmt.Errorf("%w: WDA stream supports mjpeg only; use the DeviceKit backend for h264", ErrStreamUnsupported)
	}
	query := url.Values{}
	setIfNotEmpty(query, "fps", opts.FPS)
	setIfNotEmpty(query, "quality", opts.Quality)
	setIfNotEmpty(query, "scale", opts.Scale)
	setIfNotEmpty(query, "bitrate", opts.Bitrate)
	endpoint := "/" + streamType
	if encoded := query.Encode(); encoded != "" {
		endpoint += "?" + encoded
	}
	rawURL := d.baseURL + endpoint

	if ctx == nil {
		ctx = context.Background()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("uidriver: failed creating stream request: %w", err)
	}
	golog.Info("opening ui stream", "module", logModule, "backend", string(d.backend), "url", rawURL, "udid", d.udid)
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("uidriver: stream request failed: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, &HTTPError{StatusCode: resp.StatusCode, Body: body}
	}
	return resp.Body, nil
}

// HTTPError is returned when a backend responds with a non-2xx status.
type HTTPError struct {
	StatusCode int
	Body       []byte
}

func (e *HTTPError) Error() string {
	msg := strings.TrimSpace(string(e.Body))
	if msg == "" {
		return fmt.Sprintf("uidriver: backend returned status %d", e.StatusCode)
	}
	return fmt.Sprintf("uidriver: backend returned status %d: %s", e.StatusCode, msg)
}

// wdaSession returns an existing session id or lazily creates one.
func (d *Driver) wdaSession() (string, error) {
	if d.sessionID != "" {
		return d.sessionID, nil
	}
	payload, err := marshalJSON(map[string]interface{}{
		"capabilities":        map[string]interface{}{},
		"desiredCapabilities": map[string]interface{}{},
	})
	if err != nil {
		return "", err
	}
	resp, err := d.wdaHTTP(http.MethodPost, "/session", payload)
	if err != nil {
		return "", err
	}
	sessionID := extractSessionID(resp.Body)
	if sessionID == "" {
		return "", errors.New("uidriver: WDA did not return a session id")
	}
	d.sessionID = sessionID
	golog.Debug("created wda session", "module", logModule, "sessionID", sessionID, "udid", d.udid)
	return sessionID, nil
}

func (d *Driver) deviceKitRPC(method string, params interface{}) (Response, error) {
	body, err := marshalJSON(map[string]interface{}{
		"jsonrpc": deviceKitRPCProtocol,
		"method":  method,
		"params":  params,
		"id":      1,
	})
	if err != nil {
		return Response{}, err
	}
	return d.deviceKitHTTP(http.MethodPost, "/rpc", body)
}

func (d *Driver) deviceKitHTTP(method, endpoint string, body []byte) (Response, error) {
	return d.doHTTP(method, endpoint, body)
}

func (d *Driver) wdaHTTP(method, endpoint string, body []byte) (Response, error) {
	return d.doHTTP(method, endpoint, body)
}

func (d *Driver) doHTTP(method, endpoint string, body []byte) (Response, error) {
	if !strings.HasPrefix(endpoint, "/") {
		endpoint = "/" + endpoint
	}
	req, err := http.NewRequest(method, d.baseURL+endpoint, bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("uidriver: failed creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	if len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	golog.Debug("ui request", "module", logModule, "backend", string(d.backend), "method", method, "endpoint", endpoint, "udid", d.udid)
	resp, err := d.httpClient.Do(req)
	if err != nil {
		return Response{}, fmt.Errorf("uidriver: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return Response{}, fmt.Errorf("uidriver: failed reading response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Response{}, &HTTPError{StatusCode: resp.StatusCode, Body: respBody}
	}
	return Response{StatusCode: resp.StatusCode, Header: resp.Header, Body: respBody}, nil
}

func marshalJSON(data interface{}) ([]byte, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("uidriver: failed encoding request body: %w", err)
	}
	return body, nil
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

func setIfNotEmpty(query url.Values, key, value string) {
	if value != "" {
		query.Set(key, value)
	}
}

func decodeBase64Response(body []byte) ([]byte, error) {
	encoded := findBase64Value(body)
	if encoded == "" {
		return nil, errors.New("uidriver: response did not contain a base64 image")
	}
	image, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("uidriver: failed decoding base64 image: %w", err)
	}
	return image, nil
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
