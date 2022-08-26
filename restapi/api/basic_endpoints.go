package api

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/screenshotr"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
)

// GetBookByISBN locates the book whose ISBN value matches the isbn
// GetBookByISBN                godoc
// @Summary      Get single book by isbn
// @Description  Returns the book whose ISBN value matches the isbn.
// @Tags         books
// @Produce      json
// @Param        isbn  path      string  true  "search book by isbn"
// @Success      200  {object}  map[string]interface{}
// @Router       /books/{isbn} [get]
func Info(c *gin.Context) {

	device, _ := ios.GetDevice("")

	allValues, err := ios.GetValuesPlist(device)
	if err != nil {
		print(err)
	}
	svc, err := instruments.NewDeviceInfoService(device)
	if err != nil {
		log.Debugf("could not open instruments, probably dev image not mounted %v", err)
	}
	if err == nil {
		info, err := svc.NetworkInformation()
		if err != nil {
			log.Debugf("error getting networkinfo from instruments %v", err)
		} else {
			allValues["instruments:networkInformation"] = info
		}
		info, err = svc.HardwareInformation()
		if err != nil {
			log.Debugf("error getting hardwareinfo from instruments %v", err)
		} else {
			allValues["instruments:hardwareInformation"] = info
		}
	}
	c.IndentedJSON(http.StatusOK, allValues)
}

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

func Screenshot(c *gin.Context) {

	dev, _ := ios.GetDevice("")
	conn, err := screenshotr.New(dev)
	log.Error(err)
	b, _ := conn.TakeScreenshot()

	c.Header("Content-Type", "image/png")
	c.Data(http.StatusOK, "application/octet-stream", b)
}
