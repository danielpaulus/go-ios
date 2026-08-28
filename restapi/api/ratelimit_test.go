package api

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/gin-gonic/gin"
)

// rateLimitRouter builds a minimal router: DeviceMiddleware-substitute that sets
// a fixed udid, the rate limiter, and a 200 handler.
func rateLimitRouter(udid string, perSecond float64, burst int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(IOS_KEY, ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: udid}})
	})
	r.Use(RateLimitUDID(perSecond, burst))
	r.GET("/x", func(c *gin.Context) { c.String(http.StatusOK, "ok") })
	return r
}

func TestRateLimitAllowsBurstThenRejects(t *testing.T) {
	// A tiny sustained rate so refills don't interfere within the test window.
	r := rateLimitRouter("UDID-A", 1, 5)
	var ok, limited int
	for i := 0; i < 20; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
		switch w.Code {
		case http.StatusOK:
			ok++
		case http.StatusTooManyRequests:
			limited++
		default:
			t.Fatalf("unexpected status %d", w.Code)
		}
	}
	if ok != 5 {
		t.Fatalf("expected burst of 5 to pass, got %d (limited=%d)", ok, limited)
	}
	if limited != 15 {
		t.Fatalf("expected 15 rejections, got %d", limited)
	}
}

func TestRateLimitDisabledWhenZero(t *testing.T) {
	r := rateLimitRouter("UDID-A", 0, 0)
	for i := 0; i < 100; i++ {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("rate limiting should be disabled, got %d on request %d", w.Code, i)
		}
	}
}

// TestRateLimitConcurrentPerDevice hammers one device from many goroutines and
// asserts exactly `burst` requests succeed (the bucket is shared and consumed
// atomically), with no races. Run with -race.
func TestRateLimitConcurrentPerDevice(t *testing.T) {
	const burst = 10
	r := rateLimitRouter("UDID-A", 1, burst)
	var okCount int64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			w := httptest.NewRecorder()
			r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
			if w.Code == http.StatusOK {
				atomic.AddInt64(&okCount, 1)
			}
		}()
	}
	close(start)
	wg.Wait()
	// The token bucket may refill by a token or two during the burst, so allow a
	// tiny margin; the point is it doesn't blow past the bucket under concurrency.
	if okCount < burst || okCount > burst+2 {
		t.Fatalf("concurrent successes = %d, want ~%d (bucket must not be exceeded)", okCount, burst)
	}
}

// TestRateLimitIsolatesDevices confirms one device's exhausted bucket doesn't
// reject another device's requests.
func TestRateLimitIsolatesDevices(t *testing.T) {
	limiter := RateLimitUDID(1, 2)
	run := func(udid string) int {
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
		c.Set(IOS_KEY, ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: udid}})
		limiter(c)
		return w.Code
	}
	// Drain device A's bucket (burst 2).
	run("A")
	run("A")
	if code := run("A"); code != http.StatusTooManyRequests {
		t.Fatalf("device A should be limited, got %d", code)
	}
	// Device B is unaffected.
	if code := run("B"); code == http.StatusTooManyRequests {
		t.Fatalf("device B must not be limited by device A's traffic")
	}
}
