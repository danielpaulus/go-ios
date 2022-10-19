package api

import (
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var sm sync.Map

var locksMap = make(map[string]*LockedDevice)
var lockMutex sync.Mutex

type GenericLockResponse struct {
	Message string `json:"message"`
}

type LockedDevice struct {
	UDID              string `json:"udid,omitempty"`
	LockID            string `json:"lock_id"`
	LastUsedTimestamp int64  `json:"lastUsed,omitempty"`
}

type LockResponse struct {
	LockID string `json:"lock_id"`
}

func CleanLocksCRON() {
	defer lockMutex.Unlock()

	for range time.Tick(time.Minute * 1) {
		lockMutex.Lock()
		for key, element := range locksMap {
			current_timestamp := time.Now().UnixMilli()
			diff := current_timestamp - element.LastUsedTimestamp

			if diff > 60000 {
				delete(locksMap, key)
			}
		}
		lockMutex.Unlock()
	}
}

func LockDevice(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	lockMutex.Lock()
	defer lockMutex.Unlock()

	udid := c.Param("udid")
	lock_id := randomLockID()
	time_now := time.Now().UnixMilli()

	lockedDevice := LockedDevice{LockID: lock_id, LastUsedTimestamp: time_now}

	map_udid := locksMap[udid]
	if map_udid == nil {
		locksMap[udid] = &lockedDevice
	} else {
		c.IndentedJSON(http.StatusOK, GenericLockResponse{Message: "Device with UDID: " + udid + " already locked."})
		return
	}

	c.IndentedJSON(http.StatusOK, LockResponse{LockID: lock_id})
}

func GetLocks(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	lockMutex.Lock()
	defer lockMutex.Unlock()

	var locked_devices []LockedDevice

	if len(locksMap) == 0 {
		c.IndentedJSON(http.StatusOK, GenericLockResponse{Message: "No locked devices found"})
		return
	} else {
		for key, element := range locksMap {
			locked_devices = append(locked_devices, LockedDevice{
				UDID:              key,
				LockID:            element.LockID,
				LastUsedTimestamp: element.LastUsedTimestamp,
			})
		}
	}

	c.IndentedJSON(http.StatusOK, locked_devices)
}

func RemoveDeviceLock(c *gin.Context) {
	udid := c.Param("udid")

	locked_device := locksMap[udid]
	if locked_device == nil {
		c.IndentedJSON(http.StatusOK, GenericLockResponse{Message: "Device with UDID: " + udid + " is not locked."})
		return
	} else {
		defer lockMutex.Unlock()
		lockMutex.Lock()
		delete(locksMap, udid)
		c.IndentedJSON(http.StatusOK, GenericLockResponse{Message: "Device with UDID: " + udid + " successfully unlocked."})
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
