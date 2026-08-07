package api

import (
	"io"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/gin-gonic/gin"
)

// sysmontapSamplingRate matches Xcode's default sysmontap sampling rate.
const sysmontapSamplingRate = 10

// registerMonitoringRoutes registers streaming monitoring endpoints.
//
// Note: `ios pcap` is intentionally not exposed yet — the pcap package's Start
// writes packets to a local .pcap file and blocks, with no streaming API to hook
// a response writer to. Exposing it cleanly needs a packet-callback API in
// ios/pcap first; tracked as follow-up.
func registerMonitoringRoutes(device *gin.RouterGroup) {
	device.GET("/sysmontap", streamingMiddleWare, Sysmontap)
}

// Sysmontap streams CPU usage samples (CLI: ios sysmontap). Each line of the
// response body is a JSON CPU-usage sample; the stream ends when the client
// disconnects or the device closes the channel.
// @Summary Stream CPU usage samples
// @Produce application/json
// @Param udid path string true "Device UDID"
// @Success 200 {object} interface{}
// @Router /device/{udid}/sysmontap [get]
func Sysmontap(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	sysmon, err := instruments.NewSysmontapService(device, sysmontapSamplingRate)
	if err != nil {
		RespondError(c, 500, err)
		return
	}
	defer sysmon.Close()

	cpuUsageChannel := sysmon.ReceiveCPUUsage()
	c.Stream(func(w io.Writer) bool {
		msg, ok := <-cpuUsageChannel
		if !ok {
			return false
		}
		w.Write([]byte(MustMarshal(msg)))
		w.Write([]byte("\n"))
		return true
	})
}
