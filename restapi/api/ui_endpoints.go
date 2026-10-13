package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/uidriver"
	"github.com/gin-gonic/gin"
)

// UI automation endpoints proxy to a running, forwarded WebDriverAgent (WDA) or
// DeviceKit instance via the ios/uidriver package.
//
// PREREQUISITE: these endpoints do not start WDA. The caller must first bring up
// and forward a WebDriverAgent session on the device, typically via
// `POST /jobs/runwda` followed by `POST /jobs/forward`, so the backend is
// reachable on a local port. The endpoints then connect to that forwarded
// backend per request.
//
// Backend addressing (per request):
//   - ?backend=wda|devicekit (or the "backend" header) selects the backend;
//     defaults to wda.
//   - ?wdaUrl=<url> (or the "wdaUrl" header) is the forwarded backend base URL.
//     Defaults to http://127.0.0.1:8100 for wda and http://127.0.0.1:12004 for
//     devicekit.
//   - ?timeout=<seconds> (or the "timeout" header) overrides the per-request
//     HTTP timeout (default 60s).
//
// A fresh uidriver.Driver is constructed for every request; no session state is
// held server-side beyond what WDA itself keeps.

// Sentinel errors for UI request validation.
var (
	errMissingText        = errors.New("missing required 'text' field")
	errMissingButtonName  = errors.New("missing required 'name' field")
	errMissingOrientation = errors.New("missing required 'orientation' field")
	errMissingBundleIDUI  = errors.New("missing required 'bundleId' field")
	errMissingAPIPath     = errors.New("missing required 'path' field for the wda backend")
)

// registerUIRoutes registers the UI-automation endpoints under
// /device/:udid/ui. They proxy to a forwarded WDA/DeviceKit backend; see the
// package-level doc comment above for the prerequisites and backend addressing.
func registerUIRoutes(device *gin.RouterGroup) {
	group := device.Group("/ui")
	group.POST("/tap", UITap)
	group.POST("/swipe", UISwipe)
	group.POST("/longpress", UILongPress)
	group.POST("/type", UIType)
	group.POST("/button", UIButton)
	group.GET("/screenshot", UIScreenshot)
	group.GET("/source", UISource)
	group.GET("/size", UIWindowSize)
	group.GET("/orientation", UIGetOrientation)
	group.PUT("/orientation", UISetOrientation)
	group.GET("/status", UIStatus)
	group.POST("/app/launch", UIAppLaunch)
	group.POST("/app/terminate", UIAppTerminate)
	group.POST("/app/foreground", UIAppForeground)
	group.POST("/api", UIAPI)
}

// newDriverFromRequest builds a uidriver.Driver for the current request from the
// backend/wdaUrl/timeout query params (or same-named headers) and the device
// udid. It returns the driver and the resolved udid.
func newDriverFromRequest(c *gin.Context) (*uidriver.Driver, string, error) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	udid := device.Properties.SerialNumber

	backend := uidriver.BackendWDA
	switch strings.ToLower(strings.TrimSpace(paramOrHeader(c, "backend"))) {
	case "", string(uidriver.BackendWDA):
		backend = uidriver.BackendWDA
	case string(uidriver.BackendDeviceKit):
		backend = uidriver.BackendDeviceKit
	default:
		return nil, udid, errors.New("unknown backend; expected wda or devicekit")
	}

	baseURL := strings.TrimSpace(paramOrHeader(c, "wdaUrl"))
	if baseURL == "" {
		if backend == uidriver.BackendDeviceKit {
			baseURL = uidriver.DefaultDeviceKitURL
		} else {
			baseURL = uidriver.DefaultWDAURL
		}
	}

	opts := []uidriver.Option{uidriver.WithUDID(udid)}
	if raw := strings.TrimSpace(paramOrHeader(c, "timeout")); raw != "" {
		secs, err := strconv.Atoi(raw)
		if err != nil || secs <= 0 {
			return nil, udid, errInvalidTimeout
		}
		opts = append(opts, uidriver.WithTimeout(time.Duration(secs)*time.Second))
	}

	driver, err := uidriver.New(backend, baseURL, opts...)
	if err != nil {
		return nil, udid, err
	}
	return driver, udid, nil
}

// paramOrHeader reads a value from the query string first, falling back to a
// same-named request header.
func paramOrHeader(c *gin.Context, name string) string {
	if v := c.Query(name); v != "" {
		return v
	}
	return c.GetHeader(name)
}

// respondUIError maps uidriver failures to sensible HTTP statuses and writes the
// standard error envelope.
//
//	*uidriver.HTTPError    -> 502 (the WDA/DeviceKit backend rejected/failed)
//	ErrButtonUnsupported / ErrForegroundUnsupported / ErrStreamUnsupported -> 501
//	everything else (transport failure, unreachable backend, decode error) -> 502
func respondUIError(c *gin.Context, udid string, err error) {
	var httpErr *uidriver.HTTPError
	switch {
	case errors.As(err, &httpErr):
		golog.Warn("ui backend returned error", "module", logModule, "udid", udid, "status", httpErr.StatusCode)
		RespondError(c, http.StatusBadGateway, err)
	case errors.Is(err, uidriver.ErrButtonUnsupported),
		errors.Is(err, uidriver.ErrForegroundUnsupported),
		errors.Is(err, uidriver.ErrStreamUnsupported):
		RespondError(c, http.StatusNotImplemented, err)
	default:
		golog.Warn("ui backend unreachable", "module", logModule, "udid", udid, "error", err.Error())
		RespondError(c, http.StatusBadGateway, err)
	}
}

// writeUIResponse forwards a uidriver.Response back to the client, preserving
// the backend's Content-Type when present and defaulting to JSON.
func writeUIResponse(c *gin.Context, resp uidriver.Response) {
	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	status := resp.StatusCode
	if status == 0 {
		status = http.StatusOK
	}
	c.Data(status, contentType, resp.Body)
}

// UITap taps at absolute coordinates {x,y}.
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UITap(c *gin.Context) {
	var body struct {
		X int `json:"x"`
		Y int `json:"y"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	resp, err := driver.Tap(body.X, body.Y)
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	writeUIResponse(c, resp)
}

// UISwipe drags from {x1,y1} to {x2,y2} over an optional duration (seconds).
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UISwipe(c *gin.Context) {
	var body struct {
		X1       int     `json:"x1"`
		Y1       int     `json:"y1"`
		X2       int     `json:"x2"`
		Y2       int     `json:"y2"`
		Duration float64 `json:"duration"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	resp, err := driver.Swipe(body.X1, body.Y1, body.X2, body.Y2, body.Duration)
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	writeUIResponse(c, resp)
}

// UILongPress presses and holds at {x,y} for an optional duration (seconds).
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UILongPress(c *gin.Context) {
	var body struct {
		X        int     `json:"x"`
		Y        int     `json:"y"`
		Duration float64 `json:"duration"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	resp, err := driver.LongPress(body.X, body.Y, body.Duration)
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	writeUIResponse(c, resp)
}

// UIType sends {text} as keyboard input.
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UIType(c *gin.Context) {
	var body struct {
		Text string `json:"text"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	if body.Text == "" {
		RespondError(c, http.StatusBadRequest, errMissingText)
		return
	}
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	resp, err := driver.Type(body.Text)
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	writeUIResponse(c, resp)
}

// UIButton presses a hardware button by {name} (home/volumeup/...). WDA only
// supports the home button; other buttons require the devicekit backend.
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UIButton(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	if body.Name == "" {
		RespondError(c, http.StatusBadRequest, errMissingButtonName)
		return
	}
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	resp, err := driver.PressButton(body.Name)
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	writeUIResponse(c, resp)
}

// UIScreenshot captures the screen and returns raw PNG bytes (image/png).
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UIScreenshot(c *gin.Context) {
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	image, err := driver.Screenshot()
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	c.Data(http.StatusOK, "image/png", image)
}

// UISource returns the current view hierarchy as the backend supplies it (XML
// for WDA, backend-defined for DeviceKit); the backend Content-Type is
// preserved.
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UISource(c *gin.Context) {
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	resp, err := driver.Source()
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	writeUIResponse(c, resp)
}

// UIWindowSize returns the device window/screen size payload as the backend
// returns it (typically {width,height}).
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UIWindowSize(c *gin.Context) {
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	resp, err := driver.WindowSize()
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	writeUIResponse(c, resp)
}

// UIGetOrientation returns the current device orientation payload.
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UIGetOrientation(c *gin.Context) {
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	resp, err := driver.Orientation()
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	writeUIResponse(c, resp)
}

// UISetOrientation sets the device orientation from {orientation} (for example
// PORTRAIT, LANDSCAPE).
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UISetOrientation(c *gin.Context) {
	var body struct {
		Orientation string `json:"orientation"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	if body.Orientation == "" {
		RespondError(c, http.StatusBadRequest, errMissingOrientation)
		return
	}
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	resp, err := driver.SetOrientation(body.Orientation)
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	writeUIResponse(c, resp)
}

// UIStatus returns the backend status/health payload (WDA /status or DeviceKit
// /health).
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UIStatus(c *gin.Context) {
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	resp, err := driver.Status()
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	writeUIResponse(c, resp)
}

// UIAppLaunch launches the app identified by {bundleId}.
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UIAppLaunch(c *gin.Context) {
	uiAppAction(c, func(d *uidriver.Driver, bundleID string) (uidriver.Response, error) {
		return d.AppLaunch(bundleID)
	})
}

// UIAppTerminate terminates the app identified by {bundleId}.
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UIAppTerminate(c *gin.Context) {
	uiAppAction(c, func(d *uidriver.Driver, bundleID string) (uidriver.Response, error) {
		return d.AppTerminate(bundleID)
	})
}

// UIAppForeground brings the backgrounded app to the foreground. Only the
// devicekit backend supports this; WDA returns 501. The request body is
// ignored.
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UIAppForeground(c *gin.Context) {
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	resp, err := driver.AppForeground()
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	writeUIResponse(c, resp)
}

// uiAppAction is the shared body-parsing + dispatch for launch/terminate, which
// both take a {bundleId} body and a Driver method.
func uiAppAction(c *gin.Context, action func(*uidriver.Driver, string) (uidriver.Response, error)) {
	var body struct {
		BundleID string `json:"bundleId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	if body.BundleID == "" {
		RespondError(c, http.StatusBadRequest, errMissingBundleIDUI)
		return
	}
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	resp, err := action(driver, body.BundleID)
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	writeUIResponse(c, resp)
}

// UIAPI is a raw passthrough to the backend. The JSON body is:
//
//	{"method":"POST","path":"/session/.../wda/tap/0/0","body":<raw>}   (wda)
//	{"rpcMethod":"device.info","rpcParams":{...}}                       (devicekit)
//
// For the wda backend method defaults to GET and path is required; body is the
// raw request body as bytes. For devicekit, rpcMethod is required.
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UIAPI(c *gin.Context) {
	var req uidriver.APIRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	// Validate up front so bad input maps to 400 rather than surfacing as a
	// generic uidriver error (which respondUIError would treat as a 502).
	if driver.Backend() == uidriver.BackendWDA && req.Path == "" {
		RespondError(c, http.StatusBadRequest, errMissingAPIPath)
		return
	}
	if driver.Backend() == uidriver.BackendDeviceKit && req.RPCMethod == "" {
		RespondError(c, http.StatusBadRequest, errors.New("missing required 'rpcMethod' field for the devicekit backend"))
		return
	}
	resp, err := driver.API(req)
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	writeUIResponse(c, resp)
}
