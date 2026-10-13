package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/pcap"
	"github.com/gin-gonic/gin"
)

// pcapContentType is the IANA media type for a libpcap capture file. Clients
// (wireshark, tcpdump, tshark -r -) recognize the stream by this type.
const pcapContentType = "application/vnd.tcpdump.pcap"

// pcapMaxTimeoutSeconds caps the caller-supplied ?timeout= so a single request
// cannot hold a device pcapd connection open indefinitely.
const pcapMaxTimeoutSeconds = 3600

// pcapDefaultTimeout is used when no ?timeout= is given. The capture then runs
// until this deadline or until the client disconnects, whichever comes first.
const pcapDefaultTimeout = 60 * time.Second

// flushWriter wraps a streaming http response writer so every write is pushed
// to the client immediately. pcap.Stream writes the global header and then one
// record per captured packet; flushing per write keeps the live capture from
// buffering on the server.
type flushWriter struct {
	w gin.ResponseWriter
}

func (f flushWriter) Write(p []byte) (int, error) {
	n, err := f.w.Write(p)
	if err != nil {
		return n, err
	}
	f.w.Flush()
	return n, nil
}

// Pcap streams a live packet capture from the device as a libpcap stream.
// The response begins with the pcap global header followed by one record per
// captured packet, so it can be piped straight into wireshark/tshark. The
// capture runs until the optional ?timeout= (seconds) elapses, the default
// timeout is reached, or the client disconnects.
// @Summary      Stream a live pcap capture
// @Description  Streams live network traffic from the device as a libpcap (application/vnd.tcpdump.pcap) byte stream until the timeout elapses or the client disconnects.
// @Tags         general
// @Param        timeout  query  int  false  "Capture duration in seconds (default 60, max 3600)"
// @Produce      application/vnd.tcpdump.pcap
// @Success      200  {string}  binary  "libpcap byte stream"
// @Router       /device/{udid}/pcap [get]
func Pcap(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	udid := device.Properties.SerialNumber

	timeout := pcapDefaultTimeout
	if q := c.Query("timeout"); q != "" {
		secs, err := strconv.Atoi(q)
		if err != nil || secs <= 0 {
			RespondError(c, http.StatusBadRequest, errInvalidTimeout)
			return
		}
		if secs > pcapMaxTimeoutSeconds {
			secs = pcapMaxTimeoutSeconds
		}
		timeout = time.Duration(secs) * time.Second
	}

	// Derive the capture lifetime from the request context (client disconnect)
	// and the timeout cap; whichever fires first stops the capture cleanly.
	ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
	defer cancel()

	c.Header("Content-Type", pcapContentType)
	c.Header("Transfer-Encoding", "chunked")
	c.Header("X-Content-Type-Options", "nosniff")

	golog.Info("pcap stream started", "module", logModule, "udid", udid, "timeoutSeconds", int(timeout.Seconds()))
	if err := pcap.Stream(ctx, device, flushWriter{w: c.Writer}); err != nil {
		// Once bytes are on the wire the status is already 200 and we can only
		// log; otherwise surface the failure to connect to pcapd.
		if c.Writer.Written() {
			golog.Error("pcap stream failed mid-capture", "module", logModule, "udid", udid, "error", err.Error())
			return
		}
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	golog.Info("pcap stream finished", "module", logModule, "udid", udid)
}

// registerPcapRoutes wires the live pcap streaming endpoint onto the
// device-scoped group.
func registerPcapRoutes(device *gin.RouterGroup) {
	device.GET("/pcap", Pcap)
}
