package api

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/instruments"
	"github.com/danielpaulus/go-ios/ios/uidriver"
	"github.com/gin-gonic/gin"
)

// Streaming endpoints deliver long-lived video/image streams over HTTP:
//
//   - GET /device/:udid/ui/stream        proxies a WDA/DeviceKit UI video stream
//     (mjpeg or h264) straight through to the client.
//   - GET /device/:udid/screenshot/stream serves an MJPEG stream of device
//     screenshots captured via the instruments screenshot service.
//
// Both stream until the client disconnects (request context canceled) or the
// source ends, and flush frames as they arrive.

const (
	// mjpegStreamBoundary is the multipart boundary for the screenshot MJPEG
	// stream. It matches the format used by the instruments MJPEG server.
	mjpegStreamBoundary  = "BoundaryString"
	mjpegStreamPartHead  = "--" + mjpegStreamBoundary + "\r\nContent-Type: image/jpeg\r\nContent-Length: %d\r\n\r\n"
	mjpegStreamPartFoot  = "\r\n"
	mjpegStreamCType     = "multipart/x-mixed-replace; boundary=" + mjpegStreamBoundary
	streamCopyBufferSize = 32 * 1024
)

// registerStreamRoutes registers the streaming endpoints. The UI stream lives
// under the /ui group next to the other UI-automation endpoints; the screenshot
// stream is device-scoped.
func registerStreamRoutes(device *gin.RouterGroup) {
	device.GET("/ui/stream", UIStream)
	device.GET("/screenshot/stream", ScreenshotStream)
}

// UIStream opens a UI video stream against a forwarded WDA/DeviceKit backend and
// pipes it straight through to the client, preserving the backend's Content-Type
// (mjpeg is multipart/x-mixed-replace; h264 uses the backend's type). It streams
// until the client disconnects or the backend ends.
//
//	?codec=mjpeg|h264   (default mjpeg; h264 requires the devicekit backend)
//	?backend=wda|devicekit, ?wdaUrl=<url>, ?timeout=<seconds>  (see UI endpoints)
//
// Requires a running, forwarded WDA/DeviceKit backend; see registerUIRoutes.
func UIStream(c *gin.Context) {
	driver, udid, err := newDriverFromRequest(c)
	if err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}

	var opts uidriver.StreamOptions
	switch strings.ToLower(strings.TrimSpace(paramOrHeader(c, "codec"))) {
	case "", "mjpeg":
		opts.H264 = false
	case "h264":
		opts.H264 = true
	default:
		RespondError(c, http.StatusBadRequest, errors.New("unknown codec; expected mjpeg or h264"))
		return
	}
	opts.FPS = strings.TrimSpace(paramOrHeader(c, "fps"))
	opts.Quality = strings.TrimSpace(paramOrHeader(c, "quality"))
	opts.Scale = strings.TrimSpace(paramOrHeader(c, "scale"))
	opts.Bitrate = strings.TrimSpace(paramOrHeader(c, "bitrate"))

	ctx := c.Request.Context()
	body, contentType, err := driver.StreamWithContentType(ctx, opts)
	if err != nil {
		respondUIError(c, udid, err)
		return
	}
	defer func() { _ = body.Close() }()

	if contentType == "" {
		if opts.H264 {
			contentType = "video/H264"
		} else {
			contentType = mjpegStreamCType
		}
	}
	c.Writer.Header().Set("Content-Type", contentType)
	c.Writer.Header().Set("Cache-Control", "no-cache, private")
	c.Writer.Header().Set("Connection", "close")
	c.Writer.WriteHeader(http.StatusOK)

	golog.Info("streaming ui video to client", "module", logModule, "udid", udid, "codec", streamCodec(opts.H264), "contentType", contentType)

	flusher, _ := c.Writer.(http.Flusher)
	buf := make([]byte, streamCopyBufferSize)
	for {
		if ctx.Err() != nil {
			return
		}
		n, readErr := body.Read(buf)
		if n > 0 {
			if _, writeErr := c.Writer.Write(buf[:n]); writeErr != nil {
				golog.Debug("ui stream client write failed, closing", "module", logModule, "udid", udid, "error", writeErr.Error())
				return
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
		if readErr != nil {
			if !errors.Is(readErr, io.EOF) && ctx.Err() == nil {
				golog.Warn("ui stream backend read error", "module", logModule, "udid", udid, "error", readErr.Error())
			}
			return
		}
	}
}

func streamCodec(h264 bool) string {
	if h264 {
		return "h264"
	}
	return "mjpeg"
}

// ScreenshotStream serves an MJPEG (multipart/x-mixed-replace) stream of device
// screenshots captured via the instruments screenshot service. It streams until
// the client disconnects or the screenshot source fails.
//
//	?quality=<1-100>   optional JPEG quality (default 80)
func ScreenshotStream(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	udid := device.Properties.SerialNumber

	conn, err := instruments.NewScreenshotService(device)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer conn.Close()

	quality := 0 // 0 => instruments default (80)
	if raw := strings.TrimSpace(paramOrHeader(c, "quality")); raw != "" {
		q, convErr := parseQuality(raw)
		if convErr != nil {
			RespondError(c, http.StatusBadRequest, convErr)
			return
		}
		quality = q
	}

	c.Writer.Header().Set("Content-Type", mjpegStreamCType)
	c.Writer.Header().Set("Cache-Control", "no-cache, private")
	c.Writer.Header().Set("Connection", "close")
	c.Writer.WriteHeader(http.StatusOK)

	golog.Info("streaming device screenshots to client", "module", logModule, "udid", udid)

	ctx := c.Request.Context()
	flusher, _ := c.Writer.(http.Flusher)

	// Bridge the instruments frame source into a channel so the multipart
	// framing loop (writeMJPEGFrames) is a single, device-free-testable core.
	frames := make(chan []byte, 4)
	go func() {
		defer close(frames)
		err := instruments.StreamJPEGFrames(ctx, conn, quality, func(jpg []byte) {
			// Copy: the callback does not own the slice past the call, and the
			// frame is consumed asynchronously by the writer loop.
			frame := append([]byte(nil), jpg...)
			select {
			case frames <- frame:
			case <-ctx.Done():
			}
		})
		if err != nil {
			golog.Warn("screenshot frame source stopped", "module", logModule, "udid", udid, "error", err.Error())
		}
	}()

	if writeErr := writeMJPEGFrames(ctx, c.Writer, flusher, frames); writeErr != nil {
		golog.Debug("screenshot stream ended", "module", logModule, "udid", udid, "error", writeErr.Error())
	}
}

// writeMJPEGFrames writes each JPEG frame received on frames as a
// multipart/x-mixed-replace part, flushing after every frame. It returns when
// the context is canceled, a client write fails, or the frames channel is
// closed. This is the device-free-testable framing core shared by the handler.
func writeMJPEGFrames(ctx context.Context, w io.Writer, flusher http.Flusher, frames <-chan []byte) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case jpg, ok := <-frames:
			if !ok {
				return nil
			}
			if err := writeMJPEGFrame(w, jpg); err != nil {
				return err
			}
			if flusher != nil {
				flusher.Flush()
			}
		}
	}
}

// parseQuality validates an optional JPEG quality query param (1..100).
func parseQuality(raw string) (int, error) {
	q, err := strconv.Atoi(raw)
	if err != nil || q < 1 || q > 100 {
		return 0, errors.New("invalid 'quality'; expected an integer in 1..100")
	}
	return q, nil
}

// writeMJPEGFrame writes a single multipart/x-mixed-replace JPEG part.
func writeMJPEGFrame(w io.Writer, jpg []byte) error {
	if _, err := io.WriteString(w, fmt.Sprintf(mjpegStreamPartHead, len(jpg))); err != nil {
		return err
	}
	if _, err := w.Write(jpg); err != nil {
		return err
	}
	if _, err := io.WriteString(w, mjpegStreamPartFoot); err != nil {
		return err
	}
	return nil
}
