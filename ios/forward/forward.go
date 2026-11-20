package forward

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
)

const logModule = "go-ios/forward"

type iosproxy struct {
	tcpConn    net.Conn
	deviceConn ios.DeviceConnectionInterface
}

type ConnListener struct {
	listener net.Listener
	quit     chan interface{}
}

// resolveDeviceID returns the current usbmux device id of the forwarded device.
// usbmuxd assigns a fresh id every time a device is plugged in, so the id
// captured when the forward started goes stale after a disconnect/reconnect.
// Re-resolving through this func lets a long-running forward survive a replug
// (issue #378).
type resolveDeviceID func() (int, error)

// Forward forwards every connection made to the hostPort to whatever service runs inside an app on the device on phonePort.
// Port values must be between 1 and 65535.
func Forward(device ios.DeviceEntry, hostPort uint16, phonePort uint16) (*ConnListener, error) {
	if hostPort == 0 {
		return nil, fmt.Errorf("forward: invalid host port: port must be at least 1")
	}
	if phonePort == 0 {
		return nil, fmt.Errorf("forward: invalid target port: port must be at least 1")
	}
	golog.Info("start listening, forwarding to device", "module", logModule, "udid", device.Properties.SerialNumber, "hostPort", hostPort, "phonePort", phonePort)
	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", hostPort))
	if err != nil {
		return nil, fmt.Errorf("forward: failed listener with err: %w", err)
	}
	cl := &ConnListener{
		listener: l,
		quit:     make(chan interface{}),
	}

	go connectionAccept(cl, device.DeviceID, phonePort, deviceIDResolver(device))

	return cl, nil
}

// deviceIDResolver builds a resolveDeviceID that looks the device up again by
// its udid (the serial survives a replug, the usbmux id does not). Without a
// udid there is nothing stable to search by, so the captured id is returned
// as-is.
func deviceIDResolver(device ios.DeviceEntry) resolveDeviceID {
	udid := device.Properties.SerialNumber
	deviceID := device.DeviceID
	if udid == "" {
		return func() (int, error) { return deviceID, nil }
	}
	return func() (int, error) {
		fresh, err := ios.GetDevice(udid)
		if err != nil {
			return 0, err
		}
		return fresh.DeviceID, nil
	}
}

// Close stops listening on the host port for the forwarded connection
func (cl *ConnListener) Close() error {
	close(cl.quit)

	err := cl.listener.Close()
	if err != nil {
		return fmt.Errorf("forward: failed closing listener with err: %w", err)
	}

	return nil
}

func connectionAccept(cl *ConnListener, deviceID int, phonePort uint16, resolve resolveDeviceID) {
	for {
		select {
		case <-cl.quit:
			golog.Info("closed listener successfully", "module", logModule, "deviceID", deviceID, "phonePort", phonePort)
			return
		default:
			clientConn, err := cl.listener.Accept()
			if err != nil {
				// Close() shuts the listener down, which unblocks this Accept()
				// with net.ErrClosed. That's a clean teardown, not a failure, so
				// don't log it as an error — just exit the accept loop.
				if errors.Is(err, net.ErrClosed) {
					golog.Info("closed listener successfully", "module", logModule, "deviceID", deviceID, "phonePort", phonePort)
					return
				}
				golog.Error("error accepting new connection", "module", logModule, "deviceID", deviceID, "phonePort", phonePort, "error", err)
				continue
			}
			golog.Debug("new client connected", "module", logModule, "deviceID", deviceID, "phonePort", phonePort, "conn", fmt.Sprintf("%#v", cl))
			go startProxyConnection(context.TODO(), clientConn, deviceID, phonePort, resolve)
		}
	}
}

// StartNewProxyConnection connects clientConn to phonePort on the device with
// the given usbmux device id and pumps bytes in both directions until either
// side closes.
func StartNewProxyConnection(ctx context.Context, clientConn io.ReadWriteCloser, deviceID int, phonePort uint16) error {
	return startProxyConnection(ctx, clientConn, deviceID, phonePort, nil)
}

func startProxyConnection(ctx context.Context, clientConn io.ReadWriteCloser, deviceID int, phonePort uint16, resolve resolveDeviceID) error {
	deviceConn, connectedID, err := connectToPhonePort(deviceID, phonePort, resolve)
	if err != nil {
		clientConn.Close()
		return err
	}
	golog.Debug("connected to port", "module", logModule, "deviceID", connectedID, "conn", fmt.Sprintf("%#v", clientConn), "phonePort", phonePort)

	proxyConns(ctx, clientConn, deviceConn, connectedID, phonePort)
	return nil
}

// connectToPhonePort connects to phonePort using deviceID. When that fails and
// a resolver is available, the device is looked up again — a replugged device
// keeps its serial but gets a new usbmux id — and the connect is retried once
// with the fresh id. It returns the device connection together with the id it
// was actually established on.
func connectToPhonePort(deviceID int, phonePort uint16, resolve resolveDeviceID) (ios.DeviceConnectionInterface, int, error) {
	deviceConn, err := dialPhonePort(deviceID, phonePort)
	if err == nil {
		return deviceConn, deviceID, nil
	}
	if resolve == nil {
		return nil, deviceID, err
	}
	freshID, resolveErr := resolve()
	if resolveErr != nil {
		golog.Debug("could not re-resolve device after failed connect", "module", logModule, "deviceID", deviceID, "phonePort", phonePort, "error", resolveErr)
		return nil, deviceID, err
	}
	if freshID == deviceID {
		return nil, deviceID, err
	}
	golog.Info("usbmux device id changed since the forward started, reconnecting", "module", logModule, "deviceID", freshID, "staleDeviceID", deviceID, "phonePort", phonePort)
	deviceConn, retryErr := dialPhonePort(freshID, phonePort)
	if retryErr != nil {
		return nil, freshID, retryErr
	}
	return deviceConn, freshID, nil
}

// dialPhonePort opens a fresh connection to usbmuxd and asks it to connect to
// phonePort on the device with the given usbmux id.
func dialPhonePort(deviceID int, phonePort uint16) (ios.DeviceConnectionInterface, error) {
	usbmuxConn, err := ios.NewUsbMuxConnectionSimple()
	if err != nil {
		golog.Error("could not connect to usbmuxd", "module", logModule, "deviceID", deviceID, "phonePort", phonePort, "error", err)
		return nil, fmt.Errorf("could not connect to usbmuxd: %v", err)
	}
	if muxError := usbmuxConn.Connect(deviceID, phonePort); muxError != nil {
		usbmuxConn.Close()
		golog.Debug("could not connect to phone", "module", logModule, "deviceID", deviceID, "error", muxError, "phonePort", phonePort)
		return nil, fmt.Errorf("could not connect to port:%d on iOS: %v", phonePort, muxError)
	}
	return usbmuxConn.ReleaseDeviceConnection(), nil
}

// deviceRWConn is the subset of ios.DeviceConnectionInterface that the proxy
// pump needs. Narrowing it to this lets proxyConns be exercised with in-memory
// pipes in tests, without standing up a usbmux/device connection.
type deviceRWConn interface {
	Reader() io.Reader
	Writer() io.Writer
	Close() error
}

// proxyConns is the device-free core of the forward: it pumps bytes in both
// directions between clientConn and deviceConn until either side closes (or ctx
// is cancelled), then tears both down exactly once and waits for both copy
// goroutines to finish. Teardown is funnelled through a sync.Once so the two
// copiers and the ctx watcher can all race to close without a data race or a
// double-close, and so it returns with no goroutines left running.
func proxyConns(ctx context.Context, clientConn io.ReadWriteCloser, deviceConn deviceRWConn, deviceID int, phonePort uint16) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Closing the conns unblocks whichever io.Copy is still parked in a read,
	// and cancel() releases the watcher below. sync.Once makes this safe to
	// call from all three goroutines concurrently.
	var once sync.Once
	teardown := func() {
		once.Do(func() {
			cancel()
			clientConn.Close()
			deviceConn.Close()
		})
	}

	var wg sync.WaitGroup
	copyAndClose := func(dst io.Writer, src io.Reader, msg string) {
		defer wg.Done()
		_, err := io.Copy(dst, src)
		teardown()

		if err == nil || errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
			golog.Debug(msg, "module", logModule, "deviceID", deviceID, "phonePort", phonePort)
		} else {
			golog.Error(msg, "module", logModule, "deviceID", deviceID, "phonePort", phonePort, "error", err)
		}
	}

	wg.Add(2)
	go copyAndClose(clientConn, deviceConn.Reader(), "forward: close clientConn <-- deviceConn")
	go copyAndClose(deviceConn.Writer(), clientConn, "forward: close clientConn --> deviceConn")

	<-ctx.Done()
	teardown()
	wg.Wait()
}

func (proxyConn *iosproxy) Close() {
	proxyConn.tcpConn.Close()
}
