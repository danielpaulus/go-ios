package api

import (
	"encoding/base64"
	"net/http"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/mcinstall"
	"github.com/gin-gonic/gin"
)

// registerMdmRoutes registers MDM/supervision endpoints (CLI: ios mdm ...). Every
// endpoint needs a supervisor identity: send multipart/form-data with a "p12"
// file and optional "password" field. Credentials are held only in memory for
// the duration of the request and never logged or written to disk.
func registerMdmRoutes(device *gin.RouterGroup) {
	mdm := device.Group("/mdm")
	mdm.POST("/security-info", MdmSecurityInfo)
	mdm.POST("/fetch-unlock-token", MdmFetchUnlockToken)
	mdm.POST("/clear-passcode", MdmClearPasscode)
	mdm.POST("/clear-screen-time-password", MdmClearScreenTimePassword)
}

// escalatedConn opens an mcinstall connection and escalates it to a supervised
// session using the multipart "p12"/"password" fields. The caller must Close it.
func escalatedConn(c *gin.Context) (*mcinstall.Connection, error) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	p12bytes, err := readFormFile(c, "p12")
	if err != nil {
		return nil, errMissingP12
	}
	conn, err := mcinstall.New(device)
	if err != nil {
		return nil, err
	}
	if err := conn.Escalate(p12bytes, c.PostForm("password")); err != nil {
		conn.Close()
		return nil, err
	}
	return conn, nil
}

// MdmSecurityInfo returns device security info (CLI: ios mdm security-info).
// @Summary Get MDM security info (supervised)
// @Accept multipart/form-data
// @Param udid path string true "Device UDID"
// @Param p12 formData file true "p12 supervisor identity"
// @Param password formData string false "p12 password"
// @Success 200 {object} interface{}
// @Router /device/{udid}/mdm/security-info [post]
func MdmSecurityInfo(c *gin.Context) {
	conn, err := escalatedConn(c)
	if err != nil {
		RespondError(c, statusForEscalateErr(err), err)
		return
	}
	defer conn.Close()
	info, err := conn.SecurityInfo()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, info)
}

// MdmFetchUnlockToken returns the escrow unlock token, base64-encoded (CLI: ios
// mdm fetch-unlock-token).
// @Summary Fetch the escrow unlock token (supervised)
// @Accept multipart/form-data
// @Param udid path string true "Device UDID"
// @Param p12 formData file true "p12 supervisor identity"
// @Param password formData string false "p12 password"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/mdm/fetch-unlock-token [post]
func MdmFetchUnlockToken(c *gin.Context) {
	conn, err := escalatedConn(c)
	if err != nil {
		RespondError(c, statusForEscalateErr(err), err)
		return
	}
	defer conn.Close()
	token, err := conn.FetchUnlockToken()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": base64.StdEncoding.EncodeToString(token)})
}

// MdmClearPasscode clears the device passcode (CLI: ios mdm clear-passcode). In
// addition to the p12, supply the unlock token as a base64 "token" form field.
// @Summary Clear the device passcode (supervised)
// @Accept multipart/form-data
// @Param udid path string true "Device UDID"
// @Param p12 formData file true "p12 supervisor identity"
// @Param password formData string false "p12 password"
// @Param token formData string true "base64 unlock token"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/mdm/clear-passcode [post]
func MdmClearPasscode(c *gin.Context) {
	tokenB64 := c.PostForm("token")
	if tokenB64 == "" {
		RespondError(c, http.StatusBadRequest, errMissingToken)
		return
	}
	tokenBytes, err := base64.StdEncoding.DecodeString(tokenB64)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	conn, err := escalatedConn(c)
	if err != nil {
		RespondError(c, statusForEscalateErr(err), err)
		return
	}
	defer conn.Close()
	if err := conn.ClearPasscode(tokenBytes); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// MdmClearScreenTimePassword clears the Screen Time password (CLI: ios mdm
// clear-screen-time-password).
// @Summary Clear the Screen Time password (supervised)
// @Accept multipart/form-data
// @Param udid path string true "Device UDID"
// @Param p12 formData file true "p12 supervisor identity"
// @Param password formData string false "p12 password"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/mdm/clear-screen-time-password [post]
func MdmClearScreenTimePassword(c *gin.Context) {
	conn, err := escalatedConn(c)
	if err != nil {
		RespondError(c, statusForEscalateErr(err), err)
		return
	}
	defer conn.Close()
	if err := conn.ClearScreenTimePassword(); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func statusForEscalateErr(err error) int {
	if err == errMissingP12 {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}
