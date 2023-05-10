package api_test

import (
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
)

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

var values = map[string]bool{}

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
	_, ok := values[key]
	if ok {
		t.Fail()
		return
	}
	values[key] = true
}
