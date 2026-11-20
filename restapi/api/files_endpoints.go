package api

import (
	"net/http"
	"path"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/crashreport"
	"github.com/danielpaulus/go-ios/ios/fileservice"
	"github.com/gin-gonic/gin"
)

// registerFilesRoutes registers file-transfer and crash-report endpoints
// mirroring `ios file` and `ios crash`. All routes live under /device/:udid.
//
// File transfer uses the iOS 17+ file service and streams directly to/from the
// HTTP body, so there is no caller-supplied host path and thus no host-side path
// traversal: pull writes to the response, push reads from the request body.
func registerFilesRoutes(device *gin.RouterGroup) {
	device.GET("/files", ListFiles)
	device.GET("/files/pull", PullFile)
	device.POST("/files/push", PushFile)
	device.GET("/crashes", ListCrashes)
	device.DELETE("/crashes", RemoveCrashes)
}

// fileConnFromQuery opens a file-service connection from the request's
// domain/identifier query params. domain is one of app|app-group|crash|temp.
func fileConnFromQuery(c *gin.Context) (*fileservice.Connection, error) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	identifier := c.Query("identifier")
	var domain fileservice.Domain
	switch c.Query("domain") {
	case "app":
		domain = fileservice.DomainAppDataContainer
	case "app-group":
		domain = fileservice.DomainAppGroupDataContainer
	case "crash":
		domain = fileservice.DomainSystemCrashLogs
	case "temp":
		domain = fileservice.DomainTemporary
	case "":
		return nil, errMissingDomain
	default:
		return nil, errUnknownDomain
	}
	return fileservice.New(device, domain, identifier)
}

// ListFiles lists a directory (CLI: ios file ls).
// @Summary List files in a device directory
// @Produce json
// @Param udid path string true "Device UDID"
// @Param domain query string true "app|app-group|crash|temp"
// @Param identifier query string false "bundle/group id for app/app-group domains"
// @Param path query string false "directory path (default '.')"
// @Success 200 {object} map[string]interface{}
// @Router /device/{udid}/files [get]
func ListFiles(c *gin.Context) {
	conn, err := fileConnFromQuery(c)
	if err != nil {
		RespondError(c, statusForFileErr(err), err)
		return
	}
	defer conn.Close()
	p := c.Query("path")
	if p == "" {
		p = "."
	}
	files, err := conn.ListDirectory(p)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"path": p, "files": files, "count": len(files)})
}

// PullFile streams a device file back as the response body (CLI: ios file pull).
// @Summary Download a file from the device
// @Produce application/octet-stream
// @Param udid path string true "Device UDID"
// @Param domain query string true "app|app-group|crash|temp"
// @Param identifier query string false "bundle/group id for app/app-group domains"
// @Param remote query string true "remote file path on the device"
// @Success 200 {file} binary
// @Router /device/{udid}/files/pull [get]
func PullFile(c *gin.Context) {
	remote := c.Query("remote")
	if remote == "" {
		RespondError(c, http.StatusBadRequest, errMissingRemote)
		return
	}
	conn, err := fileConnFromQuery(c)
	if err != nil {
		RespondError(c, statusForFileErr(err), err)
		return
	}
	defer conn.Close()
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=\""+path.Base(remote)+"\"")
	if err := conn.PullFile(remote, c.Writer); err != nil {
		// Headers may already be sent; best effort to signal failure otherwise.
		if !c.Writer.Written() {
			RespondError(c, http.StatusInternalServerError, err)
		}
		return
	}
}

// PushFile uploads the request body to a device path (CLI: ios file push).
// @Summary Upload a file to the device
// @Accept application/octet-stream
// @Param udid path string true "Device UDID"
// @Param domain query string true "app|app-group|crash|temp"
// @Param identifier query string false "bundle/group id for app/app-group domains"
// @Param remote query string true "destination path on the device"
// @Success 200 {object} map[string]interface{}
// @Router /device/{udid}/files/push [post]
func PushFile(c *gin.Context) {
	remote := c.Query("remote")
	if remote == "" {
		RespondError(c, http.StatusBadRequest, errMissingRemote)
		return
	}
	size := c.Request.ContentLength
	if size < 0 {
		RespondError(c, http.StatusLengthRequired, errUnknownContentLength)
		return
	}
	conn, err := fileConnFromQuery(c)
	if err != nil {
		RespondError(c, statusForFileErr(err), err)
		return
	}
	defer conn.Close()
	// Match the CLI's push defaults (0644, uid/gid 501).
	if err := conn.PushFile(remote, c.Request.Body, size, 0o644, 501, 501); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"remote": remote, "size": size})
}

// ListCrashes lists crash reports (CLI: ios crash ls).
// @Summary List crash reports
// @Produce json
// @Param udid path string true "Device UDID"
// @Param pattern query string false "glob pattern"
// @Success 200 {object} map[string]interface{}
// @Router /device/{udid}/crashes [get]
func ListCrashes(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	files, err := crashreport.ListReports(device, c.Query("pattern"))
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"files": files, "count": len(files)})
}

// RemoveCrashes deletes crash reports (CLI: ios crash rm).
// @Summary Delete crash reports
// @Param udid path string true "Device UDID"
// @Param cwd query string true "working directory on the device"
// @Param pattern query string true "glob pattern"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/crashes [delete]
func RemoveCrashes(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	cwd := c.Query("cwd")
	pattern := c.Query("pattern")
	if cwd == "" || pattern == "" {
		RespondError(c, http.StatusBadRequest, errMissingCrashArgs)
		return
	}
	if err := crashreport.RemoveReports(device, cwd, pattern); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "removed", "pattern": pattern})
}

func statusForFileErr(err error) int {
	switch err {
	case errMissingDomain, errUnknownDomain:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
