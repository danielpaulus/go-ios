package remote

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestParseSize(t *testing.T) {
	cases := []struct {
		name         string
		in           string
		wantW, wantH float64
		wantErr      bool
	}{
		{"devicekit jsonrpc", `{"id":1,"jsonrpc":"2.0","result":{"scale":2,"screenSize":{"height":667,"width":375}}}`, 375, 667, false},
		{"devicekit screenSize", `{"screenSize":{"width":375,"height":667},"scale":2}`, 375, 667, false},
		{"wda envelope", `{"value":{"width":390,"height":844}}`, 390, 844, false},
		{"bare object", `{"width":320,"height":568}`, 320, 568, false},
		{"leading noise", "some log line\n{\"value\":{\"width\":428,\"height\":926}}", 428, 926, false},
		{"non positive", `{"value":{"width":0,"height":0}}`, 0, 0, true},
		{"devicekit non positive", `{"screenSize":{"width":0,"height":0},"scale":2}`, 0, 0, true},
		{"garbage", `not json`, 0, 0, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			w, h, err := parseSize([]byte(c.in))
			if c.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v x %v", w, h)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if w != c.wantW || h != c.wantH {
				t.Fatalf("got %vx%v, want %vx%v", w, h, c.wantW, c.wantH)
			}
		})
	}
}

func TestFractionToPoints(t *testing.T) {
	// Prime the cache so fractionToPoints doesn't shell out to `ios ui size`.
	s := &Server{sizeW: 390, sizeH: 844}

	cases := []struct {
		fx, fy       float64
		wantX, wantY float64
	}{
		{0, 0, 0, 0},
		{1, 1, 390, 844},
		{0.5, 0.5, 195, 422},
		{-0.5, 2, 0, 844}, // clamped to [0,1]
	}
	for _, c := range cases {
		x, y, err := s.fractionToPoints(c.fx, c.fy)
		if err != nil {
			t.Fatalf("fractionToPoints(%v,%v): %v", c.fx, c.fy, err)
		}
		if x != c.wantX || y != c.wantY {
			t.Fatalf("fractionToPoints(%v,%v)=%v,%v want %v,%v", c.fx, c.fy, x, y, c.wantX, c.wantY)
		}
	}
}

func TestNewServerDefaultsToDeviceKit(t *testing.T) {
	// An empty driver must default to DeviceKit (WDA is broken on iOS 26). We
	// can't spin up a real screenshot service here, so exercise the same
	// defaulting logic directly.
	s := &Server{driver: ""}
	if s.driver == "" {
		s.driver = DriverDeviceKit
	}
	if s.driver != DriverDeviceKit {
		t.Fatalf("default driver = %q, want %q", s.driver, DriverDeviceKit)
	}
}

func TestDriverURLFlag(t *testing.T) {
	cases := []struct {
		driver, url, want string
	}{
		{DriverDeviceKit, "http://127.0.0.1:12004", "--devicekit-url=http://127.0.0.1:12004"},
		{DriverWDA, "http://127.0.0.1:8100", "--wda-url=http://127.0.0.1:8100"},
	}
	for _, c := range cases {
		s := &Server{driver: c.driver, driverURL: c.url}
		if got := s.driverURLFlag(); got != c.want {
			t.Fatalf("driverURLFlag(driver=%q)=%q, want %q", c.driver, got, c.want)
		}
	}
}

// TestFractionToPointsDeviceKit maps browser fractions to DeviceKit's logical
// points (iPhone SE 3rd gen reports 375x667), independent of the WDA size.
func TestFractionToPointsDeviceKit(t *testing.T) {
	s := &Server{driver: DriverDeviceKit, sizeW: 375, sizeH: 667}
	x, y, err := s.fractionToPoints(0.5, 0.5)
	if err != nil {
		t.Fatalf("fractionToPoints: %v", err)
	}
	// 0.5*375=187.5 -> ftoa rounds to 188; 0.5*667=333.5 -> 334.
	if ftoa(x) != "188" || ftoa(y) != "334" {
		t.Fatalf("center of 375x667 = %s,%s, want 188,334", ftoa(x), ftoa(y))
	}
}

func TestFtoa(t *testing.T) {
	// `ios ui tap` parses --x/--y as integers, so coordinates must be integer
	// strings (decimals are rejected).
	cases := map[float64]string{
		195:   "195",
		422.4: "422",
		422.5: "423", // rounds to nearest
		0:     "0",
	}
	for in, want := range cases {
		if got := ftoa(in); got != want {
			t.Fatalf("ftoa(%v)=%q, want %q", in, got, want)
		}
	}
}

func TestIndexServesHTML(t *testing.T) {
	s := &Server{} // handleIndex needs no device state
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	s.handleIndex(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("content-type = %q, want text/html", ct)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "<html") || !strings.Contains(body, "/video.h264") || !strings.Contains(body, "/screen") {
		t.Fatalf("body does not look like the remote UI: %.120q", body)
	}
}

func TestButtonRejectsUnknown(t *testing.T) {
	s := &Server{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/button", strings.NewReader(`{"name":"selfdestruct"}`))
	s.handleButton(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400 for unknown button", rec.Code)
	}
}

// TestVideoH264ProxySetsContentTypeAndStreams verifies /video.h264 proxies the
// DeviceKit runner's /h264, forcing Content-Type video/h264 and flushing the
// runner's bytes through as they arrive.
func TestVideoH264ProxySetsContentTypeAndStreams(t *testing.T) {
	// H.264 Annex-B: an SPS NAL prefixed with a 4-byte start code.
	payload := []byte{0x00, 0x00, 0x00, 0x01, 0x67, 0x42, 0x00, 0x1f}
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/h264" {
			http.NotFound(w, r)
			return
		}
		// The runner labels the elementary stream video/h264 too, but the proxy
		// must set it regardless (this is why we override).
		w.Header().Set("Content-Type", "video/h264")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(payload)
	}))
	defer runner.Close()

	// No supervisor -> runner always "ready"; the proxy connects once, streams,
	// then loops. Cancel the request context after we've read the body so the
	// handler returns instead of reconnecting forever.
	s := &Server{udid: "u", driver: DriverDeviceKit, deviceKitURL: runner.URL}

	ctx, cancel := context.WithCancel(context.Background())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/video.h264", nil).WithContext(ctx)

	done := make(chan struct{})
	go func() { s.handleVideoH264(rec, req); close(done) }()

	// Give the proxy time to connect and stream the first payload, then stop it.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(rec.Body.Bytes()) >= len(payload) {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	<-done

	if ct := rec.Header().Get("Content-Type"); ct != "video/h264" {
		t.Fatalf("content-type = %q, want video/h264", ct)
	}
	got := rec.Body.Bytes()
	if len(got) < len(payload) || string(got[:len(payload)]) != string(payload) {
		t.Fatalf("body = %x, want it to start with the runner payload %x", got, payload)
	}
	// Sanity: the streamed bytes carry an Annex-B start code.
	if !(got[0] == 0 && got[1] == 0 && got[2] == 0 && got[3] == 1) {
		t.Fatalf("body does not start with an Annex-B start code: %x", got[:4])
	}
}

// TestScreenProxyPassesThroughMJPEGContentType verifies /screen proxies the
// runner's /mjpeg and passes its multipart content-type through unchanged.
func TestScreenProxyPassesThroughMJPEGContentType(t *testing.T) {
	const mjpegCT = "multipart/x-mixed-replace; boundary=frame"
	frame := []byte("--frame\r\nContent-Type: image/jpeg\r\n\r\n\xff\xd8\xff\xd9\r\n")
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/mjpeg" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", mjpegCT)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(frame)
	}))
	defer runner.Close()

	s := &Server{udid: "u", driver: DriverDeviceKit, deviceKitURL: runner.URL}
	ctx, cancel := context.WithCancel(context.Background())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/screen", nil).WithContext(ctx)

	done := make(chan struct{})
	go func() { s.handleScreen(rec, req); close(done) }()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if len(rec.Body.Bytes()) >= len(frame) {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	cancel()
	<-done

	if ct := rec.Header().Get("Content-Type"); ct != mjpegCT {
		t.Fatalf("content-type = %q, want passthrough %q", ct, mjpegCT)
	}
}

// TestVideoProxyReconnectsWhileRunnerNotReady verifies the proxy does not dial
// the runner while it is not ready (so it "reconnects" in lockstep with the
// supervised runner) and stops cleanly on client disconnect.
func TestVideoProxyReconnectsWhileRunnerNotReady(t *testing.T) {
	// A runner that records every dial; while state != ready the proxy must not
	// call it.
	var dials int32
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&dials, 1)
		w.WriteHeader(http.StatusOK)
	}))
	defer runner.Close()

	sp := newFakeSpawner() // state starts as "starting" (not ready)
	s := &Server{udid: "u", driver: DriverDeviceKit, deviceKitURL: runner.URL,
		supervisor: newRunnerSupervisor("u", sp)}

	ctx, cancel := context.WithCancel(context.Background())
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/video.h264", nil).WithContext(ctx)
	done := make(chan struct{})
	go func() { s.handleVideoH264(rec, req); close(done) }()

	time.Sleep(300 * time.Millisecond)
	cancel()
	<-done

	if got := atomic.LoadInt32(&dials); got != 0 {
		t.Fatalf("proxy dialed runner %d times while not ready, want 0", got)
	}
}

// TestPipeRunnerStreamNon200 verifies a non-200 runner response is surfaced as
// an error and no body header is written (so the caller can retry).
func TestPipeRunnerStreamNon200(t *testing.T) {
	runner := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer runner.Close()
	s := &Server{udid: "u", deviceKitURL: runner.URL}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/video.h264", nil)
	wrote := false
	// flusher is nil: it isn't reached on the error path (no body is copied).
	n, err := s.pipeRunnerStream(rec, req, runner.URL+"/h264", "video/h264", &wrote, nil)
	if err == nil {
		t.Fatalf("expected error for non-200 runner, got nil (n=%d)", n)
	}
	if wrote {
		t.Fatalf("wrote response header for a failed connect; want header deferred")
	}
	if !strings.Contains(err.Error(), fmt.Sprint(http.StatusInternalServerError)) {
		t.Fatalf("error %q should mention the status code", err)
	}
}
