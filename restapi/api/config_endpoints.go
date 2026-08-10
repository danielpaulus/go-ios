package api

import (
	"encoding/hex"
	"net/http"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
	"github.com/danielpaulus/go-ios/ios/mcinstall"
	"github.com/gin-gonic/gin"
)

// registerConfigRoutes registers profile and developer-image management
// endpoints. GET /profiles and GET/PUT /image already exist in
// device_endpoints.go; this only adds the missing verbs. Routes are under
// /device/:udid.
func registerConfigRoutes(device *gin.RouterGroup) {
	device.POST("/profiles", AddProfile)
	device.DELETE("/profiles/:name", RemoveProfile)
	device.GET("/image/list", ListMountedImages)
	device.DELETE("/image", UnmountImage)
}

// AddProfile installs a configuration profile (CLI: ios profile add). Send the
// profile as the raw request body, or as multipart with a "profile" file plus an
// optional "p12" supervisor identity and "password" field (supervised install).
// @Summary Install a configuration profile
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /device/{udid}/profiles [post]
func AddProfile(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	var profileBytes, p12Bytes []byte
	var password string
	if profile, err := readFormFile(c, "profile"); err == nil {
		profileBytes = profile
		if p12, err := readFormFile(c, "p12"); err == nil {
			p12Bytes = p12
		}
		password = c.PostForm("password")
	} else {
		body, err := readAllLimited(c.Request.Body)
		if err != nil {
			RespondError(c, http.StatusBadRequest, err)
			return
		}
		profileBytes = body
	}
	if len(profileBytes) == 0 {
		RespondError(c, http.StatusBadRequest, errMissingProfile)
		return
	}

	conn, err := mcinstall.New(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()

	if len(p12Bytes) > 0 {
		err = conn.AddProfileSupervised(profileBytes, p12Bytes, password)
	} else {
		err = conn.AddProfile(profileBytes)
	}
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile installed"})
}

// RemoveProfile removes a configuration profile by identifier (CLI: ios profile remove).
// @Summary Remove a configuration profile
// @Param udid path string true "Device UDID"
// @Param name path string true "profile identifier"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/profiles/{name} [delete]
func RemoveProfile(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	conn, err := mcinstall.New(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()
	if err := conn.RemoveProfile(c.Param("name")); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "profile removed"})
}

// ListMountedImages lists mounted developer image signatures (CLI: ios image list).
// @Summary List mounted developer image signatures
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]interface{}
// @Router /device/{udid}/image/list [get]
func ListMountedImages(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	conn, err := imagemounter.NewImageMounter(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()
	signatures, err := conn.ListImages()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	hexSigs := make([]string, 0, len(signatures))
	for _, s := range signatures {
		hexSigs = append(hexSigs, hex.EncodeToString(s))
	}
	c.JSON(http.StatusOK, gin.H{"signatures": hexSigs, "count": len(hexSigs)})
}

// UnmountImage unmounts the developer disk image (CLI: ios image unmount).
// @Summary Unmount the developer disk image
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/image [delete]
func UnmountImage(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	if err := imagemounter.UnmountImage(device); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "image unmounted"})
}
