package api_test

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/restapi/api"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests are device-free. The UI stream is exercised against an
// httptest.Server standing in for a forwarded WDA/DeviceKit backend; the
// screenshot stream's multipart framing core is exercised via the exported
// WriteMJPEGFramesForTest hook with an injected frame channel.

// streamRouter wires UIStream behind a fake device middleware, matching the real
// route. Requests point at wdaURL via ?wdaUrl.
func streamRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(api.IOS_KEY, ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: "stream-test-udid"}})
		c.Next()
	})
	r.GET("/ui/stream", api.UIStream)
	return r
}

// TestUIStreamPipesBackendBody stands up a fake backend that emits a chunked
// multipart body and asserts the endpoint pipes it through verbatim with the
// backend's Content-Type preserved.
func TestUIStreamPipesBackendBody(t *testing.T) {
	chunks := []string{"--B\r\nContent-Type: image/jpeg\r\n\r\nframe-1\r\n", "--B\r\n\r\nframe-2\r\n"}
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/mjpeg", r.URL.Path)
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=B")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		for _, c := range chunks {
			_, _ = io.WriteString(w, c)
			flusher.Flush()
		}
	}))
	defer backend.Close()

	srv := httptest.NewServer(streamRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ui/stream?backend=devicekit&codec=mjpeg&wdaUrl=" + backend.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "multipart/x-mixed-replace; boundary=B", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, strings.Join(chunks, ""), string(body))
}

// TestUIStreamH264PreservesBackendContentType asserts an h264 stream over the
// devicekit backend forwards the backend's Content-Type.
func TestUIStreamH264PreservesBackendContentType(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/h264", r.URL.Path)
		w.Header().Set("Content-Type", "video/H264")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("h264-payload"))
	}))
	defer backend.Close()

	srv := httptest.NewServer(streamRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ui/stream?backend=devicekit&codec=h264&wdaUrl=" + backend.URL)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "video/H264", resp.Header.Get("Content-Type"))
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "h264-payload", string(body))
}

// TestUIStreamH264OverWDAReturns501 asserts requesting h264 against the wda
// backend maps ErrStreamUnsupported to 501.
func TestUIStreamH264OverWDAReturns501(t *testing.T) {
	srv := httptest.NewServer(streamRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ui/stream?backend=wda&codec=h264&wdaUrl=http://127.0.0.1:1")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotImplemented, resp.StatusCode)
}

// TestUIStreamUnreachableBackendReturns502 asserts a dead backend maps to 502.
func TestUIStreamUnreachableBackendReturns502(t *testing.T) {
	srv := httptest.NewServer(streamRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ui/stream?backend=devicekit&codec=mjpeg&wdaUrl=http://127.0.0.1:1")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

// TestUIStreamBadCodecReturns400 asserts an unknown codec maps to 400.
func TestUIStreamBadCodecReturns400(t *testing.T) {
	srv := httptest.NewServer(streamRouter())
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/ui/stream?codec=webm&wdaUrl=http://127.0.0.1:1")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestUIStreamStopsOnClientDisconnect asserts that cancelling the client request
// causes the endpoint to stop piping and close the backend body. The backend
// blocks forever after the first chunk; once the client cancels, its handler's
// Write must fail (context canceled), which we observe via backendDone.
func TestUIStreamStopsOnClientDisconnect(t *testing.T) {
	backendDone := make(chan struct{})
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=B")
		w.WriteHeader(http.StatusOK)
		flusher := w.(http.Flusher)
		_, _ = io.WriteString(w, "first-chunk")
		flusher.Flush()
		// Keep writing until the client (via the proxy) goes away.
		defer close(backendDone)
		for {
			select {
			case <-r.Context().Done():
				return
			default:
			}
			if _, err := io.WriteString(w, "x"); err != nil {
				return
			}
			flusher.Flush()
			time.Sleep(5 * time.Millisecond)
		}
	}))
	defer backend.Close()

	srv := httptest.NewServer(streamRouter())
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/ui/stream?backend=devicekit&codec=mjpeg&wdaUrl="+backend.URL, nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	buf := make([]byte, len("first-chunk"))
	_, err = io.ReadFull(resp.Body, buf)
	require.NoError(t, err)
	assert.Equal(t, "first-chunk", string(buf))

	cancel()
	_ = resp.Body.Close()

	select {
	case <-backendDone:
		// The proxy stopped consuming and closed the backend body, so the
		// backend handler observed the disconnect.
	case <-time.After(5 * time.Second):
		t.Fatal("backend did not observe client disconnect; proxy did not stop")
	}
}

// TestWriteMJPEGFramesFramesAndStops exercises the multipart framing core with an
// injected frame channel: it writes well-formed multipart parts for each frame
// and stops when the channel closes.
func TestWriteMJPEGFramesFramesAndStops(t *testing.T) {
	frames := make(chan []byte, 2)
	frames <- []byte("JPEGDATA1")
	frames <- []byte("JPEG2")
	close(frames)

	var buf strings.Builder
	err := api.WriteMJPEGFramesForTest(context.Background(), &buf, nil, frames)
	require.NoError(t, err)

	out := buf.String()
	// Two parts, each with the boundary + JPEG content-type + correct length.
	assert.Equal(t, 2, strings.Count(out, "--BoundaryString"))
	assert.Contains(t, out, "Content-Type: image/jpeg")
	assert.Contains(t, out, fmt.Sprintf("Content-Length: %d", len("JPEGDATA1")))
	assert.Contains(t, out, "JPEGDATA1")
	assert.Contains(t, out, "JPEG2")

	// Parse it as an actual multipart body to prove it is well-formed.
	parts := readMultipartParts(t, out)
	require.Len(t, parts, 2)
	assert.Equal(t, "JPEGDATA1", string(parts[0]))
	assert.Equal(t, "JPEG2", string(parts[1]))
}

// TestWriteMJPEGFramesStopsOnContextCancel asserts a canceled context stops the
// framing loop even while frames remain available.
func TestWriteMJPEGFramesStopsOnContextCancel(t *testing.T) {
	frames := make(chan []byte) // never closed, never fed after cancel
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	var mu sync.Mutex
	var buf strings.Builder
	go func() {
		done <- api.WriteMJPEGFramesForTest(ctx, lockedWriter{&mu, &buf}, nil, frames)
	}()

	// Feed one frame, then cancel.
	frames <- []byte("frame")
	cancel()

	select {
	case err := <-done:
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Fatal("framing loop did not stop on context cancel")
	}
}

type lockedWriter struct {
	mu  *sync.Mutex
	buf *strings.Builder
}

func (l lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.buf.Write(p)
}

// readMultipartParts parses a multipart/x-mixed-replace body (boundary
// BoundaryString) and returns each part's payload.
func readMultipartParts(t *testing.T, body string) [][]byte {
	t.Helper()
	var parts [][]byte
	reader := bufio.NewReader(strings.NewReader(body))
	for {
		// Read boundary line.
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		if !strings.HasPrefix(line, "--BoundaryString") {
			continue
		}
		// Read headers until blank line, capture Content-Length.
		length := -1
		for {
			h, err := reader.ReadString('\n')
			require.NoError(t, err)
			trimmed := strings.TrimRight(h, "\r\n")
			if trimmed == "" {
				break
			}
			if strings.HasPrefix(trimmed, "Content-Length:") {
				_, _ = fmt.Sscanf(trimmed, "Content-Length: %d", &length)
			}
		}
		require.Greater(t, length, -1, "part missing Content-Length")
		payload := make([]byte, length)
		_, err = io.ReadFull(reader, payload)
		require.NoError(t, err)
		parts = append(parts, payload)
	}
	return parts
}
