package api

import (
	"net/http"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/afc"
	"github.com/danielpaulus/go-ios/ios/diagnostics"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/pcap"
	"github.com/gin-gonic/gin"
)

// registerDiagnosticsNetRoutes registers device diagnostics and network endpoints
// that mirror the corresponding `ios` CLI commands. All routes live under
// /device/:udid and rely on DeviceMiddleware having set the device in context.
func registerDiagnosticsNetRoutes(device *gin.RouterGroup) {
	device.GET("/diskspace", GetDiskSpace)
	device.GET("/ip", GetDeviceIP)
	device.GET("/rsd", GetRsdServices)
	device.GET("/battery/registry", GetBatteryRegistry)
}

// GetDiskSpace returns filesystem information for the device (total/free/used
// bytes, block size, ...) by querying the AFC service (CLI: ios diskspace).
// @Summary Get device disk space info
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} afc.DeviceInfo
// @Failure 500 {object} map[string]string
// @Router /device/{udid}/diskspace [get]
func GetDiskSpace(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	client, err := afc.New(device)
	if err != nil {
		golog.Error("failed to open afc for diskspace", "module", logModule, "udid", device.Properties.SerialNumber, "error", err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer client.Close()

	info, err := client.DeviceInfo()
	if err != nil {
		golog.Error("failed to read afc device info", "module", logModule, "udid", device.Properties.SerialNumber, "error", err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, info)
}

// GetDeviceIP resolves the device's network addresses (MAC/IPv4/IPv6) by sniffing
// packets over the pcapd service (CLI: ios ip).
// @Summary Get device IP / network info
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} pcap.NetworkInfo
// @Failure 500 {object} map[string]string
// @Router /device/{udid}/ip [get]
func GetDeviceIP(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	info, err := pcap.FindIp(device)
	if err != nil {
		golog.Error("failed to find device ip", "module", logModule, "udid", device.Properties.SerialNumber, "error", err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, info)
}

// GetRsdServices returns the device's RSD (Remote Service Discovery) service list
// (CLI: ios rsd ls). This requires a running tunnel (iOS 17+); devices without RSD
// get a 400.
// @Summary Get device RSD service list
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]ios.RsdServiceEntry
// @Failure 400 {object} map[string]string
// @Router /device/{udid}/rsd [get]
func GetRsdServices(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	if !device.SupportsRsd() {
		golog.Warn("rsd requested but unavailable", "module", logModule, "udid", device.Properties.SerialNumber)
		RespondError(c, http.StatusBadRequest, errRsdUnavailable)
		return
	}
	services := device.Rsd.GetServices()
	c.JSON(http.StatusOK, services)
}

// GetBatteryRegistry returns the battery IORegistry stats (Temperature, Voltage,
// CurrentCapacity, ...) via the diagnostics relay (CLI: ios diagnostics ioregistry).
// @Summary Get device battery IORegistry
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} diagnostics.IORegistry
// @Failure 500 {object} map[string]string
// @Router /device/{udid}/battery/registry [get]
func GetBatteryRegistry(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	conn, err := diagnostics.New(device)
	if err != nil {
		golog.Error("failed to open diagnostics relay", "module", logModule, "udid", device.Properties.SerialNumber, "error", err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()

	registry, err := conn.Battery()
	if err != nil {
		golog.Error("failed to read battery registry", "module", logModule, "udid", device.Properties.SerialNumber, "error", err.Error())
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, registry)
}
