package reservation

import (
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var reservedDevicesMap = make(map[string]*reservedDevice)
var reserveMutex sync.Mutex

type reservedDevice struct {
	Message           string `json:"message,omitempty"`
	UDID              string `json:"udid,omitempty"`
	ReservationID     string `json:"reservationID,omitempty"`
	LastUsedTimestamp int64  `json:"lastUsed,omitempty"`
}

func CleanReservationsCRON() {
	defer reserveMutex.Unlock()

	// Every minute loop through the map of reserved devices and check if a reserved device last used timestamp was more than 5 minutes(300000 ms) ago
	// If any, remove them from the map
	for range time.Tick(time.Second * 60) {
		reserveMutex.Lock()
		for udid, reservedDevice := range reservedDevicesMap {
			currentTimestamp := time.Now().UnixMilli()
			diff := currentTimestamp - reservedDevice.LastUsedTimestamp

			if diff > 300000 {
				delete(reservedDevicesMap, udid)
			}
		}
		reserveMutex.Unlock()
	}
}

// Reserve device access
// List          godoc
// @Summary      Reserve a device
// @Description  Reserve a device by provided UDID
// @Tags         reserve
// @Produce      json
// @Success      200  {object} reservedDevice
// @Router       /reserve/:udid [post]
func ReserveDevice(c *gin.Context) {
	udid := c.Param("udid")
	reservationID := randomReservationID()

	reserveMutex.Lock()
	defer reserveMutex.Unlock()

	// Check if there is a reserved device for the respective UDID
	device := reservedDevicesMap[udid]
	if device == nil {
		newReservedDevice := reservedDevice{ReservationID: reservationID, LastUsedTimestamp: time.Now().UnixMilli()}
		reservedDevicesMap[udid] = &newReservedDevice
	} else {
		c.IndentedJSON(http.StatusOK, reservedDevice{Message: "Already reserved"})
		return
	}

	c.IndentedJSON(http.StatusOK, reservedDevice{ReservationID: reservationID})
}

// Release device access
// List          godoc
// @Summary      Release a device
// @Description  Release a device by provided UDID
// @Tags         reserve
// @Produce      json
// @Success      200  {object} reservedDevice
// @Failure      404  {object} reservedDevice
// @Router       /reserve/:udid [delete]
func ReleaseDevice(c *gin.Context) {
	udid := c.Param("udid")

	reserveMutex.Lock()
	defer reserveMutex.Unlock()

	// Check if there is a reserved device for the respective UDID
	device := reservedDevicesMap[udid]
	if device == nil {
		c.IndentedJSON(http.StatusNotFound, reservedDevice{Message: "Not reserved"})
		return
	} else {
		delete(reservedDevicesMap, udid)
		c.IndentedJSON(http.StatusOK, reservedDevice{Message: "Successfully released"})
	}
}

// Get all reserved devices
// List          godoc
// @Summary      Get a list of reserved devices
// @Description  Get a list of reserved devices with UDID, ReservationID and last used timestamp
// @Tags         reserve
// @Produce      json
// @Success      200  {object} []reservedDevice
// @Router       /reserved-devices [get]
func GetReservedDevices(c *gin.Context) {
	reserveMutex.Lock()
	defer reserveMutex.Unlock()

	var reserved_devices []reservedDevice
	if len(reservedDevicesMap) == 0 {
		c.IndentedJSON(http.StatusOK, reservedDevice{Message: "No reserved devices found"})
		return
	} else {
		// Build the JSON array of currently reserved devices
		for udid, device := range reservedDevicesMap {
			reserved_devices = append(reserved_devices, reservedDevice{
				UDID:              udid,
				ReservationID:     device.ReservationID,
				LastUsedTimestamp: device.LastUsedTimestamp,
			})
		}
	}

	c.IndentedJSON(http.StatusOK, reserved_devices)
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func randomReservationID() string {
	rand.Seed(time.Now().UnixNano())

	sb := strings.Builder{}
	sb.Grow(36)
	for i := 0; i < 36; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}

	return sb.String()
}
