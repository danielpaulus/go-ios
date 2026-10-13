package api

import (
	"net/http"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/mcinstall"
	"github.com/gin-gonic/gin"
)

// registerProvisioningDeviceRoutes registers device-scoped provisioning /
// configuration endpoints mirroring the `ios` CLI. Routes live under
// /device/:udid.
func registerProvisioningDeviceRoutes(device *gin.RouterGroup) {
	device.GET("/cloudconfig", GetCloudConfig)
}

// registerProvisioningHostRoutes registers fleet/host-level provisioning
// endpoints that are not scoped to a device. They live directly under the API
// root (e.g. /api/v1/prepare/...).
func registerProvisioningHostRoutes(router *gin.RouterGroup) {
	router.GET("/prepare/skip-options", GetPrepareSkipOptions)
}

// GetCloudConfig returns the device cloud configuration (CLI: ios devicestate /
// mcinstall GetCloudConfiguration). Includes supervision status, skip-setup
// options and organization info.
// @Summary Get the device cloud configuration
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]interface{}
// @Router /device/{udid}/cloudconfig [get]
func GetCloudConfig(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	conn, err := mcinstall.New(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()
	cfg, err := conn.GetCloudConfiguration()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, cfg)
}

// GetPrepareSkipOptions returns all setup-pane skip options usable when preparing
// a device (CLI: ios prepare printskip). This is a static, device-free list.
// @Summary List all available setup skip options
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /prepare/skip-options [get]
func GetPrepareSkipOptions(c *gin.Context) {
	options := mcinstall.GetAllSetupSkipOptions()
	c.JSON(http.StatusOK, gin.H{"options": options, "count": len(options)})
}
