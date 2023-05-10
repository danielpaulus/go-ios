package api

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var (
	reservedDevicesMap     = make(map[string]*reservedDevice)
	reserveMutex           sync.Mutex
	reservedDevicesTimeout time.Duration = 5
	reserveAdminUUID                     = "go-admin"
)

type reservedDevice struct {
	Message           string `json:"message,omitempty"`
	UDID              string `json:"udid,omitempty"`
	ReservationID     string `json:"reservationID,omitempty"`
	LastUsedTimestamp int64  `json:"lastUsed,omitempty"`
}

// Reserve device access
// List          godoc
// @Summary      Reserve a device
// @Description  Reserve a device by provided UDID
// @Tags         reservations
// @Param        udid  path      string  true  "device udid"
// @Produce      json
// @Success      200  {object} reservedDevice
// @Router       /{udid}/reservations [post]
func ReserveDevice(c *gin.Context) {
	udid := c.Param("udid")
	reservationID := uuid.New().String()

	reserveMutex.Lock()
	defer reserveMutex.Unlock()

	// Check if there is a reserved device for the respective UDID
	_, exists := reservedDevicesMap[udid]
	if exists {
		c.IndentedJSON(http.StatusOK, reservedDevice{Message: "Already reserved"})
		return
	}

	newReservedDevice := reservedDevice{ReservationID: reservationID, LastUsedTimestamp: time.Now().UnixMilli()}
	reservedDevicesMap[udid] = &newReservedDevice

	c.IndentedJSON(http.StatusOK, reservedDevice{ReservationID: reservationID})
}

// Release device access
// List          godoc
// @Summary      Release a device
// @Description  Release a device by provided UDID
// @Tags         reservations
// @Param        reservationID  path      string  true  "reservation ID generated when reserving device"
// @Produce      json
// @Success      200  {object} reservedDevice
// @Failure      404  {object} reservedDevice
// @Router       /reservations/{reservationID} [delete]
func ReleaseDevice(c *gin.Context) {
	reservationID := c.Param("reservationID")

	reserveMutex.Lock()
	defer reserveMutex.Unlock()

	for udid, device := range reservedDevicesMap {
		if device.ReservationID == reservationID {
			delete(reservedDevicesMap, udid)
			c.IndentedJSON(http.StatusOK, reservedDevice{Message: "Successfully released"})
			return
		}
	}

	c.IndentedJSON(http.StatusNotFound, reservedDevice{Message: "Not reserved or wrong reservationID"})
}

// Get all reserved devices
// List          godoc
// @Summary      Get a list of reserved devices
// @Description  Get a list of reserved devices with UDID, ReservationID and last used timestamp
// @Tags         reservations
// @Produce      json
// @Success      200  {object} []reservedDevice
// @Router       /reservations [get]
func GetReservedDevices(c *gin.Context) {
	reserveMutex.Lock()
	defer reserveMutex.Unlock()

	reserved_devices := []reservedDevice{}

	if len(reservedDevicesMap) == 0 {
		c.IndentedJSON(http.StatusOK, reserved_devices)
		return
	}

	// Build the JSON array of currently reserved devices
	for udid, device := range reservedDevicesMap {
		reserved_devices = append(reserved_devices, reservedDevice{
			UDID:              udid,
			ReservationID:     device.ReservationID,
			LastUsedTimestamp: device.LastUsedTimestamp,
		})
	}

	c.IndentedJSON(http.StatusOK, reserved_devices)
}

func cleanReservationsCRON() {
	defer reserveMutex.Unlock()

	// Every minute loop through the map of reserved devices and check if a reserved device last used timestamp was more than X minutes ago
	// If any, remove them from the map
	for range time.Tick(time.Second * 60) {
		reserveMutex.Lock()
		for udid, reservedDevice := range reservedDevicesMap {
			currentTimestamp := time.Now().UnixMilli()
			diff := currentTimestamp - reservedDevice.LastUsedTimestamp

			if diff > (time.Minute * reservedDevicesTimeout).Milliseconds() {
				delete(reservedDevicesMap, udid)
			}
		}
		reserveMutex.Unlock()
	}
}

func checkDeviceReserved(deviceUDID string, reservationID string) error {
	reserveMutex.Lock()
	defer reserveMutex.Unlock()

	reservedDevice, exists := reservedDevicesMap[deviceUDID]

	if exists {
		if reservedDevice.ReservationID == reservationID || reserveAdminUUID == reservationID {
			reservedDevice.LastUsedTimestamp = time.Now().UnixMilli()
			return nil
		}
		return errors.New("Device is already reserved with another reservationID")
	}
	return errors.New("You need to reserve the device before using it")
}
