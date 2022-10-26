package api

import "github.com/gin-gonic/gin"

func registerRoutes(router *gin.RouterGroup) {
	router.GET("/list", List)
	device := router.Group("/device/:udid")
	device.Use(DeviceMiddleware())
	device.GET("/info", Info)
	device.GET("/screenshot", Screenshot)
	device.POST("/setlocation", SetLocation)
	device.POST("/resetlocation", ResetLocation)
	device.POST("/enablestate", EnableDeviceState)

	initAppRoutes(device)
	initStreamingResponseRoutes(device, router)
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
