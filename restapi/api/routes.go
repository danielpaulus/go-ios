package api

import (
	"github.com/gin-gonic/gin"
)

func registerRoutes(router *gin.RouterGroup) {
	router.GET("/list", List)
	router.GET("/reservations", GetReservedDevices)
	router.DELETE("/reservations/:reservationID", ReleaseDevice)

	device := router.Group("/device/:udid")
	device.Use(DeviceMiddleware())
	device.GET("/info", Info)
	device.GET("/screenshot", Screenshot)
	device.PUT("/setlocation", SetLocation)
	device.POST("/resetlocation", ResetLocation)
	device.GET("/conditions", GetSupportedConditions)
	device.PUT("/enable-condition", EnableDeviceCondition)
	device.POST("/disable-condition", DisableDeviceCondition)
	device.POST("/pair", PairDevice)

	device.POST("/reservations", ReserveDevice)
	device.GET("/profiles", GetProfiles)

	initAppRoutes(device)
	initStreamingResponseRoutes(device, router)
	go cleanReservationsCRON()
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
