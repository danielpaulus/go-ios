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

	device.GET("/conditions", GetSupportedConditions)
	device.PUT("/enable-condition", EnableDeviceCondition)
	device.POST("/disable-condition", DisableDeviceCondition)

	device.GET("/image", GetImages)
	device.PUT("/image", InstallImage)

	device.GET("/notifications", streamingMiddleWare, Notifications)

	device.GET("/info", Info)
	device.GET("/listen", streamingMiddleWare, Listen)

	device.POST("/pair", PairDevice)
	device.GET("/profiles", GetProfiles)

	device.POST("/resetlocation", ResetLocation)
	device.GET("/screenshot", Screenshot)
	device.PUT("/setlocation", SetLocation)
	device.GET("/syslog", streamingMiddleWare, Syslog)

	device.POST("/wda/session", CreateWdaSession)
	device.GET("/wda/session/:sessionId", ReadWdaSession)
	device.DELETE("/wda/session/:sessionId", DeleteWdaSession)
}

func appRoutes(group *gin.RouterGroup) {
	router := group.Group("/apps")
	router.Use(LimitNumClientsUDID())
	router.GET("/", ListApps)
	router.POST("/launch", LaunchApp)
	router.POST("/kill", KillApp)
	router.POST("/install", InstallApp)
	router.POST("/uninstall", UninstallApp)
}
