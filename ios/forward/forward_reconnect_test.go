package forward

import (
	"context"
	"encoding/binary"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	plist "howett.net/plist"
)

// connectRequest is the subset of the usbmux Connect message the fake usbmuxd
// needs to decide whether the requested device exists.
type connectRequest struct {
	MessageType string
	DeviceID    uint32
	PortNumber  uint16
}

// startFakeUsbmuxd stands up a minimal usbmuxd on a unix socket in a temp dir
// and returns the socket path. It speaks just enough of the protocol for
// Connect: a Connect for goodDeviceID answers success and then echoes every
// byte back (playing the device-side service), a Connect for any other id
// answers mux error 2 — which is what real usbmuxd does for the stale id of a
// re-plugged device (issue #378).
func startFakeUsbmuxd(t *testing.T, goodDeviceID int) string {
	t.Helper()
	// os.MkdirTemp instead of t.TempDir: unix socket paths have a ~104 byte
	// limit and t.TempDir paths can get long on macOS.
	dir, err := os.MkdirTemp("", "muxsock")
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}
	sockPath := filepath.Join(dir, "usbmuxd")
	l, err := net.Listen("unix", sockPath)
	if err != nil {
		t.Fatalf("listen on fake usbmuxd socket: %v", err)
	}
	t.Cleanup(func() {
		l.Close()
		os.RemoveAll(dir)
	})
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				return
			}
			go serveFakeUsbmuxdConn(conn, goodDeviceID)
		}
	}()
	return sockPath
}

// serveFakeUsbmuxdConn handles one usbmuxd client connection: read a single
// Connect request, answer it, and on success switch the socket into echo mode
// like usbmuxd hands the socket over to the device service.
func serveFakeUsbmuxdConn(conn net.Conn, goodDeviceID int) {
	defer conn.Close()
	var header ios.UsbMuxHeader
	if err := binary.Read(conn, binary.LittleEndian, &header); err != nil {
		return
	}
	payload := make([]byte, header.Length-16)
	if _, err := io.ReadFull(conn, payload); err != nil {
		return
	}
	var req connectRequest
	if _, err := plist.Unmarshal(payload, &req); err != nil {
		return
	}
	number := uint32(0)
	if int(req.DeviceID) != goodDeviceID {
		number = 2 // usbmuxd result code for "device not found"
	}
	respBytes := ios.ToPlistBytes(ios.MuxResponse{MessageType: "Result", Number: number})
	respHeader := ios.UsbMuxHeader{Length: uint32(16 + len(respBytes)), Version: 1, Request: 8, Tag: header.Tag}
	if err := binary.Write(conn, binary.LittleEndian, respHeader); err != nil {
		return
	}
	if _, err := conn.Write(respBytes); err != nil {
		return
	}
	if number != 0 {
		return
	}
	_, _ = io.Copy(conn, conn)
}

// TestStaleDeviceIDIsReResolved reproduces issue #378: the forward captured a
// usbmux device id, the device was re-plugged and got a new id, so connecting
// with the stale id fails. The proxy must then re-resolve the device and
// succeed with the fresh id, keeping the forwarded port usable end-to-end.
func TestStaleDeviceIDIsReResolved(t *testing.T) {
	const staleID, freshID = 5, 42
	sockPath := startFakeUsbmuxd(t, freshID)
	t.Setenv("USBMUXD_SOCKET_ADDRESS", "unix://"+sockPath)

	var resolveCalls atomic.Int32
	resolve := func() (int, error) {
		resolveCalls.Add(1)
		return freshID, nil
	}

	clientSide, clientConn := net.Pipe()
	defer clientSide.Close()
	done := make(chan error, 1)
	go func() {
		done <- startProxyConnection(context.Background(), clientConn, staleID, 8100, resolve)
	}()

	// The session must be alive end-to-end: bytes written by the client come
	// back from the fake device echo.
	go func() { _, _ = clientSide.Write([]byte("ping")) }()
	buf := make([]byte, 4)
	_ = clientSide.SetReadDeadline(time.Now().Add(10 * time.Second))
	if _, err := io.ReadFull(clientSide, buf); err != nil {
		t.Fatalf("echo read through forward: %v", err)
	}
	if string(buf) != "ping" {
		t.Fatalf("echo through forward: got %q want %q", buf, "ping")
	}
	if got := resolveCalls.Load(); got != 1 {
		t.Fatalf("expected exactly one device re-resolve, got %d", got)
	}

	clientSide.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("startProxyConnection: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("startProxyConnection did not return after client close")
	}
}

// TestCurrentDeviceIDConnectsWithoutResolve verifies the happy path is
// untouched: when the captured id still works, the resolver is never invoked.
func TestCurrentDeviceIDConnectsWithoutResolve(t *testing.T) {
	const deviceID = 7
	sockPath := startFakeUsbmuxd(t, deviceID)
	t.Setenv("USBMUXD_SOCKET_ADDRESS", "unix://"+sockPath)

	resolve := func() (int, error) {
		t.Error("resolve must not be called when the current id connects")
		return deviceID, nil
	}

	clientSide, clientConn := net.Pipe()
	defer clientSide.Close()
	done := make(chan error, 1)
	go func() {
		done <- startProxyConnection(context.Background(), clientConn, deviceID, 8100, resolve)
	}()

	go func() { _, _ = clientSide.Write([]byte("ok")) }()
	buf := make([]byte, 2)
	_ = clientSide.SetReadDeadline(time.Now().Add(10 * time.Second))
	if _, err := io.ReadFull(clientSide, buf); err != nil {
		t.Fatalf("echo read through forward: %v", err)
	}

	clientSide.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("startProxyConnection: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("startProxyConnection did not return after client close")
	}
}

// TestStaleDeviceIDWithoutResolverFails pins the library-API behavior of
// StartNewProxyConnection (no resolver): a stale id fails and the client
// connection is closed.
func TestStaleDeviceIDWithoutResolverFails(t *testing.T) {
	sockPath := startFakeUsbmuxd(t, 42)
	t.Setenv("USBMUXD_SOCKET_ADDRESS", "unix://"+sockPath)

	clientSide, clientConn := net.Pipe()
	defer clientSide.Close()
	if err := StartNewProxyConnection(context.Background(), clientConn, 5, 8100); err == nil {
		t.Fatal("expected connect with stale device id to fail without a resolver")
	}
	// The client conn must have been closed on failure.
	_ = clientSide.SetReadDeadline(time.Now().Add(10 * time.Second))
	if _, err := clientSide.Read(make([]byte, 1)); err != io.EOF {
		t.Fatalf("expected client conn closed (EOF), got %v", err)
	}
}
