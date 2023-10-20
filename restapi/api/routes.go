package api

import (
	"github.com/gin-gonic/gin"
)

var streamingMiddleWare = StreamingHeaderMiddleware()

func registerRoutes(router *gin.RouterGroup) {
	router.GET("/list", List)

	device := router.Group("/device/:udid")
	device.Use(DeviceMiddleware())
	simpleDeviceRoutes(device)
	appRoutes(device)
}

func simpleDeviceRoutes(device *gin.RouterGroup) {
	device.POST("/activate", Activate)
	device.GET("/info", Info)
	device.GET("/screenshot", Screenshot)
	device.PUT("/setlocation", SetLocation)
	device.POST("/resetlocation", ResetLocation)
	device.GET("/conditions", GetSupportedConditions)
	device.PUT("/enable-condition", EnableDeviceCondition)
	device.POST("/disable-condition", DisableDeviceCondition)
	device.POST("/pair", PairDevice)

	device.GET("/profiles", GetProfiles)

	device.GET("/syslog", streamingMiddleWare, Syslog)
	device.GET("/listen", streamingMiddleWare, Listen)
}

func appRoutes(group *gin.RouterGroup) {
	router := group.Group("/apps")
	router.Use(LimitNumClientsUDID())
	router.GET("/", ListApps)
	router.POST("/launch", LaunchApp)
	router.POST("/kill", KillApp)
}
