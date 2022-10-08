package api

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
	"github.com/gin-gonic/gin"
	"net/http"
)

/*
APIs needed to solve automation problem:
1. app install
2. dev image mount and check
3. run wda
4. wda shim/ tap and screenshot
5. signing api
6. wda binary download

*/

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
