//go:build e2e

package tunnel

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
)

// TestTCPPSKTunnelE2E exercises the TLS-PSK TCP tunnel transport against a real
// device. Set TCP_PSK_UDID to the device udid. It first checks whether the
// RemotePairing network path (FindDeviceInterfaceAddress) is reachable, then
// attempts the full tunnel and prints where it gets to.
//
//	go test -tags e2e -run TestTCPPSKTunnelE2E -v ./ios/tunnel/
func TestTCPPSKTunnelE2E(t *testing.T) {
	udid := os.Getenv("TCP_PSK_UDID")
	if udid == "" {
		t.Skip("set TCP_PSK_UDID to run")
	}
	device, err := ios.GetDevice(udid)
	if err != nil {
		t.Fatalf("GetDevice(%s): %v", udid, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	addr, err := ios.FindDeviceInterfaceAddress(ctx, device)
	if err != nil {
		t.Fatalf("RemotePairing network path NOT reachable (FindDeviceInterfaceAddress): %v", err)
	}
	t.Logf("device interface address: %s", addr)

	pm, err := NewPairRecordManager("")
	if err != nil {
		t.Fatalf("NewPairRecordManager: %v", err)
	}

	tun, err := ManualPairAndConnectToTunnelTCP(ctx, device, pm)
	if err != nil {
		t.Fatalf("ManualPairAndConnectToTunnelTCP: %v", err)
	}
	defer tun.Close()
	t.Logf("TLS-PSK TCP tunnel established: address=%s rsdPort=%d", tun.Address, tun.RsdPort)
}
