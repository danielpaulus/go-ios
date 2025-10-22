package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/restapi/api"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.Use(mockDeviceMiddleware())
	return r
}

func mockDeviceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a mock device entry
		c.Set(api.IOS_KEY, ios.DeviceEntry{
			Properties: ios.DeviceProperties{
				SerialNumber: "test-device-udid-12345",
			},
		})
		c.Next()
	}
}

// TestResetAccessibility tests the ResetAccessibility endpoint
// Note: This is a basic unit test that verifies the endpoint structure.
// It will fail if run without a real device because it tries to connect to the accessibility service.
// For true unit testing, the accessibility service would need to be mocked.
// to run this unit test, you can use `go test -tags restapi ./restapi/api/...` command
func TestResetAccessibilityEndpoint(t *testing.T) {
	// This test verifies that:
	// 1. The endpoint is properly registered
	// 2. The endpoint returns the expected response structure
	// 3. Error handling works correctly

	t.Run("endpoint returns proper error when no device available", func(t *testing.T) {
		router := setupTestRouter()
		router.POST("/resetaccessibility", api.ResetAccessibility)

		req, _ := http.NewRequest("POST", "/resetaccessibility", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// We expect an error since no real device is connected
		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)

		// Should have an error field
		_, hasError := response["error"]
		assert.True(t, hasError, "Response should contain an error field")
	})
}
