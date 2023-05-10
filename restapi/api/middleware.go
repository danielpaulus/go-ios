package api

import (
	"net/http"
	"strings"
	"sync"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/gin-gonic/gin"
)

// DeviceMiddleware makes sure a udid was specified and that a device with that UDID
// is connected with the host. Will return 404 if the device is not found or 500 if something
// else went wrong. Use `device := c.MustGet(IOS_KEY).(ios.DeviceEntry)` to acquire the device
// in downstream handlers.
func DeviceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		udid := c.Param("udid")

		if udid == "" {
			c.AbortWithStatusJSON(http.StatusUnprocessableEntity, gin.H{"message": "udid is missing"})
			return
		}
		device, err := ios.GetDevice(udid)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"message": "device not found on the host"})
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		c.Set(IOS_KEY, device)
		c.Next()
	}
}

const IOS_KEY = "go_ios_device"

// LimitNumClientsUDID limits clients to one concurrent connection per device UDID at a time
func LimitNumClientsUDID() gin.HandlerFunc {
	maxClients := 1
	semaMap := sync.Map{}
	return func(c *gin.Context) {
		device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
		udid := device.Properties.SerialNumber
		var sema chan struct{}
		semaIntf, ok := semaMap.Load(udid)
		if !ok {
			sema = make(chan struct{}, maxClients)
			semaMap.Store(udid, sema)
		} else {
			sema = semaIntf.(chan struct{})
		}
		sema <- struct{}{}
		defer func() { <-sema }()
		c.Next()
		print("mid done")
	}
}

// StreamingHeaderMiddleware adds event-streaming headers
func StreamingHeaderMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")
		c.Next()
	}
}

func ReserveDevicesMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		udid := c.Param("udid")
		reservationID := c.Request.Header.Get("X-GO-IOS-RESERVE")
		if reservationID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, GenericResponse{Error: "Device reservation token empty or missing. This operation requires you to reserve the device using the /reservations endpoint and then pass the reservation token with the X-GO-IOS-RESERVE header"})
			return
		}

		err := checkDeviceReserved(udid, reservationID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, GenericResponse{Error: err.Error()})
			return
		}
		c.Next()
	}
}
