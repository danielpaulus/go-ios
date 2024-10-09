package api

import (
	"io"
	"net/http"
	"os"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

type WdaConfig struct {
	BundleID     string                 `json:"bundleId" binding:"required"`
	TestbundleID string                 `json:"testBundleId" binding:"required"`
	XCTestConfig string                 `json:"xcTestConfig" binding:"required"`
	Args         []string               `json:"args"`
	Env          map[string]interface{} `json:"env"`
}

type ChannelWriter struct {
	Channel chan []byte
}

func (w *ChannelWriter) Write(p []byte) (n int, err error) {
	log.Debugf("writing %s", string(p))
	w.Channel <- p
	return len(p), nil
}

func RunWda(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	log.WithField("udid", device.Properties.SerialNumber).Printf("Running WDA on device %t", device.UserspaceTUN)

	var config WdaConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logStream := make(chan []byte, 10)
	w := ChannelWriter{Channel: logStream}
	go func() {
		defer close(logStream)
		_, err := testmanagerd.RunXCUIWithBundleIdsCtx(c, config.BundleID, config.TestbundleID, config.XCTestConfig, device, config.Args, config.Env, nil, nil, testmanagerd.NewTestListener(&w, &w, os.TempDir()), false)
		if err != nil {
			log.WithError(err).Error("Failed running WDA")
			// return true
		}
	}()

	c.Stream(func(w io.Writer) bool {
		if msg, ok := <-logStream; ok {
			c.SSEvent("log", msg)
			return true
		}
		return false
	})
}
