package reservation

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
	randomDeviceUDID = randomReservationID()

	r := gin.Default()
	r.Use(fakeDeviceMiddleware())
	r.POST("/reserve/:udid", ReserveDevice)
	r.DELETE("/reserve/:udid", ReleaseDevice)
	r.GET("/reserved-devices", GetReservedDevices)

	reservedDevicesMap = make(map[string]*reservedDevice)
	return r
}

func fakeDeviceMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		context.Set("go_ios_device", ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: randomDeviceUDID}})
	}
}

// TESTS
func TestDeviceReservation(t *testing.T) {
	r = setupRouter()
	responseRecorder := httptest.NewRecorder()

	// Reserve the device
	reserveRequest := postReservation(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "POST to %v was unsuccessful", reserveRequest.URL)
	validateSuccessfulReservation(t, responseRecorder)
}

func TestDeviceReservationAlreadyReserved(t *testing.T) {
	r = setupRouter()
	responseRecorder := httptest.NewRecorder()

	// Reserve the device
	reserveRequest := postReservation(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "Initial POST to %v was unsuccessful", reserveRequest.URL)
	validateSuccessfulReservation(t, responseRecorder)

	// Try to reserve the already reserved device
	reserveRequest = postReservation(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "Second POST to %v was unsuccessful", reserveRequest.URL)
	validateDeviceAlreadyReserved(t, responseRecorder)
}

func TestReleasingDevice(t *testing.T) {
	r := setupRouter()
	responseRecorder := httptest.NewRecorder()

	// Reserve the device
	reserveRequest := postReservation(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "POST to %v was unsuccessful", reserveRequest.URL)
	reserveID := validateSuccessfulReservation(t, responseRecorder)

	// Validate device is in /reserved-devices list
	getDevicesRequest := getReservedDevices(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "GET %v was unsuccessful", getDevicesRequest.URL)

	var devicesResponse []reservedDevice
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &devicesResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	var lockid_exists = false
	for _, device := range devicesResponse {
		if device.ReservationID == reserveID {
			lockid_exists = true
			require.Equal(t, device.UDID, randomDeviceUDID, "Device UDID does not correspond to the ReservationID, expected UDID=%v, got=%v", randomDeviceUDID, device.UDID)
			require.NotEmpty(t, device.LastUsedTimestamp, "`lastUsed` is empty but it shouldn't be")
		}
	}
	require.True(t, lockid_exists, "Could not find device with `reservation_id`=%v in GET /reserved-devices response", reserveID)

	// Release the reserved device
	releaseDeviceRequest := deleteReservation(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "DELETE %v was unsuccessful", releaseDeviceRequest.URL)
	validateDeviceReleased(t, responseRecorder)

	// Validate no reserved devices present in /reserved-devices response
	r.ServeHTTP(responseRecorder, getDevicesRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "GET %v was unsuccessful", getDevicesRequest.URL)

	var noReservedDevicesResponse reservedDevice
	responseData, _ = ioutil.ReadAll(responseRecorder.Body)
	err = json.Unmarshal(responseData, &noReservedDevicesResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.Equal(t, noReservedDevicesResponse.Message, "No reserved devices found")
}

func TestValidateDeviceNotReserved(t *testing.T) {
	r = setupRouter()
	responseRecorder := httptest.NewRecorder()

	// Validate device not reserved response
	releaseDeviceRequest := deleteReservation(t, responseRecorder)
	require.Equal(t, http.StatusNotFound, responseRecorder.Code, "DELETE %v was unsuccessful", releaseDeviceRequest.URL)
	validateNotReserved(t, responseRecorder)
}

// HELPER FUNCTIONS
func postReservation(t *testing.T, responseRecorder *httptest.ResponseRecorder) *http.Request {
	reserveDevice, err := http.NewRequest("POST", "/reserve/"+randomDeviceUDID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, reserveDevice)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "POST to %v was unsuccessful", reserveDevice.URL)

	return reserveDevice
}

func deleteReservation(t *testing.T, responseRecorder *httptest.ResponseRecorder) *http.Request {
	releaseDeviceRequest, err := http.NewRequest("DELETE", "/reserve/"+randomDeviceUDID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, releaseDeviceRequest)

	return releaseDeviceRequest
}

func getReservedDevices(t *testing.T, responseRecorder *httptest.ResponseRecorder) *http.Request {
	getDevicesRequest, err := http.NewRequest("GET", "/reserved-devices", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, getDevicesRequest)

	return getDevicesRequest
}

func validateSuccessfulReservation(t *testing.T, responseRecorder *httptest.ResponseRecorder) string {
	var reservationIDResponse reservedDevice
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &reservationIDResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.NotEmpty(t, reservationIDResponse.ReservationID, "Device was not successfully reserved")

	return reservationIDResponse.ReservationID
}

func validateDeviceAlreadyReserved(t *testing.T, responseRecorder *httptest.ResponseRecorder) {
	var alreadyReservedResponse reservedDevice
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &alreadyReservedResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.Equal(t, "Already reserved", alreadyReservedResponse.Message)
}

func validateNotReserved(t *testing.T, responseRecorder *httptest.ResponseRecorder) {
	var notReservedResponse reservedDevice
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &notReservedResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.Equal(t, "Not reserved", notReservedResponse.Message)
}

func validateDeviceReleased(t *testing.T, responseRecorder *httptest.ResponseRecorder) {
	var deviceReleasedResponse reservedDevice
	responseData, _ := ioutil.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &deviceReleasedResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.Equal(t, "Successfully released", deviceReleasedResponse.Message)
}
