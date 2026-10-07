package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/gin-gonic/gin"
)

// newHandlerCtx builds a gin test context with a device already in context, so a
// handler's request-validation branches (which run before any device I/O) can be
// exercised without a real device.
func newHandlerCtx(method, target, body string) (*httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	c.Request = r
	c.Set(IOS_KEY, ios.DeviceEntry{})
	return w, c
}

func TestValidationRejections(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		target  string
		body    string
		handler gin.HandlerFunc
		want    int
	}{
		{"erase without confirm", "POST", "/erase", "", Erase, http.StatusBadRequest},
		{"mobilegestalt without key", "GET", "/mobilegestalt", "", GetMobileGestalt, http.StatusBadRequest},
		{"files without domain", "GET", "/files", "", ListFiles, http.StatusBadRequest},
		{"files unknown domain", "GET", "/files?domain=bogus", "", ListFiles, http.StatusBadRequest},
		{"pull without remote", "GET", "/files/pull?domain=temp", "", PullFile, http.StatusBadRequest},
		{"push without remote", "POST", "/files/push?domain=temp", "", PushFile, http.StatusBadRequest},
		{"forward without ports", "POST", "/jobs/forward", `{}`, StartForward, http.StatusBadRequest},
		{"runtest without bundle", "POST", "/jobs/runtest", `{}`, StartRunTest, http.StatusBadRequest},
		{"wifi without ssid", "PUT", "/wifi", `{"password":"x"}`, SetWifi, http.StatusBadRequest},
		{"remove wifi without ssid", "DELETE", "/wifi", "", RemoveWifi, http.StatusBadRequest},
		{"clear-passcode without token", "POST", "/mdm/clear-passcode", "", MdmClearPasscode, http.StatusBadRequest},
		{"remove crashes without args", "DELETE", "/crashes", "", RemoveCrashes, http.StatusBadRequest},
		{"set devmode bad action", "POST", "/devmode", `{"action":"bogus"}`, SetDevMode, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, c := newHandlerCtx(tc.method, tc.target, tc.body)
			tc.handler(c)
			if w.Code != tc.want {
				t.Fatalf("%s: got %d, want %d (body=%s)", tc.name, w.Code, tc.want, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), `"error"`) {
				t.Fatalf("%s: expected an error envelope, got %q", tc.name, w.Body.String())
			}
		})
	}
}

func TestJobNotFoundReturns404(t *testing.T) {
	w, c := newHandlerCtx("GET", "/jobs/nope-1", "")
	c.Params = gin.Params{{Key: "udid", Value: "UDID-X"}, {Key: "id", Value: "nope-1"}}
	GetJob(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("got %d, want 404", w.Code)
	}
}
