package instruments

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/deviceinfo"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/screencapture"
)

const screenshotServiceName string = "com.apple.instruments.server.services.screenshot"
const coreScreenshotServiceName = "com.apple.coredevice.screencaptureservice"

type screenshotCapturer interface {
	TakeScreenshot(context.Context, screencapture.ScreenshotOptions) ([]byte, error)
	Close() error
}

type displayInfoConnection interface {
	GetDisplayInfo() (map[string]interface{}, error)
	Close() error
}

type ScreenshotService struct {
	channel     *dtx.Channel
	conn        *dtx.Connection
	capture     screenshotCapturer
	displayInfo displayInfoConnection
}

// NewScreenshotService captures the active built-in display when the device
// offers CoreDevice screen capture (iOS 27+). Older devices use Instruments.
func NewScreenshotService(device ios.DeviceEntry) (*ScreenshotService, error) {
	if device.SupportsRsd() && device.Rsd.GetPort(coreScreenshotServiceName) != 0 {
		info, err := deviceinfo.NewDeviceInfo(device)
		if err != nil {
			return nil, err
		}
		capture, err := screencapture.New(device)
		if err != nil {
			_ = info.Close()
			return nil, err
		}
		return &ScreenshotService{capture: capture, displayInfo: info}, nil
	}
	dtxConn, err := connectInstruments(device)
	if err != nil {
		return nil, err
	}
	processControlChannel := dtxConn.RequestChannelIdentifier(screenshotServiceName, loggingDispatcher{dtxConn})
	return &ScreenshotService{channel: processControlChannel, conn: dtxConn}, nil
}

func (d *ScreenshotService) Close() {
	if d.capture != nil {
		_ = d.capture.Close()
		_ = d.displayInfo.Close()
		return
	}
	d.conn.Close()
}

func (d *ScreenshotService) TakeScreenshot() ([]byte, error) {
	if d.capture != nil {
		// Refresh for every frame so streaming follows folding/unfolding too.
		info, err := d.displayInfo.GetDisplayInfo()
		if err != nil {
			return nil, fmt.Errorf("TakeScreenshot: failed to get display info: %w", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return d.capture.TakeScreenshot(ctx, screencapture.ScreenshotOptions{DisplayUniqueID: activeDisplayID(info)})
	}
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

// activeDisplayID ignores external displays and prefers the primary display if
// multiple built-in displays are active. With none active (e.g. a locked phone),
// an empty ID preserves the service's main-display default.
func activeDisplayID(info map[string]interface{}) string {
	displays, _ := info["displays"].([]interface{})
	selected := ""
	for _, entry := range displays {
		display, ok := entry.(map[string]interface{})
		if !ok || display["active"] != true || display["external"] != false {
			continue
		}
		id, _ := display["uniqueId"].(string)
		if id == "" {
			continue
		}
		if display["primary"] == true {
			return id
		}
		if selected == "" {
			selected = id
		}
	}
	return selected
}

// MJPEG server code
var (
	consumers       sync.Map
	conversionQueue = make(chan []byte, 20)
)

func StartMJPEGStreamingServer(device ios.DeviceEntry, port string) error {
	conn, err := NewScreenshotService(device)
	if err != nil {
		return err
	}
	defer conn.Close()

	go startScreenshotting(conn)
	go startConversionQueue()
	http.HandleFunc("/", mjpegHandler)
	location := fmt.Sprintf("0.0.0.0:%s", port)
	golog.Info("starting server, open your browser here", "module", logModule, "udid", device.Properties.SerialNumber, "host", "0.0.0.0", "port", port, "url", fmt.Sprintf("http://%s/", location))
	return http.ListenAndServe(location, nil)
}

func startConversionQueue() {
	var opt jpeg.Options
	opt.Quality = 80

	for {
		pngBytes := <-conversionQueue
		start := time.Now()
		img, err := png.Decode(bytes.NewReader(pngBytes))
		if err != nil {
			golog.Warn("failed decoding png", "module", logModule, "error", err)
			continue
		}
		var b bytes.Buffer
		foo := bufio.NewWriter(&b)
		err = jpeg.Encode(foo, img, &opt)
		if err != nil {
			golog.Warn("failed encoding jpg", "module", logModule, "error", err)
			continue
		}
		elapsed := time.Since(start)
		golog.Debug("conversion done", "module", logModule, "seconds", elapsed.Seconds())
		consumers.Range(func(key, value interface{}) bool {
			c := value.(chan []byte)
			go func() { c <- b.Bytes() }()
			return true
		})
	}
}

func startScreenshotting(conn *ScreenshotService) {
	for {
		start := time.Now()
		pngBytes, err := conn.TakeScreenshot()
		if err != nil {
			// Stop the streaming loop instead of killing the host process; a
			// screenshot failure must not take down a caller embedding go-ios.
			golog.Error("screenshot failed, stopping screenshot loop", "module", logModule, "error", err)
			return
		}
		elapsed := time.Since(start)
		golog.Debug("shot done", "module", logModule, "seconds", elapsed.Seconds())
		conversionQueue <- pngBytes
	}
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
