package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// The accessibility/location endpoints all perform device I/O only after their
// request-validation branches. These tests exercise the device-free branches:
// bad-input (400) paths that short-circuit before any connection is attempted,
// plus multipart parsing and temp-file handling for the GPX upload.

// --- enabled parsing (VoiceOver / ZoomTouch) -------------------------------

func TestSetVoiceOverRejectsBadEnabled(t *testing.T) {
	// Non-boolean query value -> 400 before touching the device.
	w, c := deviceCtx("PUT", "/voiceover?enabled=maybe", "", "", "UDID-VO", nil)
	SetVoiceOver(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad enabled query: got %d, want 400", w.Code)
	}

	// Malformed JSON body with no query fallback -> 400.
	w2, c2 := deviceCtx("PUT", "/voiceover", "{not-json", "application/json", "UDID-VO", nil)
	SetVoiceOver(c2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("bad enabled body: got %d, want 400", w2.Code)
	}
}

func TestSetZoomTouchRejectsBadEnabled(t *testing.T) {
	w, c := deviceCtx("PUT", "/zoom?enabled=nope", "", "", "UDID-ZT", nil)
	SetZoomTouch(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("bad enabled query: got %d, want 400", w.Code)
	}
}

// TestParseEnabled covers the shared query/body precedence logic directly so we
// exercise the successful-parse branches without needing a device.
func TestParseEnabled(t *testing.T) {
	cases := []struct {
		name        string
		target      string
		body        string
		contentType string
		want        bool
		wantErr     bool
	}{
		{name: "query true", target: "/x?enabled=true", want: true},
		{name: "query false", target: "/x?enabled=false", want: false},
		{name: "query 1", target: "/x?enabled=1", want: true},
		{name: "query 0", target: "/x?enabled=0", want: false},
		{name: "query missing", target: "/x", wantErr: true},
		{name: "query invalid", target: "/x?enabled=yes", wantErr: true},
		{name: "json body true", target: "/x", body: `{"enabled":true}`, contentType: "application/json", want: true},
		{name: "json body false", target: "/x", body: `{"enabled":false}`, contentType: "application/json", want: false},
		{name: "bad body good query fallback", target: "/x?enabled=true", body: "{bad", contentType: "application/json", want: true},
		{name: "bad body no fallback", target: "/x", body: "{bad", contentType: "application/json", wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			var r *http.Request
			if tc.body != "" {
				r = httptest.NewRequest("PUT", tc.target, strings.NewReader(tc.body))
			} else {
				r = httptest.NewRequest("PUT", tc.target, nil)
			}
			if tc.contentType != "" {
				r.Header.Set("Content-Type", tc.contentType)
			}
			c.Request = r
			got, err := parseEnabled(c)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got value %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// --- AX audit timeout validation -------------------------------------------

func TestRunAXAuditRejectsBadTimeout(t *testing.T) {
	for _, q := range []string{"timeout=abc", "timeout=0", "timeout=-5"} {
		w, c := deviceCtx("POST", "/ax/audit?"+q, "", "", "UDID-AX", nil)
		RunAXAudit(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("bad %q: got %d, want 400", q, w.Code)
		}
		var resp map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("bad %q: response not JSON: %v", q, err)
		}
		if resp["error"] == "" {
			t.Fatalf("bad %q: expected error field", q)
		}
	}
}

// --- GPX upload validation & temp handling ---------------------------------

func TestSetLocationGPXMissingFile(t *testing.T) {
	// multipart form without a "gpx" file field -> 400 before any device I/O.
	ct, body := multipartBody(t, map[string]string{"unrelated": "1"}, nil)
	w, c := deviceCtx("PUT", "/setlocation/gpx", body.String(), ct, "UDID-GPX", nil)
	SetLocationGPX(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing gpx: got %d, want 400", w.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("response not JSON: %v", err)
	}
	if !strings.Contains(resp["error"], "gpx") {
		t.Fatalf("error %q should mention gpx", resp["error"])
	}
}

func TestSetLocationGPXEmptyFile(t *testing.T) {
	// A present-but-empty gpx file is rejected with 400.
	ct, body := multipartBody(t, nil, map[string][]byte{"gpx": {}})
	w, c := deviceCtx("PUT", "/setlocation/gpx", body.String(), ct, "UDID-GPX", nil)
	SetLocationGPX(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty gpx: got %d, want 400", w.Code)
	}
}

func TestSetLocationGPXNoMultipart(t *testing.T) {
	// A non-multipart request has no form file -> 400.
	w, c := deviceCtx("PUT", "/setlocation/gpx", "", "", "UDID-GPX", nil)
	SetLocationGPX(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("no multipart: got %d, want 400", w.Code)
	}
}

// TestGPXMultipartRoundTrips proves the multipart helper produces a form that
// gin/FormFile can read back the gpx bytes from (parsing path the handler relies
// on), independent of any device connection.
func TestGPXMultipartRoundTrips(t *testing.T) {
	gpx := []byte(`<gpx><trk><trkseg><trkpt lat="1" lon="2"></trkpt></trkseg></trk></gpx>`)
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("gpx", "route.gpx")
	if err != nil {
		t.Fatal(err)
	}
	fw.Write(gpx)
	mw.Close()

	_, c := deviceCtx("PUT", "/setlocation/gpx", buf.String(), mw.FormDataContentType(), "UDID-GPX", nil)
	got, err := readFormFile(c, "gpx")
	if err != nil {
		t.Fatalf("readFormFile: %v", err)
	}
	if !bytes.Equal(got, gpx) {
		t.Fatalf("gpx bytes mismatch: got %q", string(got))
	}
}
