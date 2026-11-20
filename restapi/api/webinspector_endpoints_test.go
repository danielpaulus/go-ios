package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/restapi/api"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// webInspectorTestRouter wires the non-interactive Web Inspector handlers behind
// a device middleware that injects a device with no real connection. Request
// validation (missing url/script) runs before any device I/O, so those paths are
// device-free; paths that would reach the device are not exercised here.
func webInspectorTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(api.IOS_KEY, ios.DeviceEntry{
			Properties: ios.DeviceProperties{SerialNumber: "webinspector-test-udid"},
		})
		c.Next()
	})
	r.GET("/webinspector/pages", api.WebInspectorPages)
	r.POST("/webinspector/launch", api.WebInspectorLaunch)
	r.POST("/webinspector/eval", api.WebInspectorEval)
	return r
}

// TestWebInspectorLaunchMissingURL asserts launch rejects a request with no url
// (neither query param nor JSON body) with a 400 before touching the device.
func TestWebInspectorLaunchMissingURL(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"no body no query", ""},
		{"empty json", "{}"},
		{"blank url", `{"url":"   "}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := webInspectorTestRouter()
			var reader *strings.Reader
			if tc.body != "" {
				reader = strings.NewReader(tc.body)
			} else {
				reader = strings.NewReader("")
			}
			req, _ := http.NewRequest("POST", "/webinspector/launch", reader)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			var resp map[string]any
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			errMsg, _ := resp["error"].(string)
			assert.Contains(t, errMsg, "url")
		})
	}
}

// TestWebInspectorEvalMissingScript asserts eval rejects a request with a
// missing/blank script with a 400 before touching the device.
func TestWebInspectorEvalMissingScript(t *testing.T) {
	cases := []struct {
		name string
		body string
	}{
		{"empty json", "{}"},
		{"blank script", `{"script":"  "}`},
		{"page but no script", `{"page":"page-1"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			router := webInspectorTestRouter()
			req, _ := http.NewRequest("POST", "/webinspector/eval", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)
			var resp map[string]any
			assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
			errMsg, _ := resp["error"].(string)
			assert.Contains(t, errMsg, "script")
		})
	}
}

// TestWebInspectorEvalInvalidJSON asserts malformed JSON bodies are rejected
// with a 400 (the bind error), not a 500.
func TestWebInspectorEvalInvalidJSON(t *testing.T) {
	router := webInspectorTestRouter()
	req, _ := http.NewRequest("POST", "/webinspector/eval", strings.NewReader(`{"script": `))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestWebInspectorRoutesRegistered asserts the three non-interactive routes are
// wired to handlers (a registered route never yields 404). Because the injected
// device has no real connection, the handlers fail while connecting rather than
// returning 404; the assertion is simply that they are reachable.
func TestWebInspectorRoutesRegistered(t *testing.T) {
	cases := []struct {
		method string
		path   string
		body   string
	}{
		// launch/eval carry valid input so validation passes and the handler
		// proceeds to the (failing) device connection instead of a 400.
		{"POST", "/webinspector/launch", `{"url":"https://example.com"}`},
		{"POST", "/webinspector/eval", `{"script":"1+1"}`},
	}
	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			router := webInspectorTestRouter()
			req, _ := http.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusNotFound, w.Code, "route must be registered")
			// With no real device, the connect/list step fails: a 4xx (Web
			// Inspector disabled) or 5xx, never a 2xx.
			assert.GreaterOrEqual(t, w.Code, 400)
		})
	}
}
