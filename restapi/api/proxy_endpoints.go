package api

import (
	"net/http"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/mcinstall"
	"github.com/gin-gonic/gin"
)

// registerProxyRoutes registers the global HTTP proxy endpoints (CLI: ios
// httpproxy). Setting a proxy is a supervised operation. Routes under /device/:udid.
func registerProxyRoutes(device *gin.RouterGroup) {
	device.PUT("/httpproxy", SetHTTPProxy)
	device.DELETE("/httpproxy", RemoveHTTPProxy)
}

// SetHTTPProxy configures a global HTTP proxy (CLI: ios httpproxy). Send
// multipart/form-data with "host" and "port" fields, a "p12" supervisor identity
// file, and optional "user", "pass" and "password" (p12 password) fields.
// @Summary Set a global HTTP proxy (supervised)
// @Accept multipart/form-data
// @Param udid path string true "Device UDID"
// @Param host formData string true "proxy host"
// @Param port formData string true "proxy port"
// @Param p12 formData file true "p12 supervisor identity"
// @Param user formData string false "proxy username"
// @Param pass formData string false "proxy password"
// @Param password formData string false "p12 password"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /device/{udid}/httpproxy [put]
func SetHTTPProxy(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	host := c.PostForm("host")
	port := c.PostForm("port")
	if host == "" || port == "" {
		RespondError(c, http.StatusBadRequest, errMissingProxyHostPort)
		return
	}
	p12, err := readFormFile(c, "p12")
	if err != nil {
		RespondError(c, http.StatusBadRequest, errMissingP12)
		return
	}
	if err := mcinstall.SetHttpProxy(device, host, port, c.PostForm("user"), c.PostForm("pass"), p12, c.PostForm("password")); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "http proxy set", "host": host, "port": port})
}

// RemoveHTTPProxy clears the global HTTP proxy (CLI: ios httpproxy remove).
// @Summary Remove the global HTTP proxy
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/httpproxy [delete]
func RemoveHTTPProxy(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	if err := mcinstall.RemoveProxy(device); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "http proxy removed"})
}
