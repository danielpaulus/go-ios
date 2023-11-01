package api

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/syslog"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// Notifications uses instruments to get application state change events. It will stream the events as json objects separated by line breaks until it errors out.
// Listen                godoc
// @Summary      uses instruments to get application state change events
// @Description uses instruments to get application state change events
// @Tags         general
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /notifications [get]
func Notifications(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	listenerFunc, closeFunc, err := instruments.ListenAppStateNotifications(device)
	if err != nil {
		log.Fatal(err)
	}
	c.Stream(func(w io.Writer) bool {

		notification, err := listenerFunc()
		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			closeFunc()
			return false
		}

		_, err = w.Write([]byte(MustMarshal(notification)))

		if err != nil {
			c.JSON(http.StatusInternalServerError, err)
			closeFunc()
			return false
		}
		w.Write([]byte("\n"))
		return true
	})

}

// Syslog
// Listen                godoc
// @Summary      Uses SSE to connect to the LISTEN command
// @Description Uses SSE to connect to the LISTEN command
// @Tags         general
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /listen [get]
func Syslog(c *gin.Context) {
	// We are streaming current time to clients in the interval 10 seconds
	log.Info("connect")
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	syslogConnection, err := syslog.New(device)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.Stream(func(w io.Writer) bool {
		m, _ := syslogConnection.ReadLogMessage()
		// Stream message to client from message channel
		w.Write([]byte(MustMarshal(m)))
		return true
	})
}

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
