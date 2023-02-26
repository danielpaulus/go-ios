package api

import (
	"net/http"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/gin-gonic/gin"
)

// List get device list of currently connected devices.
// List                godoc
// @Summary      Get device list
// @Description get device list of currently connected devices.
// @Tags         general
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /list [get]
func List(c *gin.Context) {
	list, err := ios.ListDevices()
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed getting device list with error", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, list)
}
