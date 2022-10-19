package lock

import (
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var locksMap = make(map[string]*lockedDevice)
var lockMutex sync.Mutex

type genericlockResponse struct {
	Message string `json:"message"`
}

type lockResponse struct {
	LockID string `json:"lock_id"`
}

type lockedDevice struct {
	UDID              string `json:"udid,omitempty"`
	LockID            string `json:"lock_id"`
	LastUsedTimestamp int64  `json:"lastUsed,omitempty"`
}

func CleanLocksCRON() {
	defer lockMutex.Unlock()

	// Every 5 minutes loop through the map of locked devices and check if a locked device last used timestamp was more than 5 minutes(300000 ms) ago
	for range time.Tick(time.Minute * 5) {
		lockMutex.Lock()
		for key, element := range locksMap {
			currentTimestamp := time.Now().UnixMilli()
			diff := currentTimestamp - element.LastUsedTimestamp

			if diff > 300000 {
				delete(locksMap, key)
			}
		}
		lockMutex.Unlock()
	}
}

func LockDevice(c *gin.Context) {
	udid := c.Param("udid")
	lock_id := randomLockID()

	lockMutex.Lock()
	defer lockMutex.Unlock()

	// Check if there is a locked device for the respective UDID
	device := locksMap[udid]
	if device == nil {
		newLockedDevice := lockedDevice{LockID: lock_id, LastUsedTimestamp: time.Now().UnixMilli()}
		locksMap[udid] = &newLockedDevice
	} else {
		c.IndentedJSON(http.StatusOK, genericlockResponse{Message: "Already locked"})
		return
	}

	c.IndentedJSON(http.StatusOK, lockResponse{LockID: lock_id})
}

func GetLocks(c *gin.Context) {
	lockMutex.Lock()
	defer lockMutex.Unlock()

	var locked_devices []lockedDevice
	if len(locksMap) == 0 {
		c.IndentedJSON(http.StatusOK, genericlockResponse{Message: "No locked devices found"})
		return
	} else {
		// Build the JSON array of currently locked devices
		for udid, device := range locksMap {
			locked_devices = append(locked_devices, lockedDevice{
				UDID:              udid,
				LockID:            device.LockID,
				LastUsedTimestamp: device.LastUsedTimestamp,
			})
		}
	}

	c.IndentedJSON(http.StatusOK, locked_devices)
}

func RemoveDeviceLock(c *gin.Context) {
	udid := c.Param("udid")

	defer lockMutex.Unlock()
	lockMutex.Lock()

	// Check if there is a locked device for the respective UDID
	device := locksMap[udid]
	if device == nil {
		c.IndentedJSON(http.StatusOK, genericlockResponse{Message: "Not locked"})
		return
	} else {
		delete(locksMap, udid)
		c.IndentedJSON(http.StatusOK, genericlockResponse{Message: "Successfully unlocked"})
	}
}

const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func randomLockID() string {
	rand.Seed(time.Now().UnixNano())

	sb := strings.Builder{}
	sb.Grow(36)
	for i := 0; i < 36; i++ {
		sb.WriteByte(charset[rand.Intn(len(charset))])
	}

	return sb.String()
}
