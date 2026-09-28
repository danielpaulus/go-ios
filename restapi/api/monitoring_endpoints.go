package api

import (
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/gin-gonic/gin"
)

// CpuUsageSample is the payload of a `sample` event (SSE /sysmontap). It is an
// open map (sampler keys vary by OS); the commonly-present CPU_TotalLoad is
// surfaced explicitly to match the spec's CpuUsageSample schema.
type CpuUsageSample struct {
	CPU_TotalLoad float64 `json:"CPU_TotalLoad"`
}

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

// Sysmontap streams CPU-usage samples as Server-Sent Events (CLI: ios
// sysmontap). Each `sample` event carries a CpuUsageSample; a `heartbeat` event
// is emitted on idle. The stream ends when the client disconnects or the device
// closes the channel.
// @Summary Stream CPU usage samples (SSE)
// @Description Streams sysmontap CPU-usage samples as text/event-stream. Events: `sample` (CpuUsageSample), `heartbeat`.
// @Produce text/event-stream
// @Param udid path string true "Device UDID"
// @Success 200 {object} CpuUsageSample
// @Router /device/{udid}/sysmontap [get]
func Sysmontap(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	udid := device.Properties.SerialNumber
	sysmon, err := instruments.NewSysmontapService(device, sysmontapSamplingRate)
	if err != nil {
		RespondError(c, 500, err)
		return
	}
	defer sysmon.Close()

	golog.Info("sysmontap stream started", "module", logModule, "udid", udid)
	cpuUsageChannel := sysmon.ReceiveCPUUsage()
	streamSSE(c, udid, func() (sseEvent, bool) {
		msg, ok := <-cpuUsageChannel
		if !ok {
			return sseEvent{}, false
		}
		return sseEvent{event: "sample", payload: CpuUsageSample{CPU_TotalLoad: msg.SystemCPUUsage.CPU_TotalLoad}}, true
	})
}
