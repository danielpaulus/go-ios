package api

import (
	"net/http"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/diagnostics"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/gin-gonic/gin"
)

// registerDeviceInfoRoutes registers read-only device information endpoints that
// mirror the corresponding `ios` CLI commands. All routes live under
// /device/:udid and rely on DeviceMiddleware having set the device in context.
func registerDeviceInfoRoutes(device *gin.RouterGroup) {
	device.GET("/devicename", GetDeviceName)
	device.GET("/date", GetDeviceDate)
	device.GET("/battery", GetBattery)
	device.GET("/diagnostics", GetDiagnostics)
	device.GET("/mobilegestalt", GetMobileGestalt)
	device.GET("/processes", GetProcesses)
	device.GET("/lockdown", GetLockdownValues)
}

// GetDeviceName returns the device name (CLI: ios devicename).
// @Summary Get device name
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/devicename [get]
func GetDeviceName(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	allValues, err := ios.GetValues(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"devicename": allValues.Value.DeviceName})
}

// GetDeviceDate returns the device date (CLI: ios date).
// @Summary Get device date
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]interface{}
// @Router /device/{udid}/date [get]
func GetDeviceDate(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	allValues, err := ios.GetValues(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	ts := allValues.Value.TimeIntervalSince1970
	c.JSON(http.StatusOK, gin.H{
		"formatedDate":          time.Unix(int64(ts), 0).Format(time.RFC850),
		"TimeIntervalSince1970": ts,
	})
}

// GetBattery returns battery diagnostics (CLI: ios batterycheck).
// @Summary Get battery diagnostics
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} ios.BatteryInfo
// @Router /device/{udid}/battery [get]
func GetBattery(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	battery, err := ios.GetBatteryDiagnostics(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, battery)
}

// GetDiagnostics returns all diagnostic values (CLI: ios diagnostics list).
// @Summary List diagnostics
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} interface{}
// @Router /device/{udid}/diagnostics [get]
func GetDiagnostics(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	conn, err := diagnostics.New(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()
	values, err := conn.AllValues()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, values)
}

// GetMobileGestalt queries mobilegestalt keys (CLI: ios mobilegestalt <key>...).
// Pass one or more keys as repeated query params, e.g. ?key=A&key=B.
// @Summary Query mobilegestalt keys
// @Produce json
// @Param udid path string true "Device UDID"
// @Param key query []string true "mobilegestalt keys"
// @Success 200 {object} interface{}
// @Failure 400 {object} map[string]string
// @Router /device/{udid}/mobilegestalt [get]
func GetMobileGestalt(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	keys := c.QueryArray("key")
	if len(keys) == 0 {
		RespondError(c, http.StatusBadRequest, errMissingKey)
		return
	}
	conn, err := diagnostics.New(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()
	result, err := conn.MobileGestaltQuery(keys)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetProcesses lists running processes (CLI: ios ps [--apps]).
// Pass ?apps=true to return only application processes.
// @Summary List running processes
// @Produce json
// @Param udid path string true "Device UDID"
// @Param apps query bool false "only application processes"
// @Success 200 {array} instruments.ProcessInfo
// @Router /device/{udid}/processes [get]
func GetProcesses(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	service, err := instruments.NewDeviceInfoService(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer service.Close()
	processList, err := service.ProcessList()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	if c.Query("apps") == "true" {
		apps := make([]instruments.ProcessInfo, 0, len(processList))
		for _, p := range processList {
			if p.IsApplication {
				apps = append(apps, p)
			}
		}
		processList = apps
	}
	c.JSON(http.StatusOK, processList)
}

// GetLockdownValues returns all lockdown values (CLI: ios lockdown get).
// @Summary Get lockdown values
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} interface{}
// @Router /device/{udid}/lockdown [get]
func GetLockdownValues(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	allValues, err := ios.GetValues(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, allValues)
}
