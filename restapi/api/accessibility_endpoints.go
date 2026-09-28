package api

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/accessibility"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/simlocation"
	"github.com/gin-gonic/gin"
)

// defaultAXTimeout bounds a single accessibility service call (audit run or
// element snapshot) so a wedged device service cannot hold a request open
// indefinitely. Clients may shorten the audit with the `timeout` query param
// (seconds); the request context still applies on top.
const defaultAXTimeout = 60 * time.Second

// registerAccessibilityRoutes registers accessibility (VoiceOver, ZoomTouch, the
// AX audit and element snapshot) and GPX location endpoints mirroring the
// corresponding `ios` CLI commands. All routes live under /device/:udid.
func registerAccessibilityRoutes(device *gin.RouterGroup) {
	device.GET("/voiceover", GetVoiceOver)
	device.PUT("/voiceover", SetVoiceOver)
	device.GET("/zoom", GetZoomTouch)
	device.PUT("/zoom", SetZoomTouch)
	device.POST("/ax/audit", RunAXAudit)
	device.GET("/ax", GetAXSnapshot)
	device.PUT("/setlocation/gpx", SetLocationGPX)
}

// parseEnabled resolves the desired boolean state from either a JSON body
// ({"enabled": true}) or an `enabled` query param, so callers can use whichever
// is convenient. A parseable JSON body wins; otherwise the query param is used.
func parseEnabled(c *gin.Context) (bool, error) {
	if c.Request != nil && c.Request.ContentLength != 0 {
		var req enabledRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			return req.Enabled, nil
		} else if c.Query("enabled") == "" {
			// The body was supplied but unparseable and there is no query fallback.
			return false, err
		}
	}
	return strconv.ParseBool(c.Query("enabled"))
}

// GetVoiceOver reports whether VoiceOver is enabled (CLI: ios voiceover get).
// @Summary Get VoiceOver state
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]bool
// @Router /device/{udid}/voiceover [get]
func GetVoiceOver(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	enabled, err := ios.GetVoiceOver(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"VoiceOverEnabled": enabled})
}

// SetVoiceOver enables/disables VoiceOver (CLI: ios voiceover enable|disable).
// The desired state comes from a JSON body {"enabled": bool} or the `enabled`
// query param.
// @Summary Set VoiceOver state
// @Param udid path string true "Device UDID"
// @Param body body enabledRequest false "enabled"
// @Param enabled query bool false "enabled"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Router /device/{udid}/voiceover [put]
func SetVoiceOver(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	enabled, err := parseEnabled(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	if err := ios.SetVoiceOver(device, enabled); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"VoiceOverEnabled": enabled})
}

// GetZoomTouch reports whether ZoomTouch is enabled (CLI: ios zoomtouch get).
// @Summary Get ZoomTouch state
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]bool
// @Router /device/{udid}/zoom [get]
func GetZoomTouch(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	enabled, err := ios.GetZoomTouch(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ZoomTouchEnabled": enabled})
}

// SetZoomTouch enables/disables ZoomTouch (CLI: ios zoomtouch enable|disable).
// The desired state comes from a JSON body {"enabled": bool} or the `enabled`
// query param.
// @Summary Set ZoomTouch state
// @Param udid path string true "Device UDID"
// @Param body body enabledRequest false "enabled"
// @Param enabled query bool false "enabled"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Router /device/{udid}/zoom [put]
func SetZoomTouch(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	enabled, err := parseEnabled(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	if err := ios.SetZoomTouch(device, enabled); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ZoomTouchEnabled": enabled})
}

// RunAXAudit runs the accessibility audit against the currently focused app and
// returns the issues found (CLI: ios ax audit). The run is bounded by the
// `timeout` query param (seconds, default 60) on top of the request context.
// @Summary Run the accessibility audit
// @Produce json
// @Param udid path string true "Device UDID"
// @Param timeout query int false "audit timeout in seconds (default 60)"
// @Success 200 {array} interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /device/{udid}/ax/audit [post]
func RunAXAudit(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	timeout := defaultAXTimeout
	if q := c.Query("timeout"); q != "" {
		secs, err := strconv.Atoi(q)
		if err != nil || secs <= 0 {
			RespondError(c, http.StatusBadRequest, errInvalidTimeout)
			return
		}
		timeout = time.Duration(secs) * time.Second
	}

	control, err := accessibility.NewWithoutEventChangeListeners(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer control.Close()

	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	issues, err := control.RunAudit(ctx)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	if issues == nil {
		issues = []accessibility.AXAuditIssue{}
	}
	c.JSON(http.StatusOK, issues)
}

// GetAXSnapshot returns a snapshot of the currently focused accessibility
// element (CLI: ios ax). Live event streaming is a separate wave; this returns
// only the current element.
// @Summary Get an accessibility element snapshot
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} interface{}
// @Failure 500 {object} map[string]string
// @Router /device/{udid}/ax [get]
func GetAXSnapshot(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	control, err := accessibility.NewWithoutEventChangeListeners(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer control.Close()

	ctx, cancel := context.WithTimeout(c.Request.Context(), defaultAXTimeout)
	defer cancel()

	element, err := control.GetElement(ctx)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, element)
}

// SetLocationGPX simulates live location tracking from an uploaded GPX file
// (CLI: ios setlocationgpx). Send multipart/form-data with a "gpx" file; it is
// written to a temp file, replayed, and the temp file is removed afterwards.
// @Summary Simulate location from a GPX file
// @Accept multipart/form-data
// @Param udid path string true "Device UDID"
// @Param gpx formData file true "GPX file"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /device/{udid}/setlocation/gpx [put]
func SetLocationGPX(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	gpxBytes, err := readFormFile(c, "gpx")
	if err != nil {
		RespondError(c, http.StatusBadRequest, errMissingGPX)
		return
	}
	if len(gpxBytes) == 0 {
		RespondError(c, http.StatusBadRequest, errMissingGPX)
		return
	}

	tmp, err := os.CreateTemp(os.TempDir(), "go-ios-gpx-*.gpx")
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	tmpPath := tmp.Name()
	defer func() {
		if err := os.Remove(tmpPath); err != nil {
			golog.Warn("failed to remove temp gpx file", "module", logModule, "udid", device.Properties.SerialNumber, "path", tmpPath, "err", err.Error())
		}
	}()

	if _, err := tmp.Write(gpxBytes); err != nil {
		tmp.Close()
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	if err := tmp.Close(); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}

	if err := simlocation.SetLocationGPX(device, tmpPath); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "gpx location simulation completed"})
}
