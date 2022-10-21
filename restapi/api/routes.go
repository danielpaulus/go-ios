package api

import (
	"github.com/danielpaulus/go-ios/restapi/api/lock"
	"github.com/gin-gonic/gin"
)

func registerRoutes(router *gin.RouterGroup) {
	router.GET("/list", List)
	device := router.Group("/device/:udid")
	device.Use(DeviceMiddleware())
	device.GET("/info", Info)
	device.GET("/screenshot", Screenshot)

	locks := router.Group("/lock/:udid")
	locks.Use(DeviceMiddleware())
	locks.POST("/", lock.LockDevice)
	locks.DELETE("/", lock.DeleteDeviceLock)

	router.GET("/locks", lock.GetLockedDevices)

	initAppRoutes(device)
	initStreamingResponseRoutes(device, router)
	go lock.CleanLocksCRON()
}
func initAppRoutes(group *gin.RouterGroup) {
	router := group.Group("/app")
	router.GET("/", ListApps)
}

func initStreamingResponseRoutes(device *gin.RouterGroup, router *gin.RouterGroup) {
	streamingDevice := device.Group("")
	streamingDevice.Use(StreamingHeaderMiddleware())
	streamingDevice.GET("/syslog", Syslog)
	streamingGeneral := router.Group("")
	streamingGeneral.Use(StreamingHeaderMiddleware())
	streamingGeneral.GET("/listen", Listen)
}
