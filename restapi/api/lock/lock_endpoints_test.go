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

var randomDeviceUDID string
var r *gin.Engine

func setupRouter() *gin.Engine {
	randomDeviceUDID = randomLockID()

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
		context.Set("go_ios_device", ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: randomDeviceUDID}})
	}
}

// TESTS
func TestDeviceLock(t *testing.T) {
	r = setupRouter()
	responseRecorder := httptest.NewRecorder()

	// Lock the device
	lockRequest := postLock(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "POST to %v was unsuccessful", lockRequest.URL)
	validateSuccessfulLock(t, responseRecorder)
}

func TestDeviceLockAlreadyLocked(t *testing.T) {
	r = setupRouter()
	responseRecorder := httptest.NewRecorder()

	// Lock the device
	lockRequest := postLock(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "Initial POST to %v was unsuccessful", lockRequest.URL)
	validateSuccessfulLock(t, responseRecorder)

	// Try to lock the already locked device
	lockRequest = postLock(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "Second POST to %v was unsuccessful", lockRequest.URL)
	validateDeviceAlreadyLocked(t, responseRecorder)
}

func TestLockedDeviceInGetLockedDevicesList(t *testing.T) {
	r = setupRouter()
	responseRecorder := httptest.NewRecorder()

	// Lock the device
	lockRequest := postLock(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "POST to %v was unsuccessful", lockRequest.URL)
	lockID := validateSuccessfulLock(t, responseRecorder)

	// Validate device is in /locks list
	getDevicesRequest := getDeviceLocks(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "GET %v was unsuccessful", getDevicesRequest.URL)

	var devicesResponse []lockedDevice
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &devicesResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	var lockid_exists = false
	for _, device := range devicesResponse {
		if device.LockID == lockID {
			lockid_exists = true
			require.Equal(t, device.UDID, randomDeviceUDID, "Device UDID does not correspond to the LockID, expected UDID=%v, got=%v", randomDeviceUDID, device.UDID)
		}
	}

	require.True(t, lockid_exists, "Could not find `lock_id`=%v in GET /locks response", lockID)
}

func TestDeletingDeviceLock(t *testing.T) {
	r := setupRouter()
	responseRecorder := httptest.NewRecorder()

	// Lock the device
	lockRequest := postLock(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "POST to %v was unsuccessful", lockRequest.URL)
	lockID := validateSuccessfulLock(t, responseRecorder)

	// Validate device is in /locks list
	getDevicesRequest := getDeviceLocks(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "GET %v was unsuccessful", getDevicesRequest.URL)

	var devicesResponse []lockedDevice
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &devicesResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	var lockid_exists = false
	for _, device := range devicesResponse {
		if device.LockID == lockID {
			lockid_exists = true
			require.Equal(t, device.UDID, randomDeviceUDID, "Device UDID does not correspond to the LockID, expected UDID=%v, got=%v", randomDeviceUDID, device.UDID)
		}
	}
	require.True(t, lockid_exists, "Could not find `lock_id`=%v in GET /locks response", lockID)

	// Delete device lock
	deleteLockRequest := deleteLock(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "DELETE %v was unsuccessful", deleteLockRequest.URL)
	validateDeviceUnlocked(t, responseRecorder)

	// Validate no devices present in /locks response
	r.ServeHTTP(responseRecorder, getDevicesRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "GET %v was unsuccessful", getDevicesRequest.URL)

	var noLockedDevicesResponse genericLockResponse
	responseData, _ = ioutil.ReadAll(responseRecorder.Body)
	err = json.Unmarshal(responseData, &noLockedDevicesResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.Equal(t, noLockedDevicesResponse.Message, "No locked devices found")
}

func TestValidateDeviceNotLocked(t *testing.T) {
	r = setupRouter()
	responseRecorder := httptest.NewRecorder()

	// Validate device not locked response
	deleteLockRequest := deleteLock(t, responseRecorder)
	require.Equal(t, http.StatusNotFound, responseRecorder.Code, "DELETE to %v was unsuccessful", deleteLockRequest.URL)
	validateNotLocked(t, responseRecorder)
}

// HELPER FUNCTIONS

func postLock(t *testing.T, responseRecorder *httptest.ResponseRecorder) *http.Request {
	lockRequest, err := http.NewRequest("POST", "/lock/"+randomDeviceUDID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, lockRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "POST to %v was unsuccessful", lockRequest.URL)

	return lockRequest
}

func deleteLock(t *testing.T, responseRecorder *httptest.ResponseRecorder) *http.Request {
	deleteLockRequest, err := http.NewRequest("DELETE", "/lock/"+randomDeviceUDID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, deleteLockRequest)

	return deleteLockRequest
}

func getDeviceLocks(t *testing.T, responseRecorder *httptest.ResponseRecorder) *http.Request {
	getDevicesRequest, err := http.NewRequest("GET", "/locks", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, getDevicesRequest)

	return getDevicesRequest
}

func validateSuccessfulLock(t *testing.T, responseRecorder *httptest.ResponseRecorder) string {
	var lockIDResponse lockResponse
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &lockIDResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.NotEmpty(t, lockIDResponse.LockID, "Device was not successfully locked")

	return lockIDResponse.LockID
}

func validateDeviceAlreadyLocked(t *testing.T, responseRecorder *httptest.ResponseRecorder) {
	var genericResponse genericLockResponse
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &genericResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.Equal(t, "Already locked", genericResponse.Message)
}

func validateNotLocked(t *testing.T, responseRecorder *httptest.ResponseRecorder) {
	var genericResponse genericLockResponse
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &genericResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.Equal(t, "Not locked", genericResponse.Message)
}

func validateDeviceUnlocked(t *testing.T, responseRecorder *httptest.ResponseRecorder) {
	var genericResponse genericLockResponse
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &genericResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.Equal(t, "Successfully unlocked", genericResponse.Message)
}
