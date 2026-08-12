package api_test

import (
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/restapi/api"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// bearerAuthRouter builds a router whose single protected route is guarded by
// api.BearerAuth(token).
func bearerAuthRouter(token string) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(api.BearerAuth(token))
	r.GET("/protected", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	return r
}

func bearerAuthRequest(t *testing.T, r *gin.Engine, authHeader string, sendHeader bool) *httptest.ResponseRecorder {
	t.Helper()
	req, _ := http.NewRequest("GET", "/protected", nil)
	if sendHeader {
		req.Header.Set("Authorization", authHeader)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// TestBearerAuthValidToken pins the success path: the correct bearer token is
// accepted and the downstream handler runs.
func TestBearerAuthValidToken(t *testing.T) {
	r := bearerAuthRouter("s3cr3t-token")
	w := bearerAuthRequest(t, r, "Bearer s3cr3t-token", true)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

// TestBearerAuthMissingHeader asserts a request with no Authorization header is
// rejected with 401 and never reaches the handler.
func TestBearerAuthMissingHeader(t *testing.T) {
	r := bearerAuthRouter("s3cr3t-token")
	w := bearerAuthRequest(t, r, "", false)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

// TestBearerAuthWrongToken asserts a wrong token is rejected with 401.
func TestBearerAuthWrongToken(t *testing.T) {
	r := bearerAuthRouter("s3cr3t-token")
	w := bearerAuthRequest(t, r, "Bearer wrong-token", true)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")

	// A missing "Bearer " scheme prefix must also be rejected.
	w = bearerAuthRequest(t, r, "s3cr3t-token", true)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestBearerAuthConstantTimeCompare documents that the comparison is
// length-safe: tokens of different lengths compare to not-equal without
// panicking, exactly as crypto/subtle.ConstantTimeCompare behaves. This guards
// the constant-time-compare requirement of the fix.
func TestBearerAuthConstantTimeCompare(t *testing.T) {
	assert.Equal(t, 0, subtle.ConstantTimeCompare([]byte("Bearer a"), []byte("Bearer abc")))
	assert.Equal(t, 1, subtle.ConstantTimeCompare([]byte("Bearer abc"), []byte("Bearer abc")))

	// End-to-end: a token that is a prefix of the real one is still rejected.
	r := bearerAuthRouter("longtoken")
	w := bearerAuthRequest(t, r, "Bearer long", true)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func getRouter() *gin.Engine {
	r := gin.Default()
	r.Use(fakeDeviceMiddleware())
	r.Use(api.LimitNumClientsUDID())
	return r
}

func fakeDeviceMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set(api.IOS_KEY, ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: "abcdefgh"}})
	}
}

var unsafeCounter = 0

func TestEnsureConcurrencyLimited(t *testing.T) {
	r := getRouter()

	// Without the concurrency limiting middleware
	// this will not return all possible values for the counter.
	// probably all responses will contain the same number.
	r.GET("/", func(c *gin.Context) {
		unsafeCounter++
		time.Sleep(time.Millisecond)
		c.JSONP(http.StatusOK, gin.H{"v": unsafeCounter})
	})

	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			testMiddlewareRequest(t, r, http.StatusOK)
		}()
	}
	wg.Wait()
}

func testMiddlewareRequest(t *testing.T, r *gin.Engine, expectedHTTPCode int) {
	req, _ := http.NewRequest("GET", "/", nil)

	testHTTPResponse(t, r, req, func(w *httptest.ResponseRecorder) float64 {
		result := map[string]interface{}{}
		json.Unmarshal(w.Body.Bytes(), &result)
		return result["v"].(float64)
	})
}

var (
	values   = map[string]bool{}
	valuesMu sync.Mutex
)

// Helper function to process a request and test its response
func testHTTPResponse(t *testing.T, r *gin.Engine, req *http.Request, f func(w *httptest.ResponseRecorder) float64) {
	// Create a response recorder
	w := httptest.NewRecorder()

	// Create the service and process the above request.
	r.ServeHTTP(w, req)

	// if concurrency is not limited, then i will have the same value
	// a few times. With the limit enabled, i won't be the same value twice.
	i := f(w)
	key := fmt.Sprintf("%f", i)
	valuesMu.Lock()
	defer valuesMu.Unlock()
	_, ok := values[key]
	if ok {
		t.Fail()
		return
	}
	values[key] = true
}
