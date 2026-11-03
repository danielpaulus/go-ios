package tunnel

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
)

func TestShouldAnnounceHostOnlyFromIOS27(t *testing.T) {
	for version, want := range map[string]bool{"17.5.1": false, "26.4": false, "27.0": true, "27.2": true} {
		if got := shouldAnnounceHost(semver.MustParse(version)); got != want {
			t.Errorf("shouldAnnounceHost(%s) = %v, want %v", version, got, want)
		}
	}
}

// fakeAnnouncer fails the first failures calls and records what it was called with.
type fakeAnnouncer struct {
	failures int
	calls    int
	address  string
	port     int
	device   ios.DeviceEntry
}

func (f *fakeAnnouncer) announce(address string, port int, d ios.DeviceEntry, _ time.Duration) error {
	f.calls++
	f.address, f.port, f.device = address, port, d
	if f.calls <= f.failures {
		return errors.New("connection reset by peer")
	}
	return nil
}

func announceWith(ctx context.Context, f *fakeAnnouncer) {
	m := &TunnelManager{hostAnnouncer: f.announce}
	m.announceHost(ctx, Tunnel{Address: "fd00::1", RsdPort: 58783, Udid: "udid", UserspaceTUN: true, UserspaceTUNPort: 60106},
		ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: "udid"}})
}

func TestAnnounceHostOnTheTunnel(t *testing.T) {
	f := &fakeAnnouncer{}

	announceWith(context.Background(), f)

	if f.calls != 1 {
		t.Fatalf("announced %d times, want once", f.calls)
	}
	if f.address != "fd00::1" || f.port != 58783 {
		t.Errorf("announced on [%s]:%d, want the tunnel's RSD service", f.address, f.port)
	}
	if !f.device.UserspaceTUN || f.device.UserspaceTUNPort != 60106 {
		t.Errorf("device = %+v, want the tunnel's userspace TUN settings", f.device)
	}
}

func TestAnnounceHostRetriesUntilItSucceeds(t *testing.T) {
	f := &fakeAnnouncer{failures: 2}

	announceWith(context.Background(), f)

	if f.calls != 3 {
		t.Fatalf("announced %d times, want 3", f.calls)
	}
}

func TestAnnounceHostGivesUp(t *testing.T) {
	f := &fakeAnnouncer{failures: 100}

	announceWith(context.Background(), f)

	if f.calls != announceHostAttempts {
		t.Fatalf("announced %d times, want %d", f.calls, announceHostAttempts)
	}
}

func TestAnnounceHostStopsWithTheContext(t *testing.T) {
	f := &fakeAnnouncer{failures: 100}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	announceWith(ctx, f)

	if f.calls != 1 {
		t.Fatalf("announced %d times, want 1", f.calls)
	}
}
