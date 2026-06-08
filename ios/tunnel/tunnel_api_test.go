package tunnel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRemoveTunnelStopsOnlyRequestedTunnel(t *testing.T) {
	var closedA atomic.Uint64
	var closedB atomic.Uint64
	tm := tunnelManagerWithTunnels(map[string]Tunnel{
		"serial-a": {
			Udid:   "serial-a",
			closer: func() error { closedA.Add(1); return nil },
		},
		"serial-b": {
			Udid:   "serial-b",
			closer: func() error { closedB.Add(1); return nil },
		},
	})

	err := tm.RemoveTunnel(context.Background(), "serial-a")
	if err != nil {
		t.Fatalf("RemoveTunnel returned error: %v", err)
	}

	if closedA.Load() != 1 {
		t.Fatalf("serial-a closer called %d times, want 1", closedA.Load())
	}
	if closedB.Load() != 0 {
		t.Fatalf("serial-b closer called %d times, want 0", closedB.Load())
	}
	if _, exists := tm.tunnels["serial-a"]; exists {
		t.Fatal("serial-a tunnel was not removed")
	}
	if _, exists := tm.tunnels["serial-b"]; !exists {
		t.Fatal("serial-b tunnel was removed")
	}
}

func TestRemoveTunnelReturnsNotFound(t *testing.T) {
	tm := tunnelManagerWithTunnels(nil)
	err := tm.RemoveTunnel(context.Background(), "missing")
	if !errors.Is(err, ErrTunnelNotFound) {
		t.Fatalf("RemoveTunnel error = %v, want ErrTunnelNotFound", err)
	}
}

func TestTunnelInfoMuxDeleteTunnel(t *testing.T) {
	var closed atomic.Uint64
	tm := tunnelManagerWithTunnels(map[string]Tunnel{
		"serial-a": {
			Udid:    "serial-a",
			Address: "fd00::1",
			RsdPort: 1234,
			closer:  func() error { closed.Add(1); return nil },
		},
		"serial-b": {
			Udid:    "serial-b",
			Address: "fd00::2",
			RsdPort: 5678,
			closer:  func() error { return nil },
		},
	})
	server := httptest.NewServer(tunnelInfoMux(tm))
	defer server.Close()

	host, port := serverHostPort(t, server.URL)
	if err := StopTunnelForDevice("serial-a", host, port); err != nil {
		t.Fatalf("StopTunnelForDevice returned error: %v", err)
	}

	if closed.Load() != 1 {
		t.Fatalf("closer called %d times, want 1", closed.Load())
	}
	if _, exists := tm.tunnels["serial-a"]; exists {
		t.Fatal("serial-a tunnel was not removed")
	}
	if _, exists := tm.tunnels["serial-b"]; !exists {
		t.Fatal("serial-b tunnel was removed")
	}
}

func TestTunnelInfoMuxDeleteMissingTunnel(t *testing.T) {
	tm := tunnelManagerWithTunnels(nil)
	server := httptest.NewServer(tunnelInfoMux(tm))
	defer server.Close()

	host, port := serverHostPort(t, server.URL)
	err := StopTunnelForDevice("missing", host, port)
	if !errors.Is(err, ErrTunnelNotFound) {
		t.Fatalf("StopTunnelForDevice error = %v, want ErrTunnelNotFound", err)
	}
}

func TestTunnelInfoMuxRejectsUnsupportedMethod(t *testing.T) {
	tm := tunnelManagerWithTunnels(nil)
	req := httptest.NewRequest(http.MethodPost, "/tunnel/serial-a", nil)
	rec := httptest.NewRecorder()

	tunnelInfoMux(tm).ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestTunnelInfoForDeviceMapsNotFound(t *testing.T) {
	tm := tunnelManagerWithTunnels(nil)
	server := httptest.NewServer(tunnelInfoMux(tm))
	defer server.Close()

	host, port := serverHostPort(t, server.URL)
	_, err := TunnelInfoForDevice("missing", host, port)
	if !errors.Is(err, ErrTunnelNotFound) {
		t.Fatalf("TunnelInfoForDevice error = %v, want ErrTunnelNotFound", err)
	}
}

func TestRefreshTunnelForDeviceWaitsForRecreatedTunnel(t *testing.T) {
	tm := tunnelManagerWithTunnels(map[string]Tunnel{
		"serial-a": {
			Udid:   "serial-a",
			closer: func() error { return nil },
		},
	})
	server := httptest.NewServer(tunnelInfoMux(tm))
	defer server.Close()

	go func() {
		time.Sleep(100 * time.Millisecond)
		tm.mux.Lock()
		tm.tunnels["serial-a"] = Tunnel{
			Udid:    "serial-a",
			Address: "fd00::new",
			RsdPort: 4321,
			closer:  func() error { return nil },
		}
		tm.mux.Unlock()
	}()

	host, port := serverHostPort(t, server.URL)
	tun, err := RefreshTunnelForDevice("serial-a", host, port, 2*time.Second)
	if err != nil {
		t.Fatalf("RefreshTunnelForDevice returned error: %v", err)
	}
	if tun.Address != "fd00::new" || tun.RsdPort != 4321 {
		t.Fatalf("refreshed tunnel = %+v, want recreated tunnel", tun)
	}
}

// TestTunnelManagerConcurrentAccess guards the TunnelManager's mutex: the
// per-device DELETE (RemoveTunnel/stopTunnel) deletes from m.tunnels while the
// UpdateTunnels loop and the tunnel-info GET handlers (ListTunnels/FindTunnel)
// read it. Run under `go test -race`: if any of those touch m.tunnels without
// the lock (the bug RemoveTunnel had before #738) the race detector trips.
func TestTunnelManagerConcurrentAccess(t *testing.T) {
	const n = 200
	tunnels := make(map[string]Tunnel, n)
	for i := 0; i < n; i++ {
		udid := fmt.Sprintf("serial-%d", i)
		tunnels[udid] = Tunnel{Udid: udid, closer: func() error { return nil }}
	}
	tm := tunnelManagerWithTunnels(tunnels)

	var wg sync.WaitGroup
	// Concurrent writers: each removes one tunnel (so removes also race removes).
	for i := 0; i < n; i++ {
		udid := fmt.Sprintf("serial-%d", i)
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = tm.RemoveTunnel(context.Background(), udid)
		}()
	}
	// Concurrent readers: list and look up while the removes happen.
	for r := 0; r < 16; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < n; j++ {
				_, _ = tm.ListTunnels()
				_, _ = tm.FindTunnel(fmt.Sprintf("serial-%d", j))
			}
		}()
	}
	wg.Wait()

	// Every tunnel was removed exactly once and the map is left consistent.
	left, err := tm.ListTunnels()
	if err != nil {
		t.Fatalf("ListTunnels: %v", err)
	}
	if len(left) != 0 {
		t.Fatalf("expected all %d tunnels removed, %d remain", n, len(left))
	}
}

func tunnelManagerWithTunnels(tunnels map[string]Tunnel) *TunnelManager {
	if tunnels == nil {
		tunnels = map[string]Tunnel{}
	}
	return &TunnelManager{
		tunnels: tunnels,
	}
}

func serverHostPort(t *testing.T, rawURL string) (string, int) {
	t.Helper()
	host, portString, err := net.SplitHostPort(rawURL[len("http://"):])
	if err != nil {
		t.Fatalf("SplitHostPort(%q): %v", rawURL, err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		t.Fatalf("Atoi(%q): %v", portString, err)
	}
	return host, port
}
