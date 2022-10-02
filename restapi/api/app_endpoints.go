package api

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/gin-gonic/gin"
	"net/http"
)

func initAppRoutes(group *gin.RouterGroup) {
	router := group.Group("/app")
	router.GET("/", ListApps)
}

func ListApps(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	svc, _ := installationproxy.New(device)
	var err error
	var response []installationproxy.AppInfo
	response, err = svc.BrowseAllApps()
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
	}
	c.IndentedJSON(http.StatusOK, response)
}
