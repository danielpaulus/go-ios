// Package remote implements `ios remote`: a tiny, self-contained browser
// remote-control for an iOS device.
//
// Architecture (two independent halves):
//
//   - Live screen is WDA-FREE. It reuses the instruments screenshot service
//     (ios/instruments) and streams JPEG frames as an MJPEG
//     (multipart/x-mixed-replace) response at GET /screen. This needs the
//     tunnel + a mounted developer disk image, but no WebDriverAgent.
//   - Input goes through WDA, because go-ios has no native touch injection on
//     main. Rather than re-implement the WDA client, the handlers shell out to
//     the SAME `ios` binary (located via os.Executable) and reuse the proven
//     `ios ui …` commands, targeting the device udid and the configured
//     --wda-url. So the screen works with no WDA running; taps/swipes/typing
//     need a reachable WDA.
package remote

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/instruments"
)

const logModule = "go-ios/remote"

// DefaultWDAURL is the WebDriverAgent base URL used when none is supplied.
const DefaultWDAURL = "http://127.0.0.1:8100"

// Server is a minimal browser remote-control for a single device.
type Server struct {
	device ios.DeviceEntry
	udid   string
	wdaURL string

	// iosBinary is the path to the `ios` CLI used to drive input via WDA.
	iosBinary string

	// screen is the shared latest-JPEG-frame broadcaster fed by the
	// instruments screenshot loop. It is WDA-free.
	screen *screenBroadcaster

	// logical device size in WDA points, resolved lazily and cached.
	sizeMu sync.Mutex
	sizeW  float64
	sizeH  float64
}

// NewServer wires up a remote-control server for the given device. wdaURL is
// already resolved by the caller (default applied upstream). The screen half
// connects to the instruments screenshot service immediately so a failure
// (e.g. missing developer disk image) surfaces before we start listening.
func NewServer(device ios.DeviceEntry, wdaURL string) (*Server, error) {
	iosBinary, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("locating ios binary: %w", err)
	}
	udid := device.Properties.SerialNumber

	bc, err := newScreenBroadcaster(device)
	if err != nil {
		return nil, fmt.Errorf("starting screenshot service: %w", err)
	}

	return &Server{
		device:    device,
		udid:      udid,
		wdaURL:    wdaURL,
		iosBinary: iosBinary,
		screen:    bc,
	}, nil
}

// ListenAndServe binds 0.0.0.0:<port> (reachable over Tailscale) and serves the
// UI until the process is stopped.
func (s *Server) ListenAndServe(port string) error {
	location := "0.0.0.0:" + port
	golog.Info("starting remote-control server, open your browser here",
		"module", logModule, "udid", s.udid, "host", "0.0.0.0", "port", port,
		"wdaURL", s.wdaURL, "url", fmt.Sprintf("http://%s/", location))
	return http.ListenAndServe(location, s.Handler())
}

// Handler returns the HTTP handler for the remote-control UI and endpoints.
// Exposed so tests can exercise it without binding a socket.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/screen", s.handleScreen)
	mux.HandleFunc("/tap", s.handleTap)
	mux.HandleFunc("/swipe", s.handleSwipe)
	mux.HandleFunc("/type", s.handleType)
	mux.HandleFunc("/button", s.handleButton)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, indexHTML)
}

func (s *Server) handleScreen(w http.ResponseWriter, r *http.Request) {
	golog.Info("new screen stream client", "module", logModule, "udid", s.udid, "remote", r.RemoteAddr)
	frames, unsubscribe := s.screen.subscribe()
	defer unsubscribe()

	w.Header().Set("Server", "go-ios-remote")
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary="+mjpegBoundary)
	w.Header().Set("Cache-Control", "no-cache, private")
	w.Header().Set("Pragma", "no-cache")
	w.WriteHeader(http.StatusOK)

	flusher, _ := w.(http.Flusher)
	for {
		select {
		case <-r.Context().Done():
			golog.Info("screen stream client disconnected", "module", logModule, "udid", s.udid, "remote", r.RemoteAddr)
			return
		case jpg := <-frames:
			if _, err := io.WriteString(w, fmt.Sprintf(mjpegFrameHeader, len(jpg))); err != nil {
				return
			}
			if _, err := w.Write(jpg); err != nil {
				return
			}
			if _, err := io.WriteString(w, mjpegFrameFooter); err != nil {
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
	}
}

// tapRequest is the fraction-based coordinate the browser sends. x and y are
// each in [0,1], measured against the displayed image. Sending fractions keeps
// the browser oblivious to the device's pixel/point sizes and CSS scaling.
type tapRequest struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type swipeRequest struct {
	FromX float64 `json:"fromX"`
	FromY float64 `json:"fromY"`
	ToX   float64 `json:"toX"`
	ToY   float64 `json:"toY"`
}

type typeRequest struct {
	Text string `json:"text"`
}

type buttonRequest struct {
	Name string `json:"name"`
}

func (s *Server) handleTap(w http.ResponseWriter, r *http.Request) {
	var req tapRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	x, y, err := s.fractionToPoints(req.X, req.Y)
	if err != nil {
		httpError(w, err)
		return
	}
	s.runUI(w, "tap", "--x="+ftoa(x), "--y="+ftoa(y))
}

func (s *Server) handleSwipe(w http.ResponseWriter, r *http.Request) {
	var req swipeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	fx, fy, err := s.fractionToPoints(req.FromX, req.FromY)
	if err != nil {
		httpError(w, err)
		return
	}
	tx, ty, err := s.fractionToPoints(req.ToX, req.ToY)
	if err != nil {
		httpError(w, err)
		return
	}
	s.runUI(w, "swipe",
		"--from-x="+ftoa(fx), "--from-y="+ftoa(fy),
		"--to-x="+ftoa(tx), "--to-y="+ftoa(ty))
}

func (s *Server) handleType(w http.ResponseWriter, r *http.Request) {
	var req typeRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	s.runUI(w, "type", "--text="+req.Text)
}

func (s *Server) handleButton(w http.ResponseWriter, r *http.Request) {
	var req buttonRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	switch req.Name {
	case "home", "lock", "volumeup", "volumedown":
	default:
		httpError(w, fmt.Errorf("unknown button %q", req.Name))
		return
	}
	s.runUI(w, "button", req.Name)
}

// runUI shells out to `ios ui <args…>` for the configured device and WDA URL,
// returning the combined output to the browser. Input is WDA-backed; a
// non-reachable WDA surfaces here as a non-zero exit + stderr in the response.
func (s *Server) runUI(w http.ResponseWriter, args ...string) {
	full := append([]string{"ui"}, args...)
	// Always drive WDA: input is WDA-backed and we pass --wda-url, so pin the
	// driver rather than let `ios ui` auto-detect (which prefers DeviceKit).
	full = append(full, "--driver=wda", "--udid="+s.udid, "--wda-url="+s.wdaURL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, s.iosBinary, full...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		golog.Warn("ui command failed", "module", logModule, "udid", s.udid,
			"args", args, "error", err, "output", string(out))
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = fmt.Fprintf(w, "ui %v failed: %v\n%s", args, err, out)
		return
	}
	golog.Debug("ui command ok", "module", logModule, "udid", s.udid, "args", args)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(out)
}

// fractionToPoints is the single place where the browser's 0..1 fraction is
// mapped to WDA logical points. WDA taps take logical points, and its
// window/size endpoint reports the same logical point space, so a fraction of
// the screen maps linearly: point = fraction * logicalSize. The logical size is
// fetched once (via `ios ui size`) and cached.
func (s *Server) fractionToPoints(fx, fy float64) (float64, float64, error) {
	w, h, err := s.logicalSize()
	if err != nil {
		return 0, 0, err
	}
	return clamp01(fx) * w, clamp01(fy) * h, nil
}

func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// ftoa formats a coordinate for `ios ui tap/swipe`, which parse --x/--y as
// integers and reject decimals ("--x is required"). Round to the nearest point.
func ftoa(v float64) string {
	return strconv.FormatInt(int64(math.Round(v)), 10)
}

// logicalSize returns the device's WDA logical window size in points, caching
// the first successful lookup. It calls `ios ui size` and parses the WDA
// window/size response shape ({"value":{"width":W,"height":H}} or a bare
// {"width":W,"height":H}).
func (s *Server) logicalSize() (float64, float64, error) {
	s.sizeMu.Lock()
	defer s.sizeMu.Unlock()
	if s.sizeW > 0 && s.sizeH > 0 {
		return s.sizeW, s.sizeH, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, s.iosBinary, "ui", "size", "--driver=wda", "--udid="+s.udid, "--wda-url="+s.wdaURL)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, fmt.Errorf("ui size failed (is WDA reachable at %s?): %v: %s", s.wdaURL, err, out)
	}
	w, h, err := parseSize(out)
	if err != nil {
		return 0, 0, err
	}
	s.sizeW, s.sizeH = w, h
	return w, h, nil
}

// parseSize extracts width/height from the JSON emitted by `ios ui size`. It
// accepts both the WDA envelope ({"value":{"width":…,"height":…}}) and a bare
// object, and tolerates surrounding whitespace/log noise by scanning for the
// first '{'.
func parseSize(out []byte) (float64, float64, error) {
	trimmed := bytes.TrimSpace(out)
	if i := bytes.IndexByte(trimmed, '{'); i > 0 {
		trimmed = trimmed[i:]
	}
	var envelope struct {
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
		Value  struct {
			Width  float64 `json:"width"`
			Height float64 `json:"height"`
		} `json:"value"`
	}
	if err := json.Unmarshal(trimmed, &envelope); err != nil {
		return 0, 0, fmt.Errorf("parsing ui size output %q: %w", out, err)
	}
	w, h := envelope.Width, envelope.Height
	if envelope.Value.Width > 0 {
		w, h = envelope.Value.Width, envelope.Value.Height
	}
	if w <= 0 || h <= 0 {
		return 0, 0, fmt.Errorf("ui size returned non-positive dimensions from %q", out)
	}
	return w, h, nil
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return false
	}
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	if err := dec.Decode(v); err != nil {
		httpError(w, fmt.Errorf("bad request body: %w", err))
		return false
	}
	return true
}

func httpError(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusBadRequest)
}

// --- MJPEG plumbing (WDA-free screen) ---

const mjpegBoundary = "goiosremoteframe"

const (
	mjpegFrameHeader = "--" + mjpegBoundary + "\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n"
	mjpegFrameFooter = "\r\n"
)

// screenBroadcaster runs the instruments screenshot loop once and fans the
// latest JPEG frame out to all connected /screen clients. Unlike the package
// helper in ios/instruments it holds no global state, so multiple servers /
// tests don't collide.
type screenBroadcaster struct {
	device ios.DeviceEntry
	udid   string

	svcMu sync.Mutex
	svc   *instruments.ScreenshotService

	mu        sync.Mutex
	consumers map[chan []byte]struct{}
}

func newScreenBroadcaster(device ios.DeviceEntry) (*screenBroadcaster, error) {
	svc, err := instruments.NewScreenshotService(device)
	if err != nil {
		return nil, err
	}
	bc := &screenBroadcaster{
		device:    device,
		svc:       svc,
		udid:      device.Properties.SerialNumber,
		consumers: make(map[chan []byte]struct{}),
	}
	go bc.loop()
	return bc, nil
}

func (bc *screenBroadcaster) subscribe() (<-chan []byte, func()) {
	ch := make(chan []byte, 1)
	bc.mu.Lock()
	bc.consumers[ch] = struct{}{}
	bc.mu.Unlock()
	return ch, func() {
		bc.mu.Lock()
		delete(bc.consumers, ch)
		bc.mu.Unlock()
	}
}

func (bc *screenBroadcaster) broadcast(jpg []byte) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	for ch := range bc.consumers {
		// Drop the oldest frame for slow consumers so a stalled client can
		// never back-pressure the capture loop.
		select {
		case ch <- jpg:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- jpg:
			default:
			}
		}
	}
}

// loop captures screenshots (~12 fps) and broadcasts them as JPEG. The
// instruments service returns PNG on some devices and JPEG on others; we
// transcode PNG to JPEG and pass JPEG through untouched.
//
// The screenshot service can wedge (e.g. a takeScreenshot timeout while a WDA
// XCUITest session contends for the same DVT channel). Rather than stopping the
// stream for good — this server is meant to run unattended — the loop tears the
// service down and reconnects, so the mirror recovers on its own.
func (bc *screenBroadcaster) loop() {
	var opt jpeg.Options
	opt.Quality = 80
	const (
		targetInterval = 80 * time.Millisecond // ~12 fps
		reconnectDelay = 2 * time.Second
	)

	for {
		start := time.Now()
		raw, err := bc.currentService().TakeScreenshot()
		if err != nil {
			golog.Error("screenshot failed, reconnecting screen service", "module", logModule, "udid", bc.udid, "error", err)
			bc.reconnect()
			time.Sleep(reconnectDelay)
			continue
		}
		jpg, err := toJPEG(raw, &opt)
		if err != nil {
			golog.Warn("failed converting frame to jpeg", "module", logModule, "udid", bc.udid, "error", err)
			continue
		}
		bc.broadcast(jpg)
		if d := targetInterval - time.Since(start); d > 0 {
			time.Sleep(d)
		}
	}
}

func (bc *screenBroadcaster) currentService() *instruments.ScreenshotService {
	bc.svcMu.Lock()
	defer bc.svcMu.Unlock()
	return bc.svc
}

// reconnect closes the wedged screenshot service and dials a fresh one. It
// retries until it succeeds so a transient tunnel/DVT hiccup doesn't kill the
// mirror permanently.
func (bc *screenBroadcaster) reconnect() {
	bc.svcMu.Lock()
	if bc.svc != nil {
		bc.svc.Close()
		bc.svc = nil
	}
	bc.svcMu.Unlock()

	for {
		svc, err := instruments.NewScreenshotService(bc.device)
		if err == nil {
			bc.svcMu.Lock()
			bc.svc = svc
			bc.svcMu.Unlock()
			golog.Info("screen service reconnected", "module", logModule, "udid", bc.udid)
			return
		}
		golog.Warn("screen service reconnect failed, retrying", "module", logModule, "udid", bc.udid, "error", err)
		time.Sleep(2 * time.Second)
	}
}

// toJPEG returns JPEG bytes for a raw screenshot payload. If the payload is
// already JPEG (SOI marker 0xFFD8) it is returned as-is; otherwise it is
// decoded as PNG and re-encoded.
func toJPEG(raw []byte, opt *jpeg.Options) ([]byte, error) {
	if len(raw) >= 2 && raw[0] == 0xFF && raw[1] == 0xD8 {
		return raw, nil
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	bw := bufio.NewWriter(&b)
	if err := jpeg.Encode(bw, img, opt); err != nil {
		return nil, err
	}
	if err := bw.Flush(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Close releases the screenshot service.
func (s *Server) Close() {
	if s.screen == nil {
		return
	}
	s.screen.svcMu.Lock()
	defer s.screen.svcMu.Unlock()
	if s.screen.svc != nil {
		s.screen.svc.Close()
		s.screen.svc = nil
	}
}
