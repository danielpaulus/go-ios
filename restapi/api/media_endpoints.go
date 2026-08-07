package api

import (
	"io"
	"net/http"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/mcinstall"
	"github.com/danielpaulus/go-ios/ios/pasteboard"
	"github.com/danielpaulus/go-ios/ios/springboard"
	"github.com/gin-gonic/gin"
)

// registerMediaRoutes registers wallpaper, icon-layout and pasteboard endpoints
// mirroring the corresponding `ios` CLI commands. All routes live under
// /device/:udid.
func registerMediaRoutes(device *gin.RouterGroup) {
	device.GET("/wallpaper", GetWallpaper)
	device.PUT("/wallpaper", SetWallpaper)
	device.GET("/icon-layout", GetIconLayout)
	device.PUT("/icon-layout", SetIconLayout)
	device.GET("/pasteboard", GetPasteboard)
	device.PUT("/pasteboard", SetPasteboard)
}

// GetWallpaper returns the home-screen wallpaper as PNG (CLI: ios get-wallpaper).
// @Summary Get the home-screen wallpaper
// @Produce image/png
// @Param udid path string true "Device UDID"
// @Success 200 {file} binary
// @Router /device/{udid}/wallpaper [get]
func GetWallpaper(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	client, err := springboard.NewClient(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer client.Close()
	png, err := client.GetHomeScreenWallpaperPNG()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusOK, "image/png", png)
}

// SetWallpaper sets the wallpaper (CLI: ios set-wallpaper). This is a supervised
// operation: send multipart/form-data with an "image" file, a "p12" supervisor
// identity file, and optional "password" and "screen" fields.
// @Summary Set the wallpaper (supervised)
// @Accept multipart/form-data
// @Param udid path string true "Device UDID"
// @Param image formData file true "image file"
// @Param p12 formData file true "p12 supervisor identity"
// @Param password formData string false "p12 password"
// @Param screen formData string false "target screen"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Router /device/{udid}/wallpaper [put]
func SetWallpaper(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	imageBytes, err := readFormFile(c, "image")
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	p12bytes, err := readFormFile(c, "p12")
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	screen, err := mcinstall.ParseWallpaperScreen(c.PostForm("screen"))
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	conn, err := mcinstall.New(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()
	if err := conn.SetWallpaperSupervised(imageBytes, screen, p12bytes, c.PostForm("password")); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "wallpaper set"})
}

// GetIconLayout returns the home-screen icon layout (CLI: ios get-icon-layout).
// @Summary Get the icon layout
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} interface{}
// @Router /device/{udid}/icon-layout [get]
func GetIconLayout(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	client, err := springboard.NewClient(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer client.Close()
	state, err := client.GetIconLayout("2")
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, state)
}

// SetIconLayout restores an icon layout (CLI: ios set-icon-layout). Body is the
// layout JSON as returned by GET.
// @Summary Set the icon layout
// @Accept json
// @Param udid path string true "Device UDID"
// @Param body body interface{} true "icon layout JSON"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/icon-layout [put]
func SetIconLayout(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	var state any
	if err := c.ShouldBindJSON(&state); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	client, err := springboard.NewClient(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer client.Close()
	if err := client.SetIconLayout(state); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "icon layout set"})
}

// GetPasteboard returns the device clipboard text (CLI: ios pasteboard get).
// @Summary Get the pasteboard text
// @Produce json
// @Param udid path string true "Device UDID"
// @Success 200 {object} map[string]interface{}
// @Router /device/{udid}/pasteboard [get]
func GetPasteboard(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	conn, err := pasteboard.New(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()
	text, ok, err := conn.GetText()
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"present": ok, "text": text})
}

// SetPasteboard sets the device clipboard from the raw request body (CLI: ios
// pasteboard set <text>).
// @Summary Set the pasteboard text
// @Accept text/plain
// @Param udid path string true "Device UDID"
// @Param body body string true "clipboard text"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/pasteboard [put]
func SetPasteboard(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	conn, err := pasteboard.New(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()
	if err := conn.SetText(string(body)); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "pasteboard set"})
}

// readFormFile reads an entire multipart form file field into memory.
func readFormFile(c *gin.Context, field string) ([]byte, error) {
	f, _, err := c.Request.FormFile(field)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}
