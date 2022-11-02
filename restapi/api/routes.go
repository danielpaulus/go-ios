package api

import (
	"github.com/gin-gonic/gin"
)

func registerRoutes(router *gin.RouterGroup) {
	router.GET("/list", List)
	router.GET("/reserved-devices", GetReservedDevices)

	device := router.Group("/device/:udid")
	device.Use(DeviceMiddleware())
	device.GET("/info", Info)
	device.GET("/screenshot", Screenshot)
	device.PUT("/setlocation", SetLocation)
	device.POST("/resetlocation", ResetLocation)
	device.GET("/conditions", GetSupportedConditions)
	device.PUT("/enable-condition", EnableDeviceCondition)
	device.POST("/disable-condition", DisableDeviceCondition)

	device.POST("/reserve", ReserveDevice)
	device.DELETE("/reserve", ReleaseDevice)

	initAppRoutes(device)
	initStreamingResponseRoutes(device, router)
	go CleanReservationsCRON()
}
func initAppRoutes(group *gin.RouterGroup) {
	router := group.Group("/apps")
	router.Use(LimitNumClientsUDID())
	router.GET("/", ListApps)
	router.POST("/launch", LaunchApp)
	router.POST("/kill", KillApp)
}

func initStreamingResponseRoutes(device *gin.RouterGroup, router *gin.RouterGroup) {
	streamingDevice := device.Group("")
	streamingDevice.Use(StreamingHeaderMiddleware())
	streamingDevice.GET("/syslog", Syslog)
	streamingGeneral := router.Group("")
	streamingGeneral.Use(StreamingHeaderMiddleware())
	streamingGeneral.GET("/listen", Listen)
}
