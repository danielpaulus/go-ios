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

func TestDeviceLock(t *testing.T) {
	r := setupRouter()
	responseRecorder := httptest.NewRecorder()

	lockRequest, err := http.NewRequest("POST", "/lock/"+randomDeviceUDID, nil)
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

	lockRequest, err := http.NewRequest("POST", "/lock/"+randomDeviceUDID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, lockRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "Initial POST to %v was unsuccessful", lockRequest.URL)

	validateSuccessfulLock(t, responseRecorder)

	lockRequest, err = http.NewRequest("POST", "/lock/"+randomDeviceUDID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, lockRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "Second POST to %v was unsuccessful", lockRequest.URL)

	validateDeviceAlreadyLocked(t, responseRecorder)
}

func TestLockedDeviceInGetLockedDevicesList(t *testing.T) {
	r := setupRouter()
	responseRecorder := httptest.NewRecorder()

	lockRequest, err := http.NewRequest("POST", "/lock/"+randomDeviceUDID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, lockRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "POST to %v was unsuccessful", lockRequest.URL)

	lockID := validateSuccessfulLock(t, responseRecorder)

	getDevicesRequest, err := http.NewRequest("GET", "/locks", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, getDevicesRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "GET /locks was unsuccessful")

	var devicesResponse []lockedDevice
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err = json.Unmarshal(responseData, &devicesResponse)
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

	lockRequest, err := http.NewRequest("POST", "/lock/"+randomDeviceUDID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, lockRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "POST to %v was unsuccessful", lockRequest.URL)

	lockID := validateSuccessfulLock(t, responseRecorder)

	getDevicesRequest, err := http.NewRequest("GET", "/locks", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, getDevicesRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "GET /locks was unsuccessful")

	var devicesResponse []lockedDevice
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err = json.Unmarshal(responseData, &devicesResponse)
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

	deleteLockRequest, err := http.NewRequest("DELETE", "/lock/"+randomDeviceUDID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, deleteLockRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "DELETE to %v was unsuccessful", deleteLockRequest.URL)

	validateDeviceUnlocked(t, responseRecorder)

	r.ServeHTTP(responseRecorder, getDevicesRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "GET /locks was unsuccessful")

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
	r := setupRouter()
	responseRecorder := httptest.NewRecorder()

	deleteLockRequest, err := http.NewRequest("DELETE", "/lock/"+randomDeviceUDID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, deleteLockRequest)
	require.Equal(t, http.StatusNotFound, responseRecorder.Code, "DELETE to %v was unsuccessful", deleteLockRequest.URL)

	validateNotLocked(t, responseRecorder)
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
