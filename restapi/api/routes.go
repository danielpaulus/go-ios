package api

import (
	"github.com/danielpaulus/go-ios/restapi/api/reservation"
	"github.com/gin-gonic/gin"
)

func registerRoutes(router *gin.RouterGroup) {
	router.GET("/list", List)
	device := router.Group("/device/:udid")
	device.Use(DeviceMiddleware())
	device.GET("/info", Info)
	device.GET("/screenshot", Screenshot)
	device.PUT("/setlocation", SetLocation)
	device.POST("/resetlocation", ResetLocation)

	router.GET("/reserved-devices", reservation.GetReservedDevices)
	reservations := router.Group("/reserve/:udid")
	reservations.Use(DeviceMiddleware())
	reservations.POST("/", reservation.ReserveDevice)
	reservations.DELETE("/", reservation.ReleaseDevice)

	initAppRoutes(device)
	initStreamingResponseRoutes(device, router)
	go reservation.CleanReservationsCRON()
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
