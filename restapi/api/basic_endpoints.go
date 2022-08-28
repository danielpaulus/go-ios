package api

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// Listen send server side events when devices are plugged in or removed
// Listen                godoc
// @Summary      Uses SSE to connect to the LISTEN command
// @Description Uses SSE to connect to the LISTEN command
// @Tags         general
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /listen [get]
func Listen(c *gin.Context) {
	// We are streaming current time to clients in the interval 10 seconds
	log.Info("connect")
	a, _, _ := ios.Listen()
	c.Stream(func(w io.Writer) bool {
		l, _ := a()
		// Stream message to client from message channel
		w.Write([]byte(MustMarshal(l)))
		return true
	})

}

// List grab device list
// List                godoc
// @Summary      Get device list
// @Description Get device list of currently attached devices
// @Tags         general
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /list [get]
func List(c *gin.Context) {
	// We are streaming current time to clients in the interval 10 seconds

	list, err := ios.ListDevices()
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Failed getting device list with error", "error": err.Error()})
		return
	}
	c.IndentedJSON(http.StatusOK, list)
}
