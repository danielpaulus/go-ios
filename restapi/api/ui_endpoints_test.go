package api_test

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/restapi/api"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests are fully device-free: the "device" is just an HTTP backend that
// mimics WebDriverAgent. We stand up an httptest.Server as the fake WDA, point
// the ui endpoints at it via ?wdaUrl=<server>, and assert each endpoint forwards
// the correct method/path/body and returns the mapped response.

// recordedRequest captures one call the ui handler made to the fake WDA backend.
type recordedRequest struct {
	Method string
	Path   string
	Body   string
}

// fakeWDA is a stand-in WebDriverAgent HTTP server. It records every request and
// answers session creation plus arbitrary paths with a canned 200 body.
type fakeWDA struct {
	server   *httptest.Server
	mu       sync.Mutex
	requests []recordedRequest
	// screenshotB64 is returned (wrapped in {"value":...}) for screenshot paths.
	screenshotB64 string
}

func newFakeWDA(t *testing.T) *fakeWDA {
	t.Helper()
	f := &fakeWDA{}
	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		f.requests = append(f.requests, recordedRequest{Method: r.Method, Path: r.URL.Path, Body: string(body)})
		f.mu.Unlock()

		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/session" && r.Method == http.MethodPost:
			_, _ = w.Write([]byte(`{"sessionId":"sess-123","value":{"sessionId":"sess-123"}}`))
		case strings.HasSuffix(r.URL.Path, "/screenshot"):
			_, _ = w.Write([]byte(`{"value":"` + f.screenshotB64 + `"}`))
		case r.URL.Path == "/status":
			_, _ = w.Write([]byte(`{"value":{"ready":true}}`))
		case strings.HasSuffix(r.URL.Path, "/source"):
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<XCUIElementTypeApplication/>`))
		default:
			_, _ = w.Write([]byte(`{"value":"ok"}`))
		}
	}))
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeWDA) recorded() []recordedRequest {
	f.mu.Lock()
	defer f.mu.Unlock()
	out := make([]recordedRequest, len(f.requests))
	copy(out, f.requests)
	return out
}

// find returns the first recorded request whose path contains sub, or fails.
func (f *fakeWDA) find(t *testing.T, sub string) recordedRequest {
	t.Helper()
	for _, req := range f.recorded() {
		if strings.Contains(req.Path, sub) {
			return req
		}
	}
	t.Fatalf("no recorded request with path containing %q; got %+v", sub, f.recorded())
	return recordedRequest{}
}

// uiRouter builds a gin engine with the exported ui handlers registered under
// paths matching the real routes, behind a fake device middleware. Requests are
// pointed at wdaURL via the ?wdaUrl query param, so no real device is involved.
func uiRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(api.IOS_KEY, ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: "ui-test-udid"}})
		c.Next()
	})
	r.POST("/ui/tap", api.UITap)
	r.POST("/ui/swipe", api.UISwipe)
	r.POST("/ui/longpress", api.UILongPress)
	r.POST("/ui/type", api.UIType)
	r.POST("/ui/button", api.UIButton)
	r.GET("/ui/screenshot", api.UIScreenshot)
	r.GET("/ui/source", api.UISource)
	r.GET("/ui/size", api.UIWindowSize)
	r.GET("/ui/orientation", api.UIGetOrientation)
	r.PUT("/ui/orientation", api.UISetOrientation)
	r.GET("/ui/status", api.UIStatus)
	r.POST("/ui/app/launch", api.UIAppLaunch)
	r.POST("/ui/app/terminate", api.UIAppTerminate)
	r.POST("/ui/app/foreground", api.UIAppForeground)
	r.POST("/ui/api", api.UIAPI)
	return r
}

// doUI issues a request to path?wdaUrl=<backend> with an optional JSON body.
func doUI(t *testing.T, r *gin.Engine, method, path, wdaURL, body string) *httptest.ResponseRecorder {
	t.Helper()
	sep := "?"
	if strings.Contains(path, "?") {
		sep = "&"
	}
	url := path
	if wdaURL != "" {
		url = path + sep + "wdaUrl=" + wdaURL
	}
	var reader io.Reader
	if body != "" {
		reader = strings.NewReader(body)
	}
	req, _ := http.NewRequest(method, url, reader)
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestUITapForwardsToWDA(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/tap", wda.server.URL, `{"x":42,"y":99}`)
	assert.Equal(t, http.StatusOK, w.Code)

	req := wda.find(t, "/wda/tap/42/99")
	assert.Equal(t, http.MethodPost, req.Method)
	// A session must have been created first.
	wda.find(t, "/session")
}

func TestUISwipeForwardsBody(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/swipe", wda.server.URL, `{"x1":1,"y1":2,"x2":3,"y2":4,"duration":0.5}`)
	assert.Equal(t, http.StatusOK, w.Code)

	req := wda.find(t, "dragfromtoforduration")
	assert.Equal(t, http.MethodPost, req.Method)
	var payload map[string]float64
	require.NoError(t, json.Unmarshal([]byte(req.Body), &payload))
	assert.Equal(t, float64(1), payload["fromX"])
	assert.Equal(t, float64(4), payload["toY"])
	assert.Equal(t, 0.5, payload["duration"])
}

func TestUILongPressForwardsBody(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/longpress", wda.server.URL, `{"x":10,"y":20,"duration":1.5}`)
	assert.Equal(t, http.StatusOK, w.Code)

	req := wda.find(t, "touchAndHold")
	assert.Equal(t, http.MethodPost, req.Method)
	var payload map[string]float64
	require.NoError(t, json.Unmarshal([]byte(req.Body), &payload))
	assert.Equal(t, float64(10), payload["x"])
	assert.Equal(t, 1.5, payload["duration"])
}

func TestUITypeForwardsText(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/type", wda.server.URL, `{"text":"hello"}`)
	assert.Equal(t, http.StatusOK, w.Code)

	req := wda.find(t, "/keys")
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Contains(t, req.Body, "hello")
}

func TestUITypeRejectsMissingText(t *testing.T) {
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/type", "http://127.0.0.1:8100", `{}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "text")
}

func TestUIButtonHomeForwards(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/button", wda.server.URL, `{"name":"home"}`)
	assert.Equal(t, http.StatusOK, w.Code)

	req := wda.find(t, "homescreen")
	assert.Equal(t, http.MethodPost, req.Method)
}

func TestUIButtonUnsupportedMapsTo501(t *testing.T) {
	wda := newFakeWDA(t)
	// WDA only supports the home button; volumeup must map to 501 and never
	// reach the backend.
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/button", wda.server.URL, `{"name":"volumeup"}`)
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestUIButtonRejectsMissingName(t *testing.T) {
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/button", "http://127.0.0.1:8100", `{}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUIScreenshotReturnsPNGBytes(t *testing.T) {
	wda := newFakeWDA(t)
	raw := []byte("\x89PNG\r\n\x1a\nfake-png-bytes")
	wda.screenshotB64 = base64.StdEncoding.EncodeToString(raw)

	w := doUI(t, uiRouter(), http.MethodGet, "/ui/screenshot", wda.server.URL, "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "image/png", w.Header().Get("Content-Type"))
	assert.Equal(t, raw, w.Body.Bytes())

	wda.find(t, "/screenshot")
}

func TestUISourceForwardsAndPreservesContentType(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodGet, "/ui/source", wda.server.URL, "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "xml")
	assert.Contains(t, w.Body.String(), "XCUIElementTypeApplication")

	req := wda.find(t, "/source")
	assert.Equal(t, http.MethodGet, req.Method)
}

func TestUIWindowSizeForwards(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodGet, "/ui/size", wda.server.URL, "")
	assert.Equal(t, http.StatusOK, w.Code)

	req := wda.find(t, "/window/size")
	assert.Equal(t, http.MethodGet, req.Method)
}

func TestUIGetOrientationForwards(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodGet, "/ui/orientation", wda.server.URL, "")
	assert.Equal(t, http.StatusOK, w.Code)

	req := wda.find(t, "/orientation")
	assert.Equal(t, http.MethodGet, req.Method)
}

func TestUISetOrientationForwardsBody(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodPut, "/ui/orientation", wda.server.URL, `{"orientation":"LANDSCAPE"}`)
	assert.Equal(t, http.StatusOK, w.Code)

	req := wda.find(t, "/orientation")
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Contains(t, req.Body, "LANDSCAPE")
}

func TestUISetOrientationRejectsMissing(t *testing.T) {
	w := doUI(t, uiRouter(), http.MethodPut, "/ui/orientation", "http://127.0.0.1:8100", `{}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUIStatusForwards(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodGet, "/ui/status", wda.server.URL, "")
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ready")

	req := wda.find(t, "/status")
	assert.Equal(t, http.MethodGet, req.Method)
}

func TestUIAppLaunchForwardsBundleID(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/app/launch", wda.server.URL, `{"bundleId":"com.example.app"}`)
	assert.Equal(t, http.StatusOK, w.Code)

	req := wda.find(t, "/wda/apps/launch")
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Contains(t, req.Body, "com.example.app")
}

func TestUIAppTerminateForwardsBundleID(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/app/terminate", wda.server.URL, `{"bundleId":"com.example.app"}`)
	assert.Equal(t, http.StatusOK, w.Code)

	req := wda.find(t, "/wda/apps/terminate")
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Contains(t, req.Body, "com.example.app")
}

func TestUIAppLaunchRejectsMissingBundleID(t *testing.T) {
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/app/launch", "http://127.0.0.1:8100", `{}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUIAppForegroundUnsupportedOnWDA(t *testing.T) {
	wda := newFakeWDA(t)
	// Foreground is DeviceKit-only; on the default WDA backend it must map to 501.
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/app/foreground", wda.server.URL, `{"bundleId":"com.example.app"}`)
	assert.Equal(t, http.StatusNotImplemented, w.Code)
}

func TestUIAPIPassthroughForwards(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/api", wda.server.URL, `{"method":"GET","path":"/wda/screen"}`)
	assert.Equal(t, http.StatusOK, w.Code)

	req := wda.find(t, "/wda/screen")
	assert.Equal(t, http.MethodGet, req.Method)
}

func TestUIAPIRejectsMissingPathForWDA(t *testing.T) {
	wda := newFakeWDA(t)
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/api", wda.server.URL, `{"method":"GET"}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "path")
}

func TestUIUnreachableBackendMapsTo502(t *testing.T) {
	// Point at a port nothing is listening on; the transport failure must map
	// to 502 Bad Gateway.
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/tap", "http://127.0.0.1:1", `{"x":1,"y":1}`)
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestUIBackendErrorMapsTo502(t *testing.T) {
	// A backend that returns a non-2xx status must surface as 502.
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/session" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"sessionId":"s"}`))
			return
		}
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer backend.Close()

	w := doUI(t, uiRouter(), http.MethodPost, "/ui/tap", backend.URL, `{"x":1,"y":1}`)
	assert.Equal(t, http.StatusBadGateway, w.Code)
}

func TestUITapRejectsBadJSON(t *testing.T) {
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/tap", "http://127.0.0.1:8100", `{not json`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUIUnknownBackendRejected(t *testing.T) {
	w := doUI(t, uiRouter(), http.MethodPost, "/ui/tap?backend=bogus", "http://127.0.0.1:8100", `{"x":1,"y":1}`)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "backend")
}
