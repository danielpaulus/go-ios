package api

import (
	"net/http"
	"strings"
	"sync"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tunnel"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
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

		info, err := tunnel.TunnelInfoForDevice(device.Properties.SerialNumber, ios.HttpApiHost(), ios.HttpApiPort())
		if err == nil {
			log.WithField("udid", device.Properties.SerialNumber).Printf("Received tunnel info %v", info)

			device.UserspaceTUNPort = info.UserspaceTUNPort
			device.UserspaceTUN = info.UserspaceTUN

			device, err = deviceWithRsdProvider(device, udid, info.Address, info.RsdPort)
			if err != nil {
				c.Error(err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()}) // Return an error response
				c.Next()
			}
		} else {
			log.WithField("udid", device.Properties.SerialNumber).Warn("failed to get tunnel info")
		}

		c.Set(IOS_KEY, device)
		c.Next()
	}
}

func deviceWithRsdProvider(device ios.DeviceEntry, udid string, address string, rsdPort int) (ios.DeviceEntry, error) {
	rsdService, err := ios.NewWithAddrPortDevice(address, rsdPort, device)
	if err != nil {
		return device, err
	}

	defer rsdService.Close()
	rsdProvider, err := rsdService.Handshake()
	if err != nil {
		return device, err
	}

	device1, err := ios.GetDeviceWithAddress(udid, address, rsdProvider)
	if err != nil {
		return device, err
	}

	device1.UserspaceTUN = device.UserspaceTUN
	device1.UserspaceTUNPort = device.UserspaceTUNPort

	return device1, nil
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
