package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
)

// The pcap endpoint streams live packets straight from the device's pcapd
// service via pcap.Stream, which opens a real device connection. That device
// I/O cannot be mocked at the endpoint layer (pcap.Stream connects internally
// and is not injectable), so these tests cover the device-free surface:
//   - the ?timeout= parse/validation branches that short-circuit with 400
//     before any connection is attempted,
//   - that a valid/oversized timeout is accepted (it passes validation and
//     fails only later at connect time, i.e. never returns 400),
//   - and that the route is wired up.
//
// The actual capture-to-writer behavior (valid pcap bytes, stop on ctx cancel,
// no goroutine leak) is verified against a fake connection in
// ios/pcap/pcap_test.go.

// TestPcapRejectsBadTimeout: malformed / non-positive timeouts are rejected
// with a 400 JSON error before touching the device.
func TestPcapRejectsBadTimeout(t *testing.T) {
	for _, q := range []string{"timeout=abc", "timeout=0", "timeout=-5", "timeout=1.5", "timeout="} {
		if q == "timeout=" {
			// empty value means "use default" and must NOT 400; skip here and
			// cover it in the accept test.
			continue
		}
		w, c := deviceCtx("GET", "/pcap?"+q, "", "", "UDID-PCAP", nil)
		Pcap(c)
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

// TestPcapAcceptsValidTimeout: a well-formed timeout (including one above the
// cap, and the no-timeout default) passes validation. There is no device, so
// pcap.Stream fails at connect time and the handler returns 500 (not 400),
// proving the parse/cap branch accepted the value.
func TestPcapAcceptsValidTimeout(t *testing.T) {
	for _, target := range []string{"/pcap", "/pcap?timeout=5", "/pcap?timeout=999999"} {
		w, c := deviceCtx("GET", target, "", "", "UDID-PCAP", nil)
		Pcap(c)
		if w.Code == http.StatusBadRequest {
			t.Fatalf("%q: got 400, want the request to pass validation", target)
		}
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("%q: got %d, want 500 (connect failure, no device)", target, w.Code)
		}
	}
}

// TestPcapContentTypeConstant guards the media type clients rely on to
// recognize the stream.
func TestPcapContentTypeConstant(t *testing.T) {
	if pcapContentType != "application/vnd.tcpdump.pcap" {
		t.Fatalf("unexpected pcap content type %q", pcapContentType)
	}
}

// TestPcapRouteRegistered ensures GET /device/:udid/pcap is wired into the
// route tree (in addition to the full-tree conflict check in routes_test.go).
func TestPcapRouteRegistered(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	registerRoutes(v1, 0, 0)

	found := false
	for _, ri := range router.Routes() {
		if ri.Method == http.MethodGet && ri.Path == "/api/v1/device/:udid/pcap" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("GET /device/:udid/pcap route not registered")
	}
}
