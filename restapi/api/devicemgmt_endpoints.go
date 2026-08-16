package api

import (
	"net/http"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/amfi"
	"github.com/danielpaulus/go-ios/ios/diagnostics"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/mcinstall"
	"github.com/gin-gonic/gin"
)

// registerDeviceMgmtRoutes registers device-management endpoints mirroring the
// corresponding `ios` CLI commands. All routes live under /device/:udid.
func registerDeviceMgmtRoutes(device *gin.RouterGroup) {
	device.POST("/reboot", Reboot)
	device.POST("/shutdown", Shutdown)
	device.POST("/erase", Erase)
	device.GET("/devmode", GetDevMode)
	device.POST("/devmode", SetDevMode)
	device.GET("/lang", GetLanguage)
	device.PUT("/lang", SetLanguage)
	device.POST("/memlimitoff", MemLimitOff)
}

// Reboot reboots the device (CLI: ios reboot).
// @Summary Reboot the device
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/reboot [post]
func Reboot(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	if err := diagnostics.Reboot(device); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "reboot triggered"})
}

// Shutdown shuts down the device (CLI: ios shutdown).
// @Summary Shut down the device
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/shutdown [post]
func Shutdown(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	if err := diagnostics.Shutdown(device); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "shutdown triggered"})
}

// Erase erases the device (CLI: ios erase). Destructive: requires ?confirm=true.
// @Summary Erase all content and settings
// @Param udid path string true "Device UDID"
// @Param confirm query bool true "must be true to proceed"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /device/{udid}/erase [post]
func Erase(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	if c.Query("confirm") != "true" {
		RespondError(c, http.StatusBadRequest, errEraseNotConfirmed)
		return
	}
	if err := mcinstall.Erase(device); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "erase triggered"})
}

// GetDevMode reports whether developer mode is enabled (CLI: ios devmode get).
// @Summary Get developer mode state
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]interface{}
// @Router /device/{udid}/devmode [get]
func GetDevMode(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	enabled, err := imagemounter.IsDevModeEnabled(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"DeveloperModeEnabled": enabled})
}

type devModeRequest struct {
	Action            string `json:"action"`
	EnablePostRestart bool   `json:"enablePostRestart"`
}

// SetDevMode enables or reveals developer mode (CLI: ios devmode enable|reveal).
// @Summary Enable or reveal developer mode
// @Param udid path string true "Device UDID"
// @Param body body devModeRequest true "action: enable|reveal"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /device/{udid}/devmode [post]
func SetDevMode(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	var req devModeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	switch req.Action {
	case "enable":
		if err := amfi.EnableDeveloperMode(device, req.EnablePostRestart); err != nil {
			RespondError(c, http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "developer mode enable requested"})
	case "reveal":
		conn, err := amfi.New(device)
		if err != nil {
			RespondError(c, http.StatusInternalServerError, err)
			return
		}
		defer conn.Close()
		if err := conn.RevealDevMode(); err != nil {
			RespondError(c, http.StatusInternalServerError, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "developer mode menu revealed"})
	default:
		RespondError(c, http.StatusBadRequest, errUnknownAction)
	}
}

// GetLanguage returns the device language configuration (CLI: ios lang).
// @Summary Get language configuration
// @Param udid path string true "Device UDID"
// @Success 200 {object} ios.LanguageConfiguration
// @Router /device/{udid}/lang [get]
func GetLanguage(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	lang, err := ios.GetLanguage(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, lang)
}

type langRequest struct {
	Language string `json:"language"`
	Locale   string `json:"locale"`
}

// SetLanguage sets the device language/locale (CLI: ios lang --setlang --setlocale).
// @Summary Set language/locale
// @Param udid path string true "Device UDID"
// @Param body body langRequest true "language and/or locale"
// @Success 200 {object} ios.LanguageConfiguration
// @Router /device/{udid}/lang [put]
func SetLanguage(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	var req langRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	if err := ios.SetLanguage(device, ios.LanguageConfiguration{Language: req.Language, Locale: req.Locale}); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	lang, err := ios.GetLanguage(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, lang)
}

type memLimitRequest struct {
	Process string `json:"process"`
}

// MemLimitOff waives the memory limit for a process (CLI: ios memlimitoff).
// Process name via ?process= or JSON body {"process":"..."}.
// @Summary Waive the memory limit for a process
// @Param udid path string true "Device UDID"
// @Param process query string false "process name"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Router /device/{udid}/memlimitoff [post]
func MemLimitOff(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	processName := c.Query("process")
	if processName == "" {
		var req memLimitRequest
		_ = c.ShouldBindJSON(&req)
		processName = req.Process
	}
	if processName == "" {
		RespondError(c, http.StatusBadRequest, errMissingProcess)
		return
	}

	pControl, err := instruments.NewProcessControl(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer pControl.Close()

	svc, err := instruments.NewDeviceInfoService(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer svc.Close()

	process, err := svc.ProcessByName(processName)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	disabled, err := pControl.DisableMemoryLimit(process.Pid)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"process": process.Name, "pid": process.Pid, "disabled": disabled})
}
