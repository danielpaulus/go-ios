package api

import (
	"fmt"
	"io"
	"net/http"

	"github.com/danielpaulus/go-ios/agent/models"
	"github.com/danielpaulus/go-ios/agent/orchestratorclient"
	"github.com/danielpaulus/go-ios/agent/wrtc"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/syslog"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func ExecuteCommand(c *gin.Context) {
	var args map[string]interface{}
	err := c.BindJSON(&args)
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(400, gin.H{"message": "Failed parsing request", "error": err.Error()})
		return
	}

	if isSet(args, "list") {
		info, err := orchestratorclient.GetDevices()
		if err != nil {
			c.IndentedJSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}
		list := createDeviceList(info)
		resp := map[string]interface{}{"deviceList": list}
		c.IndentedJSON(http.StatusOK, resp)
		return
	}
	udid := getString(args, "--udid")
	device, isLocal, err := getDevice(udid)
	if err != nil {
		c.IndentedJSON(http.StatusOK, gin.H{"uknown device": err.Error()})
		return
	}
	log.Infof("device: %v, isLocal: %v, err: %v", device, isLocal, err)
	if isSet(args, "syslog") {
		if isLocal {
			log.Info("running local syslog")
			syslogConnection, err := syslog.New(device)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
				return
			}
			log.Info("starting logstream")
			c.Stream(func(w io.Writer) bool {
				m, _ := syslogConnection.ReadLogMessage()
				log.Info(m)
				// Stream message to client from message channel
				w.Write([]byte(MustMarshal(m)))
				return true
			})
			return
		}
		log.Info("running remote syslog")
		conn, err := wrtc.Connect(device)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		reader, err := conn.StreamingResponse("syslog")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		buf := make([]byte, 1024)
		c.Stream(func(w io.Writer) bool {
			n, err := reader.Read(buf)
			if err != nil {
				log.Errorf("error reading from syslog: %v", err)
				return false
			}
			// Stream message to client from message channel
			w.Write(buf[:n])
			return true
		})
	}
}

func getDevice(udid string) (ios.DeviceEntry, bool, error) {
	devices, err := ios.ListDevices()
	if err != nil {
		return ios.DeviceEntry{}, false, err
	}
	for _, device := range devices.DeviceList {
		if device.Properties.SerialNumber == udid {
			return device, true, nil
		}
	}
	if udid == "" && len(devices.DeviceList) > 0 {
		return devices.DeviceList[0], true, nil
	}
	info, err := orchestratorclient.GetDevices()
	if err != nil {
		return ios.DeviceEntry{}, false, fmt.Errorf("getDevice: failed getting cloud devices %w", err)
	}
	for _, device := range info {
		if device.Serial == udid {
			return ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: udid}}, false, nil
		}
	}
	if udid == "" && len(info) > 0 {
		return ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: info[0].Serial}}, false, nil
	}
	return ios.DeviceEntry{}, false, fmt.Errorf("getDevice: device with udid %s not found", udid)
}

func getString(args map[string]interface{}, key string) string {
	v, ok := args[key]
	if !ok {
		return ""
	}
	if v == nil {
		return ""
	}
	val, ok := v.(string)
	if !ok {
		return ""
	}
	return val
}

func isSet(args map[string]interface{}, key string) bool {
	v, ok := args[key]
	if !ok {
		return false
	}
	if v == nil {
		return false
	}
	val, ok := v.(bool)
	if !ok {
		return false
	}
	return val
}
func createDeviceList(info []models.DeviceInfo) []string {
	var list []string
	for _, device := range info {
		list = append(list, device.Serial)
	}
	return list
}
