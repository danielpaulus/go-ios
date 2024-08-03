package api

import (
	"net/http"

	"github.com/danielpaulus/go-ios/agent/models"
	"github.com/danielpaulus/go-ios/agent/orchestratorclient"
	"github.com/gin-gonic/gin"
	"gvisor.dev/gvisor/pkg/log"
)

func ExecuteCommand(c *gin.Context) {
	var args []string
	err := c.BindJSON(&args)
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(400, gin.H{"message": "Failed parsing request", "error": err.Error()})
		return
	}
	log.Infof("Received command %v", args)
	if args[1] == "list" {
		info, err := orchestratorclient.GetDevices()
		if err != nil {
			c.IndentedJSON(http.StatusOK, gin.H{"error": err.Error()})
		}
		list := createDeviceList(info)
		resp := map[string]interface{}{"deviceList": list}
		c.IndentedJSON(http.StatusOK, resp)
	}
}
func createDeviceList(info []models.DeviceInfo) []string {
	var list []string
	for _, device := range info {
		list = append(list, device.Serial)
	}
	return list
}
