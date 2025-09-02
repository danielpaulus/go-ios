package api

import (
	"net/http"
	"sync"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/accessibility"
	"github.com/gin-gonic/gin"
)

var (
	axConn      *accessibility.ControlInterface
	axConnMux   sync.RWMutex
	isAXEnabled bool
)

// enableAXService enables the accessibility service session for the device
// @Summary      Enable accessibility service
// @Description  Starts an accessibility session on the device and enables selection mode
// @Tags         accessibility
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Success      200  {object}  map[string]string
// @Failure      500  {object}  GenericResponse
// @Router       /device/{udid}/accessibility/enable [get]
func enableAXService(c *gin.Context) {
	axConnMux.Lock()
	defer axConnMux.Unlock()

	if isAXEnabled && axConn != nil {
		c.JSON(http.StatusOK, map[string]string{"message": "Accessibility service already enabled"})
		return
	}

	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	conn, err := accessibility.New(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	conn.SwitchToDevice()

	conn.EnableSelectionMode()

	axConn = &conn
	isAXEnabled = true
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

// performDtxAction performs an accessibility action via the DTX channel
// @Summary      Perform accessibility action via DTX
// @Description  Performs the default action on the current focused element using DTX
// @Tags         accessibility
// @Produce      json
// @Param        udid path string true "Device UDID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  GenericResponse
// @Failure      500  {object}  GenericResponse
// @Router       /device/{udid}/accessibility/perform-action [post]
func performDtxAction(c *gin.Context) {
	axConnMux.RLock()
	defer axConnMux.RUnlock()

	if !isAXEnabled || axConn == nil {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "Accessibility service not enabled"})
		return
	}

	// Construct and send the DTX message using the stored platform value
	// Example: axConn.PerformAction(currentPlatformElementValue, "Activate")
	// You'll need to implement PerformAction in your accessibility library

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

	host := "http://10.15.47.131:8100"
	if host == "" {
		c.JSON(http.StatusBadRequest, GenericResponse{Error: "missing X-WDA-Host header (e.g. http://<ip>:8100)"})
		return
	}

	uuid, err := axConn.PerformWDAAction(host)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"uuid": uuid})
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
