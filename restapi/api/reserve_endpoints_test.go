package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var (
	randomDeviceUDID string
	r                *gin.Engine
)

func setupRouter() *gin.Engine {
	randomDeviceUDID = uuid.New().String()

	r := gin.Default()
	r.Use(fakeDeviceMiddleware())
	r.POST("/:udid/reservations", ReserveDevice)
	r.DELETE("/reservations/:reservationID", ReleaseDevice)
	r.GET("/reservations", GetReservedDevices)

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
	postReservation(t, responseRecorder)
	validateSuccessfulReservation(t, responseRecorder)
}

func TestDeviceReservationAlreadyReserved(t *testing.T) {
	r = setupRouter()
	responseRecorder := httptest.NewRecorder()

	// Reserve the device
	postReservation(t, responseRecorder)
	validateSuccessfulReservation(t, responseRecorder)

	// Try to reserve the already reserved device
	responseRecorder = httptest.NewRecorder()
	postReservation(t, responseRecorder)
	validateDeviceAlreadyReserved(t, responseRecorder)
}

func TestReleasingDevice(t *testing.T) {
	r := setupRouter()
	responseRecorder := httptest.NewRecorder()

	// Reserve the device
	postReservation(t, responseRecorder)
	reservationID := validateSuccessfulReservation(t, responseRecorder)

	// Validate device is in /reservations list
	responseRecorder = httptest.NewRecorder()
	getDevicesRequest := getReservedDevices(t, responseRecorder)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "GET %v was unsuccessful", getDevicesRequest.URL)

	var devicesResponse []reservedDevice
	responseData, _ := io.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &devicesResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	reservationid_exists := false
	for _, device := range devicesResponse {
		if device.ReservationID == reservationID {
			reservationid_exists = true
			require.Equal(t, device.UDID, randomDeviceUDID, "Device UDID does not correspond to the ReservationID, expected UDID=%v, got=%v", randomDeviceUDID, device.UDID)
			require.NotEmpty(t, device.LastUsedTimestamp, "`lastUsed` is empty but it shouldn't be")
		}
	}
	require.True(t, reservationid_exists, "Could not find device with `reservationID`=%v in GET /reservations response", reservationID)

	// Release the reserved device
	responseRecorder = httptest.NewRecorder()
	releaseDeviceRequest := deleteReservation(t, responseRecorder, reservationID)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "DELETE %v was unsuccessful", releaseDeviceRequest.URL)
	validateDeviceReleased(t, responseRecorder)

	// Validate no reserved devices present in /reservations response
	r.ServeHTTP(responseRecorder, getDevicesRequest)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "GET %v was unsuccessful", getDevicesRequest.URL)

	var noReservedDevicesResponse []reservedDevice
	responseData, _ = io.ReadAll(responseRecorder.Body)
	err = json.Unmarshal(responseData, &noReservedDevicesResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.Equal(t, noReservedDevicesResponse, []reservedDevice{})
}

func TestValidateDeviceNotReserved(t *testing.T) {
	r = setupRouter()
	responseRecorder := httptest.NewRecorder()

	// Validate device not reserved response
	releaseDeviceRequest := deleteReservation(t, responseRecorder, "test")
	require.Equal(t, http.StatusNotFound, responseRecorder.Code, "DELETE %v was unsuccessful", releaseDeviceRequest.URL)
	validateNotReserved(t, responseRecorder)
}

func TestValidateMiddlewareHeaderMissing(t *testing.T) {
	r = setupRouter()
	r.Use(ReserveDevicesMiddleware())
	r.POST("/:udid/launch", LaunchApp)

	responseRecorder := httptest.NewRecorder()

	launchAppRequest, err := http.NewRequest("POST", "/"+randomDeviceUDID+"/launch?bundleID=com.apple.Preferences", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, launchAppRequest)
	require.Equal(t, http.StatusBadRequest, responseRecorder.Code, "Code should be BadRequest if X-GO-IOS-RESERVE header is missing")

	var response GenericResponse
	responseData, _ := io.ReadAll(responseRecorder.Body)
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.NotEmpty(t, response.Error, "There is no error message returned when X-GO-IOS-RESERVE header is missing")
}

func TestValidateMiddlewareHeaderEmpty(t *testing.T) {
	r = setupRouter()
	r.Use(ReserveDevicesMiddleware())
	r.POST("/:udid/launch", LaunchApp)

	responseRecorder := httptest.NewRecorder()

	launchAppRequest, err := http.NewRequest("POST", "/"+randomDeviceUDID+"/launch?bundleID=com.apple.Preferences", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	launchAppRequest.Header.Add("X-GO-IOS-RESERVE", "")
	r.ServeHTTP(responseRecorder, launchAppRequest)
	require.Equal(t, http.StatusBadRequest, responseRecorder.Code, "Code should be BadRequest if X-GO-IOS-RESERVE header is empty")

	var response GenericResponse
	responseData, _ := io.ReadAll(responseRecorder.Body)
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.NotEmpty(t, response.Error, "There is no error message returned when X-GO-IOS-RESERVE header is missing")
}

func TestValidateMiddlewareHeaderDeviceNotReserved(t *testing.T) {
	r = setupRouter()
	r.Use(ReserveDevicesMiddleware())
	r.POST("/:udid/launch", LaunchApp)

	responseRecorder := httptest.NewRecorder()

	launchAppRequest, err := http.NewRequest("POST", "/"+randomDeviceUDID+"/launch?bundleID=com.apple.Preferences", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	launchAppRequest.Header.Add("X-GO-IOS-RESERVE", "go-admin")
	r.ServeHTTP(responseRecorder, launchAppRequest)
	require.Equal(t, http.StatusBadRequest, responseRecorder.Code, "Code should be BadRequest if X-GO-IOS-RESERVE header is empty")

	var response GenericResponse
	responseData, _ := io.ReadAll(responseRecorder.Body)
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.NotEmpty(t, response.Error, "There is no error message returned when X-GO-IOS-RESERVE header is missing")
}

func TestValidateMiddlewareDeviceReservedWrongUUID(t *testing.T) {
	r = setupRouter()
	r.Use(ReserveDevicesMiddleware())
	r.POST("/:udid/launch", LaunchApp)

	responseRecorder := httptest.NewRecorder()

	postReservation(t, responseRecorder)

	responseRecorder = httptest.NewRecorder()
	launchAppRequest, err := http.NewRequest("POST", "/"+randomDeviceUDID+"/launch?bundleID=com.apple.Preferences", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	launchAppRequest.Header.Set("X-GO-IOS-RESERVE", "bad-uuid")
	r.ServeHTTP(responseRecorder, launchAppRequest)
	require.Equal(t, http.StatusBadRequest, responseRecorder.Code, "Code should be BadRequest if X-GO-IOS-RESERVE header is empty")

	var response GenericResponse
	responseData, _ := io.ReadAll(responseRecorder.Body)
	err = json.Unmarshal(responseData, &response)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.NotEmpty(t, response.Error, "There is no error message returned when X-GO-IOS-RESERVE header is missing")
}

func TestValidateMiddlewareDeviceReservedValidUUID(t *testing.T) {
	r = setupRouter()
	r.Use(ReserveDevicesMiddleware())
	r.POST("/:udid/launch", LaunchApp)

	responseRecorder := httptest.NewRecorder()

	postReservation(t, responseRecorder)
	reservationID := validateSuccessfulReservation(t, responseRecorder)

	responseRecorder = httptest.NewRecorder()
	launchAppRequest, err := http.NewRequest("POST", "/"+randomDeviceUDID+"/launch?bundleID=com.apple.Preferences", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	launchAppRequest.Header.Set("X-GO-IOS-RESERVE", reservationID)
	r.ServeHTTP(responseRecorder, launchAppRequest)
	// Launching app does not really work with a mocked device
	// We check that status is not 400 because in the current scenario 400 is only returned when there is a problem with the reservation header
	require.NotEqual(t, http.StatusBadRequest, responseRecorder.Code)
}

func TestValidateMiddlewareDeviceReservedAdminUUID(t *testing.T) {
	r = setupRouter()
	r.Use(ReserveDevicesMiddleware())
	r.POST("/:udid/launch", LaunchApp)

	responseRecorder := httptest.NewRecorder()

	postReservation(t, responseRecorder)
	validateSuccessfulReservation(t, responseRecorder)

	responseRecorder = httptest.NewRecorder()
	launchAppRequest, err := http.NewRequest("POST", "/"+randomDeviceUDID+"/launch?bundleID=com.apple.Preferences", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	launchAppRequest.Header.Set("X-GO-IOS-RESERVE", reserveAdminUUID)
	r.ServeHTTP(responseRecorder, launchAppRequest)
	// Launching app does not really work with a mocked device
	// We check that status is not 400 because in the current scenario 400 is only returned when there is a problem with the reservation header
	require.NotEqual(t, http.StatusBadRequest, responseRecorder.Code)
}

// HELPER FUNCTIONS
func postReservation(t *testing.T, responseRecorder *httptest.ResponseRecorder) *http.Request {
	reserveDevice, err := http.NewRequest("POST", "/"+randomDeviceUDID+"/reservations", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, reserveDevice)
	require.Equal(t, http.StatusOK, responseRecorder.Code, "POST to %v was unsuccessful", reserveDevice.URL)

	return reserveDevice
}

func deleteReservation(t *testing.T, responseRecorder *httptest.ResponseRecorder, reservationID string) *http.Request {
	releaseDeviceRequest, err := http.NewRequest("DELETE", "/reservations/"+reservationID, nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, releaseDeviceRequest)

	return releaseDeviceRequest
}

func getReservedDevices(t *testing.T, responseRecorder *httptest.ResponseRecorder) *http.Request {
	getDevicesRequest, err := http.NewRequest("GET", "/reservations", nil)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	r.ServeHTTP(responseRecorder, getDevicesRequest)

	return getDevicesRequest
}

func validateSuccessfulReservation(t *testing.T, responseRecorder *httptest.ResponseRecorder) string {
	var reservationIDResponse reservedDevice
	responseData, _ := io.ReadAll(responseRecorder.Body)
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
	responseData, _ := io.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &alreadyReservedResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.Equal(t, "Already reserved", alreadyReservedResponse.Message)
}

func validateNotReserved(t *testing.T, responseRecorder *httptest.ResponseRecorder) {
	var notReservedResponse reservedDevice
	responseData, _ := io.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &notReservedResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.Equal(t, "Not reserved or wrong reservationID", notReservedResponse.Message)
}

func validateDeviceReleased(t *testing.T, responseRecorder *httptest.ResponseRecorder) {
	var deviceReleasedResponse reservedDevice
	responseData, _ := io.ReadAll(responseRecorder.Body)
	err := json.Unmarshal(responseData, &deviceReleasedResponse)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	require.Equal(t, "Successfully released", deviceReleasedResponse.Message)
}
