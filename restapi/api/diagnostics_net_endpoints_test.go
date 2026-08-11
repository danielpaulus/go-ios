package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/gin-gonic/gin"
)

// newDeviceCtx builds a gin test context with the supplied device already in
// context, so a handler's pre-I/O branches (validation / capability checks) can
// be exercised without a real device.
func newDeviceCtx(method, target string, device ios.DeviceEntry) (*httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, target, nil)
	c.Set(IOS_KEY, device)
	return w, c
}

// TestGetRsdServicesUnavailable verifies the RSD handler returns a clear 400
// error envelope when the device has no RSD provider (older iOS / no tunnel).
// This path runs entirely before any device I/O, so it is device-free.
func TestGetRsdServicesUnavailable(t *testing.T) {
	w, c := newDeviceCtx("GET", "/rsd", ios.DeviceEntry{})

	GetRsdServices(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("got %d, want %d (body=%s)", w.Code, http.StatusBadRequest, w.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response is not a JSON error envelope: %v (body=%s)", err, w.Body.String())
	}
	if resp["error"] == "" {
		t.Fatalf("expected a non-empty error field, got %q", w.Body.String())
	}
}

// TestGetRsdServicesAvailable verifies that when the device does have an RSD
// provider, the handler returns 200 with the service map. An empty
// RsdPortProviderJson is non-nil, so SupportsRsd() is true and GetServices()
// yields an empty map, exercising the success path without any device I/O.
func TestGetRsdServicesAvailable(t *testing.T) {
	device := ios.DeviceEntry{Rsd: ios.RsdPortProviderJson{}}
	w, c := newDeviceCtx("GET", "/rsd", device)

	GetRsdServices(c)

	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want %d (body=%s)", w.Code, http.StatusOK, w.Body.String())
	}
	var services map[string]ios.RsdServiceEntry
	if err := json.Unmarshal(w.Body.Bytes(), &services); err != nil {
		t.Fatalf("response is not a JSON service map: %v (body=%s)", err, w.Body.String())
	}
	if len(services) != 0 {
		t.Fatalf("expected an empty service map, got %v", services)
	}
}

// TestDiagnosticsNetRoutesRegistered ensures the new endpoints are wired into the
// full route tree without conflicting with existing routes. gin panics at
// registration on a conflict, so a successful build here proves clean wiring.
func TestDiagnosticsNetRoutesRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	device := router.Group("/device/:udid")
	registerDiagnosticsNetRoutes(device)

	want := map[string]string{
		"GET /device/:udid/diskspace":        "",
		"GET /device/:udid/ip":               "",
		"GET /device/:udid/rsd":              "",
		"GET /device/:udid/battery/registry": "",
	}
	for _, r := range router.Routes() {
		delete(want, r.Method+" "+r.Path)
	}
	if len(want) != 0 {
		t.Fatalf("missing expected routes: %v", want)
	}
}
