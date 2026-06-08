package forward

import (
	"context"
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// fakeDeviceConn adapts a net.Conn to the deviceRWConn interface so proxyConns
// can be driven by an in-memory net.Pipe instead of a real usbmux/device
// connection. The far end of the pipe stands in for "the device".
type fakeDeviceConn struct{ c net.Conn }

func (f fakeDeviceConn) Reader() io.Reader { return f.c }
func (f fakeDeviceConn) Writer() io.Writer { return f.c }
func (f fakeDeviceConn) Close() error      { return f.c.Close() }

// newProxySession wires up a proxyConns instance over two net.Pipes and returns
// the test-side ends plus a cancel func and a done channel that closes when
// proxyConns returns. clientSide writes/reads as the host client; deviceSide
// writes/reads as the device.
func newProxySession(t *testing.T) (clientSide, deviceSide net.Conn, cancel context.CancelFunc, done chan struct{}) {
	t.Helper()
	clientSide, clientConn := net.Pipe()
	deviceSide, deviceConn := net.Pipe()
	ctx, cancel := context.WithCancel(context.Background())
	done = make(chan struct{})
	go func() {
		proxyConns(ctx, clientConn, fakeDeviceConn{deviceConn}, 1, 8100)
		close(done)
	}()
	return clientSide, deviceSide, cancel, done
}

// awaitReturn fails the test if proxyConns does not return promptly — i.e. if a
// teardown path leaks a goroutine.
func awaitReturn(t *testing.T, done chan struct{}, what string) {
	t.Helper()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatalf("proxyConns leaked: %s did not return", what)
	}
}

// TestProxyConnsBidirectional verifies bytes flow both directions through the
// pump, and that it tears down cleanly when the client side closes.
func TestProxyConnsBidirectional(t *testing.T) {
	clientSide, deviceSide, cancel, done := newProxySession(t)
	defer cancel()

	// client -> device
	go func() { _, _ = clientSide.Write([]byte("ping")) }()
	buf := make([]byte, 4)
	if _, err := io.ReadFull(deviceSide, buf); err != nil {
		t.Fatalf("device read: %v", err)
	}
	if string(buf) != "ping" {
		t.Fatalf("client->device: got %q want %q", buf, "ping")
	}

	// device -> client
	go func() { _, _ = deviceSide.Write([]byte("pong")) }()
	if _, err := io.ReadFull(clientSide, buf); err != nil {
		t.Fatalf("client read: %v", err)
	}
	if string(buf) != "pong" {
		t.Fatalf("device->client: got %q want %q", buf, "pong")
	}

	// Closing the client end must tear the whole session down.
	clientSide.Close()
	awaitReturn(t, done, "client close")
}

// TestProxyConnsCtxCancelTeardown verifies cancelling the parent context tears
// the session down and returns, even with no IO in flight.
func TestProxyConnsCtxCancelTeardown(t *testing.T) {
	_, _, cancel, done := newProxySession(t)
	cancel()
	awaitReturn(t, done, "ctx cancel")
}

// TestProxyConnsDeviceCloseTeardown verifies the device side closing tears the
// session down and returns.
func TestProxyConnsDeviceCloseTeardown(t *testing.T) {
	_, deviceSide, cancel, done := newProxySession(t)
	defer cancel()
	deviceSide.Close()
	awaitReturn(t, done, "device close")
}

// TestProxyConnsSimultaneousTeardown hammers the teardown path: for many
// iterations it fires ctx-cancel, client-close and device-close all at once at
// a freshly started session. Run under -race this is the strongest check that
// the three concurrent teardown routes don't race or double-close.
func TestProxyConnsSimultaneousTeardown(t *testing.T) {
	for i := 0; i < 300; i++ {
		clientSide, deviceSide, cancel, done := newProxySession(t)

		var start sync.WaitGroup
		start.Add(1)
		var racers sync.WaitGroup
		for _, fn := range []func(){
			func() { cancel() },
			func() { clientSide.Close() },
			func() { deviceSide.Close() },
		} {
			racers.Add(1)
			go func(f func()) {
				defer racers.Done()
				start.Wait() // line all three up to fire together
				f()
			}(fn)
		}
		start.Done()
		racers.Wait()

		awaitReturn(t, done, "simultaneous teardown")
		cancel()
	}
}

// TestProxyConnsConcurrentHammer runs many sessions in parallel, each pushing a
// little traffic and then tearing down via a different route, so io.Copy is
// genuinely mid-flight when teardown hits. The goal is to surface any residual
// race or goroutine leak across the whole forward pump under -race.
func TestProxyConnsConcurrentHammer(t *testing.T) {
	const sessions = 400
	var wg sync.WaitGroup
	for i := 0; i < sessions; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			clientSide, deviceSide, cancel, done := newProxySession(t)
			defer cancel()

			// Push a byte each way so both copy goroutines are active. Use
			// goroutines because net.Pipe is synchronous (write blocks on read).
			var io1, io2 sync.WaitGroup
			io1.Add(2)
			go func() { defer io1.Done(); _, _ = clientSide.Write([]byte{byte(i)}) }()
			go func() {
				defer io1.Done()
				b := make([]byte, 1)
				_, _ = io.ReadFull(deviceSide, b)
			}()
			io2.Add(2)
			go func() { defer io2.Done(); _, _ = deviceSide.Write([]byte{byte(i)}) }()
			go func() {
				defer io2.Done()
				b := make([]byte, 1)
				_, _ = io.ReadFull(clientSide, b)
			}()
			io1.Wait()
			io2.Wait()

			// Tear down via a rotating route while traffic was just flowing.
			switch i % 4 {
			case 0:
				cancel()
			case 1:
				clientSide.Close()
			case 2:
				deviceSide.Close()
			case 3:
				clientSide.Close()
				deviceSide.Close()
				cancel()
			}
			awaitReturn(t, done, "hammer session")
			clientSide.Close()
			deviceSide.Close()
		}(i)
	}
	wg.Wait()
}

// TestForwardCloseLifecycle exercises the listener accept loop: Forward stands
// up a real local listener, and Close must stop it cleanly. With no client ever
// connecting, StartNewProxyConnection (which needs usbmux) is never reached, so
// this stays device-free. It guards the #639 fix — a clean Close unblocks
// Accept with net.ErrClosed and the loop exits without treating it as an error.
func TestForwardCloseLifecycle(t *testing.T) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	cl := &ConnListener{listener: l, quit: make(chan interface{})}

	accepted := make(chan struct{})
	go func() {
		connectionAccept(cl, 1, 8100)
		close(accepted)
	}()

	// Let the accept loop park in Accept(), then close it.
	time.Sleep(20 * time.Millisecond)
	if err := cl.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	select {
	case <-accepted:
	case <-time.After(5 * time.Second):
		t.Fatal("connectionAccept did not exit after Close")
	}
}
