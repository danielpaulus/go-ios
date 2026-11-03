// Package remote implements `ios remote`: a tiny, self-contained browser
// remote-control for an iOS device.
//
// Architecture: ONE supervised DeviceKit runner backs BOTH the screen and
// input, so `ios remote` needs no WebDriverAgent and no instruments DTX
// channel.
//
//   - Live screen is a passthrough of the DeviceKit runner's hardware video:
//     GET /video.h264 proxies the runner's /h264 (an H.264 Annex-B elementary
//     stream, hardware-encoded and delta-compressed — a few KB/s), and GET
//     /screen proxies the runner's /mjpeg as a browser-native fallback. Bytes
//     are flushed through as they arrive; the proxy reconnects with backoff so
//     the screen self-heals when the supervised runner respawns.
//   - Input goes through the same DeviceKit runner. go-ios has no native touch
//     injection here, so rather than re-implement the driver client the handlers
//     shell out to the SAME `ios` binary (located via os.Executable) and reuse
//     the proven `ios ui …` commands, targeting the device udid and the
//     configured driver/URL. (WDA is still selectable with --driver=wda for
//     input; the video proxy always targets the DeviceKit URL.)
package remote

import (
	"context"
	"encoding/json"
	"fmt"
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
)

const logModule = "go-ios/remote"

const (
	// DriverDeviceKit and DriverWDA are the supported input drivers. DeviceKit
	// is the default because WDA is broken on iOS 26 while DeviceKit works.
	DriverDeviceKit = "devicekit"
	DriverWDA       = "wda"

	// DefaultWDAURL is the WebDriverAgent base URL used when none is supplied.
	DefaultWDAURL = "http://127.0.0.1:8100"
	// DefaultDeviceKitURL is the DeviceKit base URL used when none is supplied.
	DefaultDeviceKitURL = "http://127.0.0.1:12004"

	// DefaultFPS and DefaultBitrate are the frame rate and bitrate requested from
	// the DeviceKit runner's /h264 endpoint. The runner's own default is ~27fps;
	// asking for 60 (the runner clamps to what the hardware encoder delivers,
	// ~45-50fps in practice) makes the browser mirror noticeably smoother.
	DefaultFPS     = 60
	DefaultBitrate = 8000000
)

// Server is a minimal browser remote-control for a single device.
type Server struct {
	device ios.DeviceEntry
	udid   string

	// driver is the input backend ("devicekit" or "wda") and driverURL is its
	// base URL, passed through to `ios ui … --driver=… --{devicekit,wda}-url=…`.
	driver    string
	driverURL string

	// deviceKitURL is the DeviceKit runner base URL the screen video is proxied
	// from (its /h264 and /mjpeg endpoints), independent of the input driver.
	deviceKitURL string

	// fps and bitrate are passed through to the runner's /h264 as query params to
	// raise the frame rate/quality above the runner's low defaults.
	fps     int
	bitrate int

	// iosBinary is the path to the `ios` CLI used to drive input.
	iosBinary string

	// supervisor spawns and auto-respawns the input runner (`ios ui run
	// devicekit`) when --manage-runner is on. nil when the runner is externally
	// managed (--manage-runner=false or --driver=wda), in which case input
	// endpoints assume the driver is always reachable.
	supervisor *runnerSupervisor

	// logical device size in the driver's points, resolved lazily and cached.
	sizeMu sync.Mutex
	sizeW  float64
	sizeH  float64
}

// Config configures a remote-control Server.
type Config struct {
	// Driver is the input backend: "devicekit" (default) or "wda". DriverURL is
	// that driver's base URL (defaults applied upstream).
	Driver    string
	DriverURL string
	// DeviceKitURL is the DeviceKit runner base URL the screen video is proxied
	// from. Defaults to DriverURL when the driver is DeviceKit, otherwise to
	// DefaultDeviceKitURL.
	DeviceKitURL string
	// ManageRunner, when true and Driver is DeviceKit, makes the Server spawn and
	// supervise the input runner (`ios ui run devicekit`) itself, auto-respawning
	// it across the intermittent testmanagerd DTX EOF disconnects. When false the
	// runner is assumed to be started externally at DriverURL (today's behavior).
	ManageRunner bool

	// FPS and Bitrate are forwarded to the DeviceKit runner's /h264 endpoint as
	// query params. Zero values fall back to DefaultFPS / DefaultBitrate.
	FPS     int
	Bitrate int
}

// NewServer wires up a remote-control server for the given device from cfg.
// Screen and input are both served by the DeviceKit runner; when
// cfg.ManageRunner is set (and the driver is DeviceKit) the Server supervises
// that runner (started by Run) and the video proxy reconnects as it respawns.
func NewServer(device ios.DeviceEntry, cfg Config) (*Server, error) {
	iosBinary, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("locating ios binary: %w", err)
	}
	driver := cfg.Driver
	if driver == "" {
		driver = DriverDeviceKit
	}
	udid := device.Properties.SerialNumber

	deviceKitURL := cfg.DeviceKitURL
	if deviceKitURL == "" {
		if driver == DriverDeviceKit {
			deviceKitURL = cfg.DriverURL
		}
		if deviceKitURL == "" {
			deviceKitURL = DefaultDeviceKitURL
		}
	}

	fps := cfg.FPS
	if fps <= 0 {
		fps = DefaultFPS
	}
	bitrate := cfg.Bitrate
	if bitrate <= 0 {
		bitrate = DefaultBitrate
	}

	s := &Server{
		device:       device,
		udid:         udid,
		driver:       driver,
		driverURL:    cfg.DriverURL,
		deviceKitURL: deviceKitURL,
		fps:          fps,
		bitrate:      bitrate,
		iosBinary:    iosBinary,
	}

	// Supervision only makes sense for the DeviceKit runner we know how to spawn.
	if cfg.ManageRunner && driver == DriverDeviceKit {
		s.supervisor = newRunnerSupervisor(udid, &execRunnerSpawner{
			iosBinary: iosBinary,
			udid:      udid,
			healthURL: cfg.DriverURL + "/health",
			stdout:    os.Stdout,
			stderr:    os.Stderr,
		})
	}
	return s, nil
}

// driverURLFlag returns the `ios ui` flag that carries driverURL for the
// configured driver (--devicekit-url for DeviceKit, --wda-url for WDA).
func (s *Server) driverURLFlag() string {
	if s.driver == DriverWDA {
		return "--wda-url=" + s.driverURL
	}
	return "--devicekit-url=" + s.driverURL
}

// Run binds 0.0.0.0:<port> (reachable over Tailscale), supervises the input
// runner (when managed), and serves the UI until ctx is cancelled (SIGINT/
// SIGTERM). On shutdown it stops the HTTP server and terminates the supervised
// runner so no `ios ui run devicekit` child is orphaned.
func (s *Server) Run(ctx context.Context, port string) error {
	location := "0.0.0.0:" + port
	golog.Info("starting remote-control server, open your browser here",
		"module", logModule, "udid", s.udid, "host", "0.0.0.0", "port", port,
		"driver", s.driver, "driverURL", s.driverURL, "deviceKitURL", s.deviceKitURL,
		"manageRunner", s.supervisor != nil, "url", fmt.Sprintf("http://%s/", location))

	// Supervise the input runner in the background; run() returns when ctx is
	// cancelled, having stopped the child.
	var wg sync.WaitGroup
	if s.supervisor != nil {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.supervisor.run(ctx)
		}()
	}

	srv := &http.Server{Addr: location, Handler: s.Handler()}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		golog.Info("remote-control server shutting down", "module", logModule, "udid", s.udid)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
		wg.Wait() // wait for the runner to be terminated before returning
		return nil
	}
}

// Handler returns the HTTP handler for the remote-control UI and endpoints.
// Exposed so tests can exercise it without binding a socket.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/status", s.handleStatus)
	mux.HandleFunc("/healthz", s.handleStatus)
	mux.HandleFunc("/video.h264", s.handleVideoH264)
	mux.HandleFunc("/screen", s.handleScreen)
	mux.HandleFunc("/tap", s.handleTap)
	mux.HandleFunc("/swipe", s.handleSwipe)
	mux.HandleFunc("/type", s.handleType)
	mux.HandleFunc("/button", s.handleButton)
	return mux
}

// runnerStateString returns the current input-runner lifecycle state. When the
// runner is externally managed (no supervisor) input is always assumed usable,
// reported as "ready".
func (s *Server) runnerStateString() runnerState {
	if s.supervisor == nil {
		return runnerReady
	}
	return s.supervisor.State()
}

// handleStatus reports the input-runner lifecycle so the UI can show
// "reconnecting…" and health checks can gate on readiness.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	state := s.runnerStateString()
	w.Header().Set("Content-Type", "application/json")
	if state != runnerReady {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(map[string]any{
		"runnerState":  string(state),
		"manageRunner": s.supervisor != nil,
		"driver":       s.driver,
	})
}

// inputReady reports whether input can be dispatched; when not, it writes a 503
// with a clear JSON body and returns false so the handler stops.
func (s *Server) inputReady(w http.ResponseWriter) bool {
	if state := s.runnerStateString(); state != runnerReady {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"error":       "input runner starting, retry shortly",
			"runnerState": string(state),
		})
		return false
	}
	return true
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

// handleVideoH264 streams the DeviceKit runner's hardware H.264 elementary
// stream (Annex-B) to the browser, which decodes it via WebCodecs. It is the
// primary, efficient screen source. The runner's low default frame rate is
// raised by forwarding the configured fps/bitrate as query params (the runner
// honors them; the hardware encoder clamps to what it can actually deliver).
func (s *Server) handleVideoH264(w http.ResponseWriter, r *http.Request) {
	path := "/h264?fps=" + strconv.Itoa(s.fps) + "&bitrate=" + strconv.Itoa(s.bitrate)
	s.proxyRunnerStream(w, r, path, "video/h264")
}

// handleScreen streams the DeviceKit runner's MJPEG to the browser as the
// WebCodecs-less fallback screen source.
func (s *Server) handleScreen(w http.ResponseWriter, r *http.Request) {
	s.proxyRunnerStream(w, r, "/mjpeg", "")
}

// proxyRunnerStream passthrough-proxies one of the DeviceKit runner's streaming
// endpoints (path) to the browser, flushing bytes as they arrive (never
// buffering the whole stream). It reconnects to the runner with a small backoff
// so the screen self-heals when the supervised runner respawns, and stops when
// the browser disconnects. contentType, when non-empty, overrides the runner's
// (H.264 needs an explicit video/h264); otherwise the runner's is passed
// through (MJPEG carries its own multipart boundary).
func (s *Server) proxyRunnerStream(w http.ResponseWriter, r *http.Request, path, contentType string) {
	target := s.deviceKitURL + path
	golog.Info("new screen stream client", "module", logModule, "udid", s.udid,
		"remote", r.RemoteAddr, "path", path, "target", target)

	flusher, _ := w.(http.Flusher)
	wroteHeader := false
	backoff := 250 * time.Millisecond
	const backoffMax = 2 * time.Second

	for {
		if r.Context().Err() != nil {
			return
		}
		// Only attempt to stream when the runner is ready; otherwise wait so the
		// screen "reconnects" in lockstep with the supervised runner respawning.
		if s.runnerStateString() != runnerReady {
			if !sleepCtx(r.Context(), backoff) {
				return
			}
			continue
		}

		n, err := s.pipeRunnerStream(w, r, target, contentType, &wroteHeader, flusher)
		if r.Context().Err() != nil {
			return
		}
		// A stream that carried data then dropped is a runner restart: reconnect
		// promptly. A stream that never connected backs off so we don't spin.
		if n > 0 {
			backoff = 250 * time.Millisecond
		}
		golog.Info("screen stream reconnecting", "module", logModule, "udid", s.udid,
			"path", path, "bytes", n, "error", errString(err))
		if !sleepCtx(r.Context(), backoff) {
			return
		}
		if backoff *= 2; backoff > backoffMax {
			backoff = backoffMax
		}
	}
}

// pipeRunnerStream opens one connection to the runner and copies its body to w,
// flushing per read. It writes the response header (once, on the first
// successful connect) via wroteHeader. It returns the number of bytes copied and
// the error that ended the copy (nil on a clean EOF).
func (s *Server) pipeRunnerStream(w http.ResponseWriter, r *http.Request, target, contentType string, wroteHeader *bool, flusher http.Flusher) (int64, error) {
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, target, nil)
	if err != nil {
		return 0, err
	}
	// No client timeout: these are long-lived streams. The request context
	// (browser disconnect / server shutdown) is the only lifetime bound.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("runner %s returned %d", target, resp.StatusCode)
	}

	if !*wroteHeader {
		ct := contentType
		if ct == "" {
			ct = resp.Header.Get("Content-Type")
		}
		if ct != "" {
			w.Header().Set("Content-Type", ct)
		}
		w.Header().Set("Server", "go-ios-remote")
		w.Header().Set("Cache-Control", "no-cache, private")
		w.Header().Set("Pragma", "no-cache")
		w.WriteHeader(http.StatusOK)
		*wroteHeader = true
	}

	// Copy chunk-by-chunk, flushing each so bytes reach the browser immediately.
	buf := make([]byte, 32*1024)
	var total int64
	for {
		nr, rerr := resp.Body.Read(buf)
		if nr > 0 {
			if _, werr := w.Write(buf[:nr]); werr != nil {
				return total, werr
			}
			total += int64(nr)
			if flusher != nil {
				flusher.Flush()
			}
		}
		if rerr != nil {
			if rerr == io.EOF {
				return total, nil
			}
			return total, rerr
		}
	}
}

// sleepCtx sleeps for d unless ctx is cancelled first; it returns false if the
// context was cancelled (so the caller should stop).
func sleepCtx(ctx context.Context, d time.Duration) bool {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-t.C:
		return true
	}
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
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
	if !s.inputReady(w) {
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
	if !s.inputReady(w) {
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
	if !s.inputReady(w) {
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
	if !s.inputReady(w) {
		return
	}
	s.runUI(w, "button", req.Name)
}

// runUI shells out to `ios ui <args…>` for the configured device and driver,
// returning the combined output to the browser. A non-reachable driver surfaces
// here as a non-zero exit + stderr in the response.
func (s *Server) runUI(w http.ResponseWriter, args ...string) {
	full := append([]string{"ui"}, args...)
	// Pin the driver + its URL rather than let `ios ui` auto-detect.
	full = append(full, "--driver="+s.driver, "--udid="+s.udid, s.driverURLFlag())

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
// mapped to the driver's logical points. Both WDA and DeviceKit take taps in
// logical points and report their size in the same logical point space, so a
// fraction of the screen maps linearly: point = fraction * logicalSize. The
// logical size is fetched once (via `ios ui size` on the same driver) and
// cached — never a hardcoded/WDA size.
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

// logicalSize returns the device's logical window size in the driver's points,
// caching the first successful lookup. It calls `ios ui size` on the configured
// driver and parses that driver's response shape.
func (s *Server) logicalSize() (float64, float64, error) {
	s.sizeMu.Lock()
	defer s.sizeMu.Unlock()
	if s.sizeW > 0 && s.sizeH > 0 {
		return s.sizeW, s.sizeH, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, s.iosBinary, "ui", "size", "--driver="+s.driver, "--udid="+s.udid, s.driverURLFlag())
	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, 0, fmt.Errorf("ui size failed (is the %s driver reachable at %s?): %v: %s", s.driver, s.driverURL, err, out)
	}
	w, h, err := parseSize(out)
	if err != nil {
		return 0, 0, err
	}
	s.sizeW, s.sizeH = w, h
	return w, h, nil
}

// parseSize extracts logical width/height from the JSON emitted by `ios ui
// size`. It accepts:
//   - DeviceKit, which wraps a device.info result in a JSON-RPC envelope:
//     {"result":{"screenSize":{"width":W,"height":H},"scale":S}} — the logical
//     points are screenSize (the MJPEG frames are points × scale pixels).
//   - the same {"screenSize":{…}} object unwrapped.
//   - WDA: {"value":{"width":W,"height":H}} envelope.
//   - a bare {"width":W,"height":H} object.
//
// It tolerates surrounding whitespace/log noise by scanning for the first '{'.
func parseSize(out []byte) (float64, float64, error) {
	trimmed := trimToJSON(out)
	type screenSize struct {
		Width  float64 `json:"width"`
		Height float64 `json:"height"`
	}
	var envelope struct {
		Width      float64    `json:"width"`
		Height     float64    `json:"height"`
		ScreenSize screenSize `json:"screenSize"`
		// DeviceKit JSON-RPC envelope: {"result":{"screenSize":{…}}}.
		Result struct {
			ScreenSize screenSize `json:"screenSize"`
			Width      float64    `json:"width"`
			Height     float64    `json:"height"`
		} `json:"result"`
		// WDA envelope: {"value":{"width":…,"height":…}}.
		Value screenSize `json:"value"`
	}
	if err := json.Unmarshal(trimmed, &envelope); err != nil {
		return 0, 0, fmt.Errorf("parsing ui size output %q: %w", out, err)
	}
	w, h := envelope.Width, envelope.Height
	switch {
	case envelope.Result.ScreenSize.Width > 0:
		w, h = envelope.Result.ScreenSize.Width, envelope.Result.ScreenSize.Height
	case envelope.ScreenSize.Width > 0:
		w, h = envelope.ScreenSize.Width, envelope.ScreenSize.Height
	case envelope.Result.Width > 0:
		w, h = envelope.Result.Width, envelope.Result.Height
	case envelope.Value.Width > 0:
		w, h = envelope.Value.Width, envelope.Value.Height
	}
	if w <= 0 || h <= 0 {
		return 0, 0, fmt.Errorf("ui size returned non-positive dimensions from %q", out)
	}
	return w, h, nil
}

// trimToJSON strips leading log noise before the first '{'.
func trimToJSON(out []byte) []byte {
	for i := 0; i < len(out); i++ {
		if out[i] == '{' {
			return out[i:]
		}
	}
	return out
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

// Close releases server resources. The screen is now a stateless proxy of the
// DeviceKit runner, so there is nothing to tear down here; the supervised runner
// is stopped by Run on ctx cancellation.
func (s *Server) Close() {}
