package lock

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

var lockID string

func setupRouter() *gin.Engine {
	lockID = randomLockID()

	r := gin.Default()
	r.Use(fakeDeviceMiddleware())
	r.POST("/lock/:udid", LockDevice)
	r.DELETE("/lock/:udid", DeleteDeviceLock)
	r.GET("/locks", GetLockedDevices)

	locksMap = make(map[string]*lockedDevice)
	return r
}

func fakeDeviceMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set("go_ios_device", ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: lockID}})
	}
}

func TestDeviceLock(t *testing.T) {
	r := setupRouter()
	responseRecorder := httptest.NewRecorder()

	lockRequest, err := http.NewRequest("POST", "/lock/"+lockID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, lockRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "POST to %v was unsuccessful", lockRequest.URL)

	validateSuccessfulLock(t, responseRecorder)
}

func TestDeviceLockAlreadyLocked(t *testing.T) {
	r := setupRouter()
	responseRecorder := httptest.NewRecorder()

	lockRequest, err := http.NewRequest("POST", "/lock/"+lockID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, lockRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "Initial POST to %v was unsuccessful", lockRequest.URL)

	validateSuccessfulLock(t, responseRecorder)

	lockRequest, err = http.NewRequest("POST", "/lock/"+lockID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, lockRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "Second POST to %v was unsuccessful", lockRequest.URL)

	validateDeviceLocked(t, responseRecorder)
}

func validateSuccessfulLock(t *testing.T, responseRecorder *httptest.ResponseRecorder) {
	var lockIDResponse lockResponse
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &lockIDResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.NotEmpty(t, lockIDResponse.LockID, "Device was not successfully locked")
}

func validateDeviceLocked(t *testing.T, responseRecorder *httptest.ResponseRecorder) {
	var genericResponse genericLockResponse
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &genericResponse)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.Equal(t, "Already locked", genericResponse.Message)
}
