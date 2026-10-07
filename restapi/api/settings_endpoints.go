package api

import (
	"net/http"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/mcinstall"
	"github.com/gin-gonic/gin"
)

// registerSettingsRoutes registers device-settings endpoints (accessibility and
// wifi) mirroring the corresponding `ios` CLI commands. Routes under /device/:udid.
func registerSettingsRoutes(device *gin.RouterGroup) {
	device.GET("/assistivetouch", GetAssistiveTouch)
	device.PUT("/assistivetouch", SetAssistiveTouch)
	device.GET("/timeformat", GetTimeFormat)
	device.PUT("/timeformat", SetTimeFormat)
	device.PUT("/wifi", SetWifi)
	device.DELETE("/wifi", RemoveWifi)
}

type enabledRequest struct {
	Enabled bool `json:"enabled"`
}

// GetAssistiveTouch reports whether AssistiveTouch is enabled (CLI: ios assistivetouch get).
// @Summary Get AssistiveTouch state
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]bool
// @Router /device/{udid}/assistivetouch [get]
func GetAssistiveTouch(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	enabled, err := ios.GetAssistiveTouch(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"AssistiveTouchEnabled": enabled})
}

// SetAssistiveTouch enables/disables AssistiveTouch (CLI: ios assistivetouch enable|disable).
// @Summary Set AssistiveTouch state
// @Param udid path string true "Device UDID"
// @Param body body enabledRequest true "enabled"
// @Success 200 {object} map[string]bool
// @Router /device/{udid}/assistivetouch [put]
func SetAssistiveTouch(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	var req enabledRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	if err := ios.SetAssistiveTouch(device, req.Enabled); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"AssistiveTouchEnabled": req.Enabled})
}

// GetTimeFormat reports whether the device uses a 24-hour clock (CLI: ios timeformat get).
// @Summary Get time-format state
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]bool
// @Router /device/{udid}/timeformat [get]
func GetTimeFormat(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	uses24, err := ios.GetUses24HourClock(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"Uses24HourClock": uses24})
}

type timeFormatRequest struct {
	Uses24Hour bool `json:"uses24Hour"`
}

// SetTimeFormat sets 24h/12h clock (CLI: ios timeformat 24h|12h).
// @Summary Set time format
// @Param udid path string true "Device UDID"
// @Param body body timeFormatRequest true "uses24Hour"
// @Success 200 {object} map[string]bool
// @Router /device/{udid}/timeformat [put]
func SetTimeFormat(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	var req timeFormatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	if err := ios.SetUses24HourClock(device, req.Uses24Hour); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"Uses24HourClock": req.Uses24Hour})
}

type wifiRequest struct {
	SSID     string `json:"ssid"`
	Password string `json:"password"`
	EncType  string `json:"encType"`
}

// SetWifi provisions a wifi network (CLI: ios wifi).
// @Summary Provision a wifi network
// @Param udid path string true "Device UDID"
// @Param body body wifiRequest true "ssid/password/encType"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/wifi [put]
func SetWifi(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	var req wifiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	if req.SSID == "" {
		RespondError(c, http.StatusBadRequest, errMissingSSID)
		return
	}
	if err := mcinstall.PrepareWifi(device, req.SSID, req.Password, req.EncType); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "wifi provisioned", "ssid": req.SSID})
}

// RemoveWifi removes a provisioned wifi network (CLI: ios wifi --remove).
// @Summary Remove a provisioned wifi network
// @Param udid path string true "Device UDID"
// @Param ssid query string true "network SSID"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/wifi [delete]
func RemoveWifi(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	ssid := c.Query("ssid")
	if ssid == "" {
		RespondError(c, http.StatusBadRequest, errMissingSSID)
		return
	}
	if err := mcinstall.RemoveWifi(device, ssid); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "wifi removed", "ssid": ssid})
}
