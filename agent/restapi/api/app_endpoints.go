package api

import (
	"net/http"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/gin-gonic/gin"
)

// List apps on a device
// @Summary      List apps on a device
// @Description  List the installed apps on a device
// @Tags         apps
// @Produce      json
// @Success      200 {object} []installationproxy.AppInfo
// @Failure      500 {object} GenericResponse
// @Router       /device/{udid}/apps [post]
func ListApps(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	svc, _ := installationproxy.New(device)
	var err error
	var response []installationproxy.AppInfo
	response, err = svc.BrowseAllApps()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
	}
	c.IndentedJSON(http.StatusOK, response)
}

// Launch app on a device
// @Summary      Launch app on a device
// @Description  Launch app on a device by provided bundleID
// @Tags         apps
// @Produce      json
// @Param        bundleID query string true "bundle identifier of the targeted app"
// @Success      200  {object} GenericResponse
// @Failure      500  {object} GenericResponse
// @Router       /device/{udid}/apps/launch [post]
func LaunchApp(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	bundleID := c.Query("bundleID")
	if bundleID == "" {
		c.JSON(http.StatusUnprocessableEntity, GenericResponse{Error: "bundleID query param is missing"})
		return
	}

	pControl, err := instruments.NewProcessControl(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	_, err = pControl.LaunchApp(bundleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, GenericResponse{Message: bundleID + " launched successfully"})
}

// Kill running app on a device
// @Summary      Kill running app on a device
// @Description  Kill running app on a device by provided bundleID
// @Tags         apps
// @Produce      json
// @Param        bundleID query string true "bundle identifier of the targeted app"
// @Success      200 {object} GenericResponse
// @Failure      500 {object} GenericResponse
// @Router       /device/{udid}/apps/kill [post]
func KillApp(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	var processName = ""

	bundleID := c.Query("bundleID")
	if bundleID == "" {
		c.JSON(http.StatusUnprocessableEntity, GenericResponse{Error: "bundleID query param is missing"})
		return
	}

	pControl, err := instruments.NewProcessControl(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	svc, err := installationproxy.New(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	response, err := svc.BrowseAllApps()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	for _, app := range response {
		if app.CFBundleIdentifier == bundleID {
			processName = app.CFBundleExecutable
			break
		}
	}

	if processName == "" {
		c.JSON(http.StatusNotFound, GenericResponse{Message: bundleID + " is not installed"})
		return
	}

	service, err := instruments.NewDeviceInfoService(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}
	defer service.Close()

	processList, err := service.ProcessList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	for _, p := range processList {
		if p.Name == processName {
			err = pControl.KillProcess(p.Pid)
			if err != nil {
				c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
				return
			}
			c.JSON(http.StatusOK, GenericResponse{Message: bundleID + " successfully killed"})
			return
		}
	}

	c.JSON(http.StatusOK, GenericResponse{Message: bundleID + " is not running"})
}
