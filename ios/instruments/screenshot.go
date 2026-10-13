package instruments

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/golog"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"sync"
	"time"
)

const screenshotServiceName string = "com.apple.instruments.server.services.screenshot"

type ScreenshotService struct {
	channel *dtx.Channel
	conn    *dtx.Connection
}

func NewScreenshotService(device ios.DeviceEntry) (*ScreenshotService, error) {
	dtxConn, err := connectInstruments(device)
	if err != nil {
		return nil, err
	}
	processControlChannel := dtxConn.RequestChannelIdentifier(screenshotServiceName, loggingDispatcher{dtxConn})
	return &ScreenshotService{channel: processControlChannel, conn: dtxConn}, nil
}

func (d *ScreenshotService) Close() {
	d.conn.Close()
}

func (d *ScreenshotService) TakeScreenshot() ([]byte, error) {
	msg, err := d.channel.MethodCall("takeScreenshot")
	if err != nil {
		return nil, fmt.Errorf("TakeScreenshot: %s", err)
	}
	if len(msg.Payload) == 0 {
		return nil, fmt.Errorf("TakeScreenshot: empty response payload")
	}
	imageBytes, ok := msg.Payload[0].([]byte)
	if !ok {
		return nil, fmt.Errorf("TakeScreenshot: unexpected payload type %T", msg.Payload[0])
	}

	return imageBytes, nil
}

// pngToJPEG decodes a PNG frame (as returned by TakeScreenshot) and re-encodes
// it as JPEG at the given quality. It is the shared conversion step used by both
// the built-in MJPEG server and any external MJPEG consumer (for example the
// REST API).
func pngToJPEG(pngBytes []byte, quality int) ([]byte, error) {
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, fmt.Errorf("pngToJPEG: decode png: %w", err)
	}
	var b bytes.Buffer
	w := bufio.NewWriter(&b)
	if err := jpeg.Encode(w, img, &jpeg.Options{Quality: quality}); err != nil {
		return nil, fmt.Errorf("pngToJPEG: encode jpeg: %w", err)
	}
	if err := w.Flush(); err != nil {
		return nil, fmt.Errorf("pngToJPEG: flush jpeg: %w", err)
	}
	return b.Bytes(), nil
}

// StreamJPEGFrames is the reusable frame source shared by every MJPEG consumer.
// It takes screenshots on conn in a loop, converts each PNG to a JPEG frame at
// the given quality (a quality <= 0 selects the default of 80), and calls emit
// with the raw JPEG bytes for every frame. The emit callback owns the slice for
// the duration of the call; it must not retain it across calls.
//
// It runs until ctx is canceled or a screenshot/conversion error occurs, then
// returns. A screenshot error is returned; ctx cancellation returns nil. The
// caller owns conn and its lifecycle.
func StreamJPEGFrames(ctx context.Context, conn *ScreenshotService, quality int, emit func([]byte)) error {
	if quality <= 0 {
		quality = 80
	}
	for {
		if err := ctx.Err(); err != nil {
			return nil
		}
		start := time.Now()
		pngBytes, err := conn.TakeScreenshot()
		if err != nil {
			// Stop the streaming loop instead of killing the host process; a
			// screenshot failure must not take down a caller embedding go-ios.
			golog.Error("screenshot failed, stopping screenshot loop", "module", logModule, "error", err)
			return err
		}
		golog.Debug("shot done", "module", logModule, "seconds", time.Since(start).Seconds())
		jpg, err := pngToJPEG(pngBytes, quality)
		if err != nil {
			golog.Warn("failed converting frame", "module", logModule, "error", err)
			continue
		}
		golog.Debug("conversion done", "module", logModule, "seconds", time.Since(start).Seconds())
		// Re-check cancellation before emitting so a canceled context does not
		// push a final frame to a gone consumer.
		if ctx.Err() != nil {
			return nil
		}
		emit(jpg)
	}
}

// MJPEG server code
var consumers sync.Map

func StartMJPEGStreamingServer(device ios.DeviceEntry, port string) error {
	conn, err := NewScreenshotService(device)
	if err != nil {
		return err
	}
	defer conn.Close()

	// Fan a single frame source out to every connected mjpegHandler consumer,
	// preserving the previous behavior (one screenshot loop shared by all
	// clients registered in the consumers map).
	go func() {
		err := StreamJPEGFrames(context.Background(), conn, 80, func(jpg []byte) {
			consumers.Range(func(key, value interface{}) bool {
				c := value.(chan []byte)
				// Copy: the callback does not own the slice past the call, and
				// consumers receive asynchronously.
				frame := append([]byte(nil), jpg...)
				go func() { c <- frame }()
				return true
			})
		})
		if err != nil {
			golog.Error("mjpeg frame source stopped", "module", logModule, "udid", device.Properties.SerialNumber, "error", err)
		}
	}()

	http.HandleFunc("/", mjpegHandler)
	location := fmt.Sprintf("0.0.0.0:%s", port)
	golog.Info("starting server, open your browser here", "module", logModule, "udid", device.Properties.SerialNumber, "host", "0.0.0.0", "port", port, "url", fmt.Sprintf("http://%s/", location))
	return http.ListenAndServe(location, nil)
}

const (
	mjpegFrameFooter = "\r\n\r\n"
	mjpegFrameHeader = "--BoundaryString\r\nContent-type: image/jpg\r\nContent-Length: %d\r\n\r\n"
)

func mjpegHandler(w http.ResponseWriter, r *http.Request) {
	golog.Info("starting mjpeg stream for new client", "module", logModule)
	c := make(chan []byte)
	consumers.Store(r, c)
	w.Header().Add("Server", "go-ios-screenshotr-mjpeg-stream")
	w.Header().Add("Connection", "Close")
	w.Header().Add("Content-Type", "multipart/x-mixed-replace; boundary=--BoundaryString")
	w.Header().Add("Max-Age", "0")
	w.Header().Add("Expires", "0")
	w.Header().Add("Cache-Control", "no-cache, private")
	w.Header().Add("Pragma", "no-cache")
	w.Header().Add("Access-Control-Allow-Origin", "*")

	// io.WriteString(w, mjpegStreamHeader)
	w.WriteHeader(200)
	for {
		jpg := <-c
		_, err := io.WriteString(w, fmt.Sprintf(mjpegFrameHeader, len(jpg)))
		if err != nil {
			break
		}
		_, err = w.Write(jpg)
		if err != nil {
			break
		}
		_, err = io.WriteString(w, mjpegFrameFooter)
		if err != nil {
			break
		}
	}
	consumers.Delete(r)
	close(c)
	golog.Info("client disconnected", "module", logModule)
}
