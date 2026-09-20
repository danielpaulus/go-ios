package uidriver_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/ios/uidriver"
)

// capturedRequest records what a test server received.
type capturedRequest struct {
	method string
	path   string
	query  string
	body   []byte
}

// newRecordingServer returns a server that records the last request and replies
// with the given status/body. Requests to the WDA /session endpoint always get
// a session id so session-scoped actions can proceed.
func newRecordingServer(t *testing.T, status int, respBody string) (*httptest.Server, *capturedRequest) {
	t.Helper()
	captured := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		// The WDA session-creation POST is bookkeeping; answer it and don't
		// overwrite the recorded action request.
		if r.URL.Path == "/session" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"sessionId":"sess-123"}`))
			return
		}
		captured.method = r.Method
		captured.path = r.URL.Path
		captured.query = r.URL.RawQuery
		captured.body = body
		w.WriteHeader(status)
		_, _ = w.Write([]byte(respBody))
	}))
	t.Cleanup(srv.Close)
	return srv, captured
}

func newDriver(t *testing.T, backend uidriver.Backend, baseURL string, opts ...uidriver.Option) *uidriver.Driver {
	t.Helper()
	d, err := uidriver.New(backend, baseURL, opts...)
	if err != nil {
		t.Fatalf("New(%s): %v", backend, err)
	}
	return d
}

func decodeJSONBody(t *testing.T, body []byte) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatalf("body is not JSON object: %v (%s)", err, body)
	}
	return m
}

func TestNewValidation(t *testing.T) {
	if _, err := uidriver.New("bogus", "http://x"); err == nil {
		t.Fatal("expected error for unknown backend")
	}
	if _, err := uidriver.New(uidriver.BackendWDA, "  "); err == nil {
		t.Fatal("expected error for empty baseURL")
	}
	d := newDriver(t, uidriver.BackendWDA, "http://127.0.0.1:8100/")
	if d.BaseURL() != "http://127.0.0.1:8100" {
		t.Fatalf("trailing slash not trimmed: %q", d.BaseURL())
	}
	if d.Backend() != uidriver.BackendWDA {
		t.Fatalf("unexpected backend %q", d.Backend())
	}
}

func TestDeviceKitTap(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{"result":"ok"}`)
	d := newDriver(t, uidriver.BackendDeviceKit, srv.URL)

	resp, err := d.Tap(10, 20)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	if captured.method != http.MethodPost || captured.path != "/rpc" {
		t.Fatalf("got %s %s, want POST /rpc", captured.method, captured.path)
	}
	rpc := decodeJSONBody(t, captured.body)
	if rpc["jsonrpc"] != "2.0" || rpc["method"] != "device.io.tap" {
		t.Fatalf("unexpected rpc envelope: %v", rpc)
	}
	params, _ := rpc["params"].(map[string]interface{})
	if params["x"].(float64) != 10 || params["y"].(float64) != 20 {
		t.Fatalf("unexpected params: %v", params)
	}
}

func TestWDATap(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{"value":null}`)
	d := newDriver(t, uidriver.BackendWDA, srv.URL)

	if _, err := d.Tap(5, 6); err != nil {
		t.Fatal(err)
	}
	if captured.method != http.MethodPost {
		t.Fatalf("method = %s", captured.method)
	}
	if captured.path != "/session/sess-123/wda/tap/5/6" {
		t.Fatalf("path = %s", captured.path)
	}
	if d.SessionID() != "sess-123" {
		t.Fatalf("session id not stored: %q", d.SessionID())
	}
}

func TestWDASwipe(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendWDA, srv.URL)

	if _, err := d.Swipe(1, 2, 3, 4, 0.5); err != nil {
		t.Fatal(err)
	}
	if captured.path != "/session/sess-123/wda/dragfromtoforduration" {
		t.Fatalf("path = %s", captured.path)
	}
	body := decodeJSONBody(t, captured.body)
	if body["fromX"].(float64) != 1 || body["toY"].(float64) != 4 || body["duration"].(float64) != 0.5 {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestDeviceKitSwipe(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendDeviceKit, srv.URL)

	if _, err := d.Swipe(1, 2, 3, 4, 1.5); err != nil {
		t.Fatal(err)
	}
	rpc := decodeJSONBody(t, captured.body)
	if rpc["method"] != "device.io.swipe" {
		t.Fatalf("method = %v", rpc["method"])
	}
}

func TestWDAType(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendWDA, srv.URL)

	if _, err := d.Type("hello"); err != nil {
		t.Fatal(err)
	}
	if captured.path != "/session/sess-123/keys" {
		t.Fatalf("path = %s", captured.path)
	}
	body := decodeJSONBody(t, captured.body)
	if body["text"] != "hello" {
		t.Fatalf("text = %v", body["text"])
	}
	value, ok := body["value"].([]interface{})
	if !ok || len(value) != 1 || value[0] != "hello" {
		t.Fatalf("value = %v", body["value"])
	}
}

func TestDeviceKitType(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendDeviceKit, srv.URL)

	if _, err := d.Type("world"); err != nil {
		t.Fatal(err)
	}
	rpc := decodeJSONBody(t, captured.body)
	params := rpc["params"].(map[string]interface{})
	if rpc["method"] != "device.io.text" || params["text"] != "world" {
		t.Fatalf("unexpected: %v", rpc)
	}
}

func TestWDAPressButtonHome(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendWDA, srv.URL)

	if _, err := d.PressButton("HOME"); err != nil {
		t.Fatal(err)
	}
	if captured.path != "/session/sess-123/wda/homescreen" {
		t.Fatalf("path = %s", captured.path)
	}
}

func TestWDAPressButtonUnsupported(t *testing.T) {
	srv, _ := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendWDA, srv.URL)

	_, err := d.PressButton("volumeup")
	if !errors.Is(err, uidriver.ErrButtonUnsupported) {
		t.Fatalf("want ErrButtonUnsupported, got %v", err)
	}
}

func TestDeviceKitPressButton(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendDeviceKit, srv.URL)

	if _, err := d.PressButton("lock"); err != nil {
		t.Fatal(err)
	}
	rpc := decodeJSONBody(t, captured.body)
	params := rpc["params"].(map[string]interface{})
	if rpc["method"] != "device.io.button" || params["button"] != "lock" {
		t.Fatalf("unexpected: %v", rpc)
	}
}

func TestWDAOrientationGetSet(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{"value":"PORTRAIT"}`)
	d := newDriver(t, uidriver.BackendWDA, srv.URL)

	if _, err := d.Orientation(); err != nil {
		t.Fatal(err)
	}
	if captured.method != http.MethodGet || captured.path != "/session/sess-123/orientation" {
		t.Fatalf("get: %s %s", captured.method, captured.path)
	}

	if _, err := d.SetOrientation("LANDSCAPE"); err != nil {
		t.Fatal(err)
	}
	if captured.method != http.MethodPost || captured.path != "/session/sess-123/orientation" {
		t.Fatalf("set: %s %s", captured.method, captured.path)
	}
	body := decodeJSONBody(t, captured.body)
	if body["orientation"] != "LANDSCAPE" {
		t.Fatalf("orientation = %v", body["orientation"])
	}
}

func TestDeviceKitOrientation(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendDeviceKit, srv.URL)

	if _, err := d.SetOrientation("PORTRAIT"); err != nil {
		t.Fatal(err)
	}
	rpc := decodeJSONBody(t, captured.body)
	if rpc["method"] != "device.io.orientation.set" {
		t.Fatalf("method = %v", rpc["method"])
	}
}

func TestWDAApp(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendWDA, srv.URL)

	if _, err := d.AppLaunch("com.example.app"); err != nil {
		t.Fatal(err)
	}
	if captured.path != "/session/sess-123/wda/apps/launch" {
		t.Fatalf("launch path = %s", captured.path)
	}
	body := decodeJSONBody(t, captured.body)
	if body["bundleId"] != "com.example.app" {
		t.Fatalf("bundleId = %v", body["bundleId"])
	}

	if _, err := d.AppTerminate("com.example.app"); err != nil {
		t.Fatal(err)
	}
	if captured.path != "/session/sess-123/wda/apps/terminate" {
		t.Fatalf("terminate path = %s", captured.path)
	}
}

func TestDeviceKitApp(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendDeviceKit, srv.URL)

	if _, err := d.AppLaunch("com.example.app"); err != nil {
		t.Fatal(err)
	}
	rpc := decodeJSONBody(t, captured.body)
	params := rpc["params"].(map[string]interface{})
	if rpc["method"] != "device.apps.launch" || params["bundleId"] != "com.example.app" {
		t.Fatalf("unexpected: %v", rpc)
	}
}

func TestAppForegroundUnsupportedOnWDA(t *testing.T) {
	srv, _ := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendWDA, srv.URL)

	_, err := d.AppForeground()
	if !errors.Is(err, uidriver.ErrForegroundUnsupported) {
		t.Fatalf("want ErrForegroundUnsupported, got %v", err)
	}
}

func TestDeviceKitAppForeground(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendDeviceKit, srv.URL)

	if _, err := d.AppForeground(); err != nil {
		t.Fatal(err)
	}
	rpc := decodeJSONBody(t, captured.body)
	if rpc["method"] != "device.apps.foreground" {
		t.Fatalf("method = %v", rpc["method"])
	}
}

func TestStatus(t *testing.T) {
	// DeviceKit hits /health.
	dkSrv, dkCaptured := newRecordingServer(t, http.StatusOK, `{"status":"ok"}`)
	dk := newDriver(t, uidriver.BackendDeviceKit, dkSrv.URL)
	resp, err := dk.Status()
	if err != nil {
		t.Fatal(err)
	}
	if dkCaptured.method != http.MethodGet || dkCaptured.path != "/health" {
		t.Fatalf("devicekit status: %s %s", dkCaptured.method, dkCaptured.path)
	}
	parsed := decodeJSONBody(t, resp.Body)
	if parsed["status"] != "ok" {
		t.Fatalf("status body = %v", parsed)
	}

	// WDA hits /status.
	wdaSrv, wdaCaptured := newRecordingServer(t, http.StatusOK, `{"value":{"state":"success"}}`)
	wda := newDriver(t, uidriver.BackendWDA, wdaSrv.URL)
	if _, err := wda.Status(); err != nil {
		t.Fatal(err)
	}
	if wdaCaptured.path != "/status" {
		t.Fatalf("wda status path = %s", wdaCaptured.path)
	}
}

func TestScreenshotDecodesBase64(t *testing.T) {
	raw := []byte("this-is-a-png")
	encoded := base64.StdEncoding.EncodeToString(raw)
	body, _ := json.Marshal(map[string]interface{}{"value": encoded})
	srv, _ := newRecordingServer(t, http.StatusOK, string(body))
	d := newDriver(t, uidriver.BackendDeviceKit, srv.URL)

	img, err := d.Screenshot()
	if err != nil {
		t.Fatal(err)
	}
	if string(img) != string(raw) {
		t.Fatalf("decoded image mismatch: %q", img)
	}
}

func TestScreenshotNoImage(t *testing.T) {
	srv, _ := newRecordingServer(t, http.StatusOK, `{"value":{}}`)
	d := newDriver(t, uidriver.BackendDeviceKit, srv.URL)

	if _, err := d.Screenshot(); err == nil {
		t.Fatal("expected error when no base64 image present")
	}
}

func TestSourceAndWindowSize(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{"value":"<xml/>"}`)
	d := newDriver(t, uidriver.BackendWDA, srv.URL)

	if _, err := d.Source(); err != nil {
		t.Fatal(err)
	}
	if captured.path != "/session/sess-123/source" {
		t.Fatalf("source path = %s", captured.path)
	}
	if _, err := d.WindowSize(); err != nil {
		t.Fatal(err)
	}
	if captured.path != "/session/sess-123/window/size" {
		t.Fatalf("window size path = %s", captured.path)
	}
}

func TestAPIPassthroughWDA(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendWDA, srv.URL)

	if _, err := d.API(uidriver.APIRequest{Method: "post", Path: "/wda/deactivateApp", Body: []byte(`{"duration":2}`)}); err != nil {
		t.Fatal(err)
	}
	if captured.method != http.MethodPost || captured.path != "/wda/deactivateApp" {
		t.Fatalf("api: %s %s", captured.method, captured.path)
	}
	if string(captured.body) != `{"duration":2}` {
		t.Fatalf("body = %s", captured.body)
	}

	// Missing path is an error.
	if _, err := d.API(uidriver.APIRequest{}); err == nil {
		t.Fatal("expected error for missing path")
	}
}

func TestAPIPassthroughDeviceKit(t *testing.T) {
	srv, captured := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendDeviceKit, srv.URL)

	if _, err := d.API(uidriver.APIRequest{RPCMethod: "device.custom", RPCParams: map[string]interface{}{"k": "v"}}); err != nil {
		t.Fatal(err)
	}
	rpc := decodeJSONBody(t, captured.body)
	params := rpc["params"].(map[string]interface{})
	if rpc["method"] != "device.custom" || params["k"] != "v" {
		t.Fatalf("unexpected: %v", rpc)
	}

	// Missing rpc method is an error.
	if _, err := d.API(uidriver.APIRequest{}); err == nil {
		t.Fatal("expected error for missing rpc method")
	}
}

func TestHTTPErrorOnNon2xx(t *testing.T) {
	srv, _ := newRecordingServer(t, http.StatusBadRequest, `{"error":"bad"}`)
	d := newDriver(t, uidriver.BackendDeviceKit, srv.URL)

	_, err := d.Tap(1, 1)
	var httpErr *uidriver.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("want *HTTPError, got %v", err)
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", httpErr.StatusCode)
	}
	if !strings.Contains(httpErr.Error(), "bad") {
		t.Fatalf("error message missing body: %s", httpErr.Error())
	}
}

func TestHealthy(t *testing.T) {
	okSrv, _ := newRecordingServer(t, http.StatusOK, `{}`)
	if !newDriver(t, uidriver.BackendDeviceKit, okSrv.URL).Healthy() {
		t.Fatal("expected healthy")
	}
	badSrv, _ := newRecordingServer(t, http.StatusInternalServerError, `{}`)
	if newDriver(t, uidriver.BackendWDA, badSrv.URL).Healthy() {
		t.Fatal("expected unhealthy on 5xx")
	}
	if newDriver(t, uidriver.BackendWDA, "http://127.0.0.1:1").Healthy() {
		t.Fatal("expected unhealthy on connection error")
	}
}

func TestStreamPipesChunkedBody(t *testing.T) {
	chunks := []string{"frame-1", "frame-2", "frame-3"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mjpeg" {
			t.Errorf("stream path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("fps") != "30" || r.URL.Query().Get("quality") != "80" {
			t.Errorf("stream query = %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "multipart/x-mixed-replace")
		w.WriteHeader(http.StatusOK)
		flusher, _ := w.(http.Flusher)
		for _, c := range chunks {
			_, _ = w.Write([]byte(c))
			if flusher != nil {
				flusher.Flush()
			}
		}
	}))
	defer srv.Close()

	d := newDriver(t, uidriver.BackendDeviceKit, srv.URL)
	body, err := d.Stream(context.Background(), uidriver.StreamOptions{FPS: "30", Quality: "80"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = body.Close() }()
	got, err := io.ReadAll(body)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != strings.Join(chunks, "") {
		t.Fatalf("stream body = %q", got)
	}
}

func TestStreamH264RejectedOnWDA(t *testing.T) {
	srv, _ := newRecordingServer(t, http.StatusOK, `{}`)
	d := newDriver(t, uidriver.BackendWDA, srv.URL)

	_, err := d.Stream(context.Background(), uidriver.StreamOptions{H264: true})
	if !errors.Is(err, uidriver.ErrStreamUnsupported) {
		t.Fatalf("want ErrStreamUnsupported, got %v", err)
	}
}

func TestStreamNon2xxReturnsHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte("no stream"))
	}))
	defer srv.Close()

	d := newDriver(t, uidriver.BackendDeviceKit, srv.URL)
	_, err := d.Stream(context.Background(), uidriver.StreamOptions{})
	var httpErr *uidriver.HTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("want *HTTPError, got %v", err)
	}
}

func TestWithSessionIDSkipsCreation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/session" {
			t.Errorf("session should not be created when WithSessionID is set")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	d := newDriver(t, uidriver.BackendWDA, srv.URL, uidriver.WithSessionID("preset"))
	if _, err := d.Tap(1, 1); err != nil {
		t.Fatal(err)
	}
	if d.SessionID() != "preset" {
		t.Fatalf("session id = %q", d.SessionID())
	}
}
