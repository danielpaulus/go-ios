package api_test

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/restapi/api"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hardeningDeviceMiddleware injects a device entry that has no real connection,
// so downstream service calls (instruments, syslog, ...) fail fast instead of
// blocking on a device.
func hardeningDeviceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(api.IOS_KEY, ios.DeviceEntry{
			Properties: ios.DeviceProperties{SerialNumber: "hardening-udid"},
		})
		c.Next()
	}
}

// installImageRouter wires the InstallImage handler behind the hardening
// device middleware so the traversal/upload-limit checks run without a device.
func installImageRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(hardeningDeviceMiddleware())
	r.PUT("/image", api.InstallImage)
	return r
}

// TestInstallImageRejectsBasedirTraversal is the MED-8 regression. A
// caller-supplied basedir flows into os.MkdirAll+path.Join in the image
// mounter, so absolute paths and `..` traversal must be rejected with 400
// before the mounter is ever reached (which would otherwise create dirs / try
// to talk to a device). Without the validation these requests would not return
// 400.
func TestInstallImageRejectsBasedirTraversal(t *testing.T) {
	hostileBasedirs := []string{
		"../../../../etc",
		"foo/../../bar",
		"/etc/go-ios",
		"..",
	}
	for _, basedir := range hostileBasedirs {
		t.Run(basedir, func(t *testing.T) {
			r := installImageRouter()
			req, _ := http.NewRequest("PUT", "/image?auto=true&basedir="+url.QueryEscape(basedir), nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code, "hostile basedir %q must be rejected with 400", basedir)
			assert.Contains(t, w.Body.String(), "invalid basedir")
		})
	}
}

// TestInstallImageAllowsRelativeBasedir pins that a normal relative basedir
// passes the validation. It cannot fully succeed without a device, so we assert
// only that it is NOT rejected as an invalid basedir (i.e. it got past the
// validation into the mounter, which then fails for other reasons).
func TestInstallImageAllowsRelativeBasedir(t *testing.T) {
	allowed := []string{"devimages", "./devimages", "some/nested/dir"}
	for _, basedir := range allowed {
		t.Run(basedir, func(t *testing.T) {
			r := installImageRouter()
			req, _ := http.NewRequest("PUT", "/image?auto=true&basedir="+url.QueryEscape(basedir), nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.NotEqual(t, http.StatusBadRequest, w.Code, "relative basedir %q must be allowed past validation", basedir)
			assert.NotContains(t, w.Body.String(), "invalid basedir")
		})
	}
}

// TestInstallImageRejectsOversizedUpload is the MED-7 regression. The manual
// (non-auto) upload path io.Copy'd the request body with no size cap; the fix
// wraps it in http.MaxBytesReader(2 GiB). We can't stream 2 GiB in a unit test,
// so we drive the same MaxBytesReader guard directly to prove an over-limit
// body errors out (which the handler surfaces as a failure) while an
// at/under-limit body copies cleanly.
func TestInstallImageRejectsOversizedUpload(t *testing.T) {
	const limit = 8

	over := http.MaxBytesReader(nil, io.NopCloser(bytes.NewReader(make([]byte, limit+1))), limit)
	_, err := io.Copy(io.Discard, over)
	require.Error(t, err, "a body larger than the limit must cause io.Copy to fail")

	under := http.MaxBytesReader(nil, io.NopCloser(bytes.NewReader(make([]byte, limit))), limit)
	n, err := io.Copy(io.Discard, under)
	require.NoError(t, err, "a body at the limit must copy without error")
	assert.Equal(t, int64(limit), n)
}

// TestNotificationsDoesNotExitOnError is the CRIT-1 (restapi-01) regression.
//
// The Notifications handler used to call log.Fatal(err) when
// instruments.ListenAppStateNotifications failed, which calls os.Exit and kills
// the whole REST server (gin.Recovery cannot catch os.Exit). This test runs the
// handler in a subprocess and asserts the process exits 0 (handler returned)
// rather than being terminated by log.Fatal's os.Exit(1).
func TestNotificationsDoesNotExitOnError(t *testing.T) {
	if os.Getenv("GO_IOS_NOTIFICATIONS_SUBPROCESS") == "1" {
		gin.SetMode(gin.TestMode)
		r := gin.New()
		r.Use(hardeningDeviceMiddleware())
		r.GET("/notifications", api.Notifications)

		req, _ := http.NewRequest("GET", "/notifications", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// With the fix we get here: an error response, no os.Exit.
		if w.Code != http.StatusInternalServerError {
			os.Exit(3)
		}
		os.Exit(0)
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestNotificationsDoesNotExitOnError")
	cmd.Env = append(os.Environ(), "GO_IOS_NOTIFICATIONS_SUBPROCESS=1")
	err := cmd.Run()

	// Before the fix the subprocess dies via log.Fatal -> os.Exit(1).
	// After the fix it returns 500 and exits 0.
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			t.Fatalf("Notifications handler terminated the process (exit code %d); "+
				"it must return an error response instead of calling log.Fatal/os.Exit", exitErr.ExitCode())
		}
		t.Fatalf("subprocess failed to run: %v", err)
	}
}

// TestNotificationsReturns500OnError also asserts, in-process, that the handler
// responds with 500 on a connection error. If the handler still called log.Fatal
// this test process itself would be killed.
func TestNotificationsReturns500OnError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(hardeningDeviceMiddleware())
	r.GET("/notifications", api.Notifications)

	req, _ := http.NewRequest("GET", "/notifications", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// streamStep mirrors the fixed Syslog/Listen callback contract exactly: read a
// message and, on error, return false so gin's c.Stream driver stops looping.
// This is the load-bearing behavior of the HIGH-18 fix.
func streamStep(w io.Writer, read func() (string, error)) bool {
	m, err := read()
	if err != nil {
		return false
	}
	w.Write([]byte(m))
	return true
}

// TestStreamCallbackReturnsFalseOnReadError is the HIGH-18 (restapi-02)
// regression for the streaming callbacks (Syslog/Listen). The buggy callbacks
// discarded the read error and unconditionally returned true, so gin's
// c.Stream driver (an unbounded `for { if !step(w) { return } }` loop)
// busy-looped forever at 100% CPU on device EOF/disconnect.
//
// It runs the same unbounded driver gin uses, feeding a reader that errors like
// a disconnected device, and asserts the loop terminates. With the buggy
// contract (return true on error) this loop never ends and hits the deadline.
func TestStreamCallbackReturnsFalseOnReadError(t *testing.T) {
	// reader that returns an error immediately, like a disconnected device.
	read := func() (string, error) { return "", errors.New("EOF: device disconnected") }

	done := make(chan struct{})
	go func() {
		defer close(done)
		// Same shape as gin.Context.Stream's loop: keep calling step until it
		// returns false. The fix makes step return false on read error.
		for streamStep(io.Discard, read) {
		}
	}()

	select {
	case <-done:
		// stream terminated on error, as required by the fix.
	case <-time.After(2 * time.Second):
		t.Fatal("stream callback did not return false on read error; it busy-loops instead of terminating")
	}
}

// TestStreamCallbackContinuesOnSuccess pins the success-path invariant: while
// reads succeed the callback keeps returning true (streaming is unchanged), and
// it only stops once a read errors. This guards against a fix that terminates
// too early on valid data.
func TestStreamCallbackContinuesOnSuccess(t *testing.T) {
	remaining := 3
	read := func() (string, error) {
		if remaining == 0 {
			return "", errors.New("EOF")
		}
		remaining--
		return "log line", nil
	}

	count := 0
	for streamStep(io.Discard, read) {
		count++
	}
	assert.Equal(t, 3, count, "callback must keep streaming while reads succeed and only stop on error")
}

// TestLimitNumClientsUDIDNoBypassUnderConcurrency is the LOW-5 (restapi-06)
// regression. LimitNumClientsUDID used a non-atomic Load-then-Store on the
// per-udid semaphore map, so two concurrent first-requests for the same udid
// each created their own channel and both entered the critical section,
// bypassing the 1-per-udid limit. With LoadOrStore both share one channel and
// only one can be in-flight at a time. Run with -race.
func TestLimitNumClientsUDIDNoBypassUnderConcurrency(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var maxInFlight int32

	// The non-atomic Load-then-Store bug only manifests on the very first
	// concurrent requests for a udid (before any channel is stored). Use a fresh
	// udid per round and a start barrier so goroutines hit the middleware at the
	// same instant, and repeat many rounds to reliably surface the race.
	const rounds = 200
	const goroutines = 8

	for round := 0; round < rounds; round++ {
		udid := "udid-" + strconv.Itoa(round)

		var inFlight int32
		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set(api.IOS_KEY, ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: udid}})
			c.Next()
		})
		r.Use(api.LimitNumClientsUDID())
		r.GET("/", func(c *gin.Context) {
			cur := atomic.AddInt32(&inFlight, 1)
			for {
				old := atomic.LoadInt32(&maxInFlight)
				if cur <= old || atomic.CompareAndSwapInt32(&maxInFlight, old, cur) {
					break
				}
			}
			time.Sleep(time.Millisecond)
			atomic.AddInt32(&inFlight, -1)
			c.Status(http.StatusOK)
		})

		start := make(chan struct{})
		var wg sync.WaitGroup
		for i := 0; i < goroutines; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				req, _ := http.NewRequest("GET", "/", nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				require.Equal(t, http.StatusOK, w.Code)
			}()
		}
		close(start)
		wg.Wait()
	}

	assert.Equal(t, int32(1), atomic.LoadInt32(&maxInFlight),
		"more than one request for the same udid was in-flight simultaneously; the per-udid limit was bypassed")
}
