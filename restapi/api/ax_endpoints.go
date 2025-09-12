package api

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/accessibility"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

var (
	axConn      *accessibility.ControlInterface
	axConnMux   sync.RWMutex
	isAXEnabled bool
)

// enableAXService enables the accessibility service session for the device
// @Summary      Enable accessibility service
// @Description  Starts an accessibility session on the device and enables selection mode. Optionally accepts WDA host for alert detection.
// @Tags         accessibility
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Param        wda_host query string false "WDA host for alert detection (e.g., http://192.168.2.196:8100)"
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  GenericResponse
// @Router       /device/{udid}/accessibility/enable [get]
func enableAXService(c *gin.Context) {
	axConnMux.Lock()
	defer axConnMux.Unlock()

	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	// Get WDA host from query parameter
	wdaHost := c.Query("wda_host")
	log.Infof("enableAXService called with wda_host: %q", wdaHost)

	// If service is already enabled, just update WDA host if provided
	if isAXEnabled && axConn != nil {
		log.Infof("Service already enabled, updating WDA host to: %q", wdaHost)
		if wdaHost != "" {
			axConn.SetWDAHost(wdaHost)
			c.JSON(http.StatusOK, map[string]string{"message": "Accessibility service already enabled, WDA host updated"})
		} else {
			axConn.ClearWDAHost()
			c.JSON(http.StatusOK, map[string]string{"message": "Accessibility service already enabled, WDA host cleared"})
		}
		return
	}

	// Create new connection with or without WDA host
	var conn *accessibility.ControlInterface
	var err error
	if wdaHost != "" {
		log.Infof("Creating connection with WDA host: %q", wdaHost)
		conn, err = accessibility.NewWithWDA(device, wdaHost)
	} else {
		log.Infof("Creating connection without WDA host")
		conn, err = accessibility.New(device)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	// Verify WDA host was set correctly
	log.Infof("Connection created, WDA host is: %q", conn.GetWDAHost())

	conn.SwitchToDevice()

	conn.EnableSelectionMode()

	axConn = conn
	isAXEnabled = true
	log.Infof("Accessibility service enabled, final WDA host: %q", axConn.GetWDAHost())
	c.JSON(http.StatusOK, map[string]string{"message": "Accessibility service enabled"})
}

// disableAXService disables the accessibility service session for the device
// @Summary      Disable accessibility service
// @Description  Turns off the accessibility session on the device
// @Tags         accessibility
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Success      200  {object}  map[string]string
// @Router       /device/{udid}/accessibility/disable [get]
func disableAXService(c *gin.Context) {

	if !isAXEnabled || axConn == nil {
		c.JSON(http.StatusOK, map[string]string{"message": "Accessibility service already disabled"})
		return
	}

	axConn.TurnOff()

	// Close the connection // Assuming there's a Close method
	axConn = nil
	isAXEnabled = false
	c.JSON(http.StatusOK, map[string]string{"message": "Accessibility service disabled"})
}

// navigateToNextElement moves VoiceOver focus to the next element
// @Summary      Navigate to next accessible element
// @Description  Moves the selection to the next element using accessibility service
// @Tags         accessibility
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  GenericResponse
// @Router       /device/{udid}/accessibility/next [post]
func navigateToNextElement(c *gin.Context) {
	axConnMux.RLock()
	defer axConnMux.RUnlock()

	if !isAXEnabled || axConn == nil {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "Accessibility service not enabled. Call /enable first"})
		return
	}

	axConn.Navigate(accessibility.DirectionNext)

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Navigated to next element",
	})
}

// navigateToPrevElement moves VoiceOver focus to the previous element
// @Summary      Navigate to previous accessible element
// @Description  Moves the selection to the previous element using accessibility service
// @Tags         accessibility
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  GenericResponse
// @Router       /device/{udid}/accessibility/previous [post]
func navigateToPrevElement(c *gin.Context) {
	axConnMux.RLock()
	defer axConnMux.RUnlock()

	if !isAXEnabled || axConn == nil {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "Accessibility service not enabled. Call /enable first"})
		return
	}

	axConn.Navigate(accessibility.DirectionPrevious)

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Navigated to previous element",
	})
}

// navigateToFirstElement moves VoiceOver focus to the first element
// @Summary      Navigate to first accessible element
// @Description  Moves the selection to the first element using accessibility service
// @Tags         accessibility
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  GenericResponse
// @Router       /device/{udid}/accessibility/first [post]
func navigateToFirstElement(c *gin.Context) {
	axConnMux.RLock()
	defer axConnMux.RUnlock()

	if !isAXEnabled || axConn == nil {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "Accessibility service not enabled. Call /enable first"})
		return
	}

	axConn.Navigate(accessibility.DirectionFirst)

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Navigated to first element",
	})
}

// navigateToLastElement moves VoiceOver focus to the last element
// @Summary      Navigate to last accessible element
// @Description  Moves the selection to the last element using accessibility service
// @Tags         accessibility
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  GenericResponse
// @Router       /device/{udid}/accessibility/last [post]
func navigateToLastElement(c *gin.Context) {
	axConnMux.RLock()
	defer axConnMux.RUnlock()

	if !isAXEnabled || axConn == nil {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "Accessibility service not enabled. Call /enable first"})
		return
	}

	axConn.Navigate(accessibility.DirectionLast)

	c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Navigated to last element",
	})
}

// performDtxAction performs an accessibility action via the DTX channel
// @Summary      Perform accessibility action via DTX
// @Description  Performs the default action on the current focused element using DTX. Optionally accepts WDA host for alert detection.
// @Tags         accessibility
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Param        wda_host query string false "WDA host for alert detection (e.g., http://192.168.2.196:8100)"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  GenericResponse
// @Failure      500  {object}  GenericResponse
// @Router       /device/{udid}/accessibility/perform-action [post]
func performDtxAction(c *gin.Context) {
	axConnMux.Lock()
	defer axConnMux.Unlock()

	if !isAXEnabled || axConn == nil {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "Accessibility service not enabled"})
		return
	}

	// Get WDA host from query parameter
	wdaHost := c.Query("wda_host")
	log.Infof("performDtxAction called with wda_host: %q", wdaHost)

	// Only set WDA host if explicitly provided
	if wdaHost != "" {
		axConn.SetWDAHost(wdaHost)
	}
	// Note: We don't clear the WDA host if not provided - we keep the existing value

	err := axConn.PerformAction("AXAction-2010")
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, map[string]string{"message": "Action performed"})
}

// performWDAAction performs an accessibility action via WebDriverAgent
// @Summary      Perform accessibility action via WDA
// @Description  Performs an action using WebDriverAgent and returns action UUID
// @Tags         accessibility
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  GenericResponse
// @Failure      500  {object}  GenericResponse
// @Router       /device/{udid}/accessibility/wda/perform-action [post]
func performWDAAction(c *gin.Context) {
	axConnMux.RLock()
	defer axConnMux.RUnlock()

	if !isAXEnabled || axConn == nil {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "Accessibility service not enabled"})
		return
	}

	log.Infof("performWDAAction: Current WDA host is: %q", axConn.GetWDAHost())
	uuid, err := axConn.PerformWDAAction()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"uuid": uuid})
}

// getWDAHostStatus returns the current WDA host configuration
// @Summary      Get WDA host status
// @Description  Returns the current WDA host configuration for alert detection
// @Tags         accessibility
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  GenericResponse
// @Router       /device/{udid}/accessibility/wda/status [get]
func getWDAHostStatus(c *gin.Context) {
	axConnMux.RLock()
	defer axConnMux.RUnlock()

	if !isAXEnabled || axConn == nil {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "Accessibility service not enabled"})
		return
	}

	wdaHost := axConn.GetWDAHost()
	c.JSON(http.StatusOK, map[string]interface{}{
		"wda_host": wdaHost,
		"enabled":  wdaHost != "",
	})
}

// setElementChangeTimeout sets the timeout for waiting for element changes
// @Summary      Set element change timeout
// @Description  Sets the timeout for waiting for element changes in accessibility navigation
// @Tags         accessibility
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Param        timeout query int false "Timeout in seconds (default: 10)"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  GenericResponse
// @Router       /device/{udid}/accessibility/timeout [post]
func setElementChangeTimeout(c *gin.Context) {
	axConnMux.Lock()
	defer axConnMux.Unlock()

	if !isAXEnabled || axConn == nil {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "Accessibility service not enabled"})
		return
	}

	timeoutStr := c.Query("timeout")
	if timeoutStr == "" {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "timeout parameter is required"})
		return
	}

	timeoutSeconds, err := strconv.Atoi(timeoutStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "invalid timeout value, must be a number"})
		return
	}

	if timeoutSeconds <= 0 {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "timeout must be greater than 0"})
		return
	}

	timeout := time.Duration(timeoutSeconds) * time.Second
	axConn.SetElementChangeTimeout(timeout)

	c.JSON(http.StatusOK, map[string]string{
		"message": fmt.Sprintf("Element change timeout set to %d seconds", timeoutSeconds),
	})
}

// runAccessibilityAudit triggers an accessibility audit and returns found issues
// @Summary      Run accessibility audit
// @Description  Runs an accessibility audit for all supported types (fetched internally).
// @Tags         accessibility
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Success      200  {object}  RunAuditResponse
// @Failure      400  {object}  GenericResponse
// @Failure      500  {object}  GenericResponse
// @Router       /device/{udid}/accessibility/audit/run [post]
func runAccessibilityAudit(c *gin.Context) {
	axConnMux.RLock()
	defer axConnMux.RUnlock()

	if !isAXEnabled || axConn == nil {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "Accessibility service not enabled"})
		return
	}

	issues, err := axConn.RunAudit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, RunAuditResponse{Issues: issues})
}

// RunAuditResponse is the response body for audit run
// swagger:model
// nolint: revive
// RunAuditResponse wraps the list of issues
// Each issue is returned as provided by the accessibility package.
type RunAuditResponse struct {
	Issues []accessibility.AXAuditIssueV1 `json:"issues"`
}

// Optional: Add cleanup function for graceful shutdown
func cleanupAXService() {
	axConnMux.Lock()
	defer axConnMux.Unlock()

	if axConn != nil {
		axConn.TurnOff()
		axConn = nil
		isAXEnabled = false
	}
}
