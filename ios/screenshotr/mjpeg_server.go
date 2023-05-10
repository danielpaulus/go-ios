package screenshotr

import (
	"bufio"
	"bytes"
	"fmt"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

var (
	consumers       sync.Map
	conversionQueue = make(chan []byte, 20)
)

func StartStreamingServer(device ios.DeviceEntry, port string) error {
	conn, err := New(device)
	if err != nil {
		return err
	}
	go startScreenshotting(conn)
	go startConversionQueue()
	http.HandleFunc("/", mjpegHandler)
	location := fmt.Sprintf("0.0.0.0:%s", port)
	log.WithFields(log.Fields{"host": "0.0.0.0", "port": port}).Infof("starting server, open your browser here: http://%s/", location)
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
			log.Warnf("failed decoding png %v", err)
			continue
		}
		var b bytes.Buffer
		foo := bufio.NewWriter(&b)
		err = jpeg.Encode(foo, img, &opt)
		if err != nil {
			log.Warnf("failed encoding jpg %v", err)
			continue
		}
		elapsed := time.Since(start)
		log.Debugf("conversion took %fs", elapsed.Seconds())
		consumers.Range(func(key, value interface{}) bool {
			c := value.(chan []byte)
			go func() { c <- b.Bytes() }()
			return true
		})
	}
}

func startScreenshotting(conn *Connection) {
	for {
		start := time.Now()
		pngBytes, err := conn.TakeScreenshot()
		if err != nil {
			log.Fatal("screenshotr failed", err)
		}
		elapsed := time.Since(start)
		log.Debugf("shot took %fs", elapsed.Seconds())
		conversionQueue <- pngBytes
	}
}

const (
	mjpegFrameFooter = "\r\n\r\n"
	mjpegFrameHeader = "--BoundaryString\r\nContent-type: image/jpg\r\nContent-Length: %d\r\n\r\n"
)

func mjpegHandler(w http.ResponseWriter, r *http.Request) {
	log.Infof("starting mjpeg stream for new client")
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
	log.Info("client disconnected")
}
