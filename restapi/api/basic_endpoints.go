package api

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

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

func List(c *gin.Context) {
	// We are streaming current time to clients in the interval 10 seconds
	log.Info("connect")
	a, _, _ := ios.Listen()
	list, _ := a()
	c.IndentedJSON(http.StatusOK, list)
}
