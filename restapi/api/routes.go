package api

import (
	"github.com/gin-gonic/gin"
)

var streamingMiddleWare = StreamingHeaderMiddleware()

func registerRoutes(router *gin.RouterGroup, rateLimit float64, rateBurst int) {
	router.GET("/list", List)
	registerTunnelRoutes(router)
	// feat/restapi-w1c-fsync: fleet/host-level provisioning routes (no device).
	registerProvisioningHostRoutes(router)

	device := router.Group("/device/:udid")
	device.Use(DeviceMiddleware())
	device.Use(RateLimitUDID(rateLimit, rateBurst))
	simpleDeviceRoutes(device)
	registerDeviceInfoRoutes(device)
	// --- feat/restapi-w1a-diagnostics: diagnostics & network parity endpoints ---
	registerDiagnosticsNetRoutes(device)
	// --- end feat/restapi-w1a-diagnostics ---
	registerDeviceMgmtRoutes(device)
	registerFilesRoutes(device)
	registerMediaRoutes(device)
	registerConfigRoutes(device)
	registerSettingsRoutes(device)
	registerMonitoringRoutes(device)
	registerMdmRoutes(device)
	registerJobRoutes(device)
	registerProxyRoutes(device)
	// w1b accessibility + location parity (feat/restapi-w1b-a11y)
	registerAccessibilityRoutes(device)
	// feat/restapi-w1c-fsync: AFC filesystem + cloud-config device routes.
	registerFsyncRoutes(device)
	registerProvisioningDeviceRoutes(device)
	// feat/restapi-w1d-webinspector: WebInspector device routes.
	registerWebInspectorRoutes(device)
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

	device.POST("/resetaccessibility", ResetAccessibility)
	device.POST("/resetlocation", ResetLocation)
	device.GET("/screenshot", Screenshot)
	device.PUT("/setlocation", SetLocation)
	device.GET("/syslog", streamingMiddleWare, Syslog)
	device.GET("/ostrace", streamingMiddleWare, OsTrace)

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
