package api

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/gin-gonic/gin"
	"sync"
)

func LimitNumClientsUDID() gin.HandlerFunc {
	maxClients := 1
	var semaMap = sync.Map{}
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
