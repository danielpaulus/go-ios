package api

import (
	"log"
	"net/http"
	"os"
	"path"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/zipconduit"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	_, err = pControl.LaunchApp(bundleID, nil)
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
	processName := ""

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

// Install app on a device
// @Summary      Install app on a device
// @Description  Install app on a device by uploading an ipa file
// @Tags         apps
// @Produce      json
// @Param        file formData file true "ipa file to install"
// @Success      200 {object} GenericResponse
// @Failure      500 {object} GenericResponse
// @Router       /device/{udid}/apps/install [post]
func InstallApp(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	file, err := c.FormFile("file")

	log.Printf("Received file: %s", file.Filename)

	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, GenericResponse{Error: "file form-data is missing"})
		return
	}

	if file.Size == 0 { // 100 MB limit
		c.JSON(http.StatusRequestEntityTooLarge, GenericResponse{Error: "uploaded file is empty"})
		return
	}

	if file.Size > 200*1024*1024 { // 100 MB limit
		c.JSON(http.StatusRequestEntityTooLarge, GenericResponse{Error: "file size exceeds the 200MB limit"})
		return
	}

	appDownloadFolder := os.Getenv("APP_DOWNLOAD_FOLDER")
	if appDownloadFolder == "" {
		appDownloadFolder = os.TempDir()
	}

	dst := path.Join(appDownloadFolder, uuid.New().String()+".ipa")
	defer func() {
		if err := os.Remove(dst); err != nil {
			c.JSON(http.StatusInternalServerError, GenericResponse{Error: "failed to delete temporary file"})
		}
	}()

	c.SaveUploadedFile(file, dst)

	conn, err := zipconduit.New(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: "Unable to setup ZipConduit connection"})
		return
	}

	err = conn.SendFile(dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: "Unable to install uploaded app"})
		return
	}

	c.JSON(http.StatusOK, GenericResponse{Message: "App installed successfully"})
}

// Uninstall app on a device
// @Summary      Uninstall app on a device
// @Description  Uninstall app on a device by provided bundleID
// @Tags         apps
// @Produce      json
// @Param        bundleID query string true "bundle identifier of the targeted app"
// @Success      200 {object} GenericResponse
// @Failure      500 {object} GenericResponse
func UninstallApp(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	bundleID := c.Query("bundleID")
	if bundleID == "" {
		c.JSON(http.StatusUnprocessableEntity, GenericResponse{Error: "bundleID query param is missing"})
		return
	}

	svc, err := installationproxy.New(device)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}
	defer svc.Close()

	err = svc.Uninstall(bundleID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, GenericResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, GenericResponse{Message: bundleID + " uninstalled successfully"})
}
