package tunnel

// Unit tests for the TunnelManager robustness behaviors:
//   - liveness probing of existing tunnel records (issue #765)
//   - permanent skip of devices with unsupported iOS versions (issue #523)
//   - the udid filter restricting a manager to one device (issues #607, #479)
// All device-free: the tunnelStarter/deviceLister/tunnelProber/productVersion
// seams are faked.

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
)

type fakeStarter struct {
	mu    sync.Mutex
	calls []string
	err   error
	// rsdPort is assigned to created tunnels so tests can tell rebuilds apart.
	rsdPort int
}

func (f *fakeStarter) StartTunnel(ctx context.Context, device ios.DeviceEntry, p PairRecordManager, version *semver.Version, userspaceTUN bool) (Tunnel, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	udid := device.Properties.SerialNumber
	f.calls = append(f.calls, udid)
	if f.err != nil {
		return Tunnel{}, f.err
	}
	return Tunnel{Udid: udid, Address: "fd00::1", RsdPort: f.rsdPort, closer: func() error { return nil }}, nil
}

func (f *fakeStarter) callsFor(udid string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	n := 0
	for _, c := range f.calls {
		if c == udid {
			n++
		}
	}
	return n
}

type fakeProber struct {
	err    error
	probed []string
}

func (f *fakeProber) Probe(t Tunnel) error {
	f.probed = append(f.probed, t.Udid)
	return f.err
}

// robustnessManager builds a fully faked TunnelManager that probes on every
// UpdateTunnels cycle and resolves every device to iOS 18.0.0.
func robustnessManager(ts tunnelStarter, pr tunnelProber, entries ...ios.DeviceEntry) *TunnelManager {
	return &TunnelManager{
		ts:                 ts,
		dl:                 stubDeviceLister{list: ios.DeviceList{DeviceList: entries}},
		pr:                 pr,
		tunnels:            map[string]Tunnel{},
		failedDevices:      map[string]failedDevice{},
		unsupportedDevices: map[string]bool{},
		lastProbe:          map[string]time.Time{},
		probeInterval:      time.Nanosecond,
		startTunnelTimeout: time.Second,
		productVersion: func(ios.DeviceEntry) (*semver.Version, error) {
			return semver.MustParse("18.0.0"), nil
		},
	}
}

// A tunnel record whose device is still connected but whose probe fails must be
// torn down and rebuilt in the same update cycle — the agent-side self-heal
// from issue #765.
func TestUpdateTunnelsRebuildsDeadTunnel(t *testing.T) {
	starter := &fakeStarter{rsdPort: 4321}
	prober := &fakeProber{err: errors.New("connection timed out")}
	tm := robustnessManager(starter, prober, devEntry("dead-1", "USB"))
	var closed int
	tm.tunnels["dead-1"] = Tunnel{Udid: "dead-1", Address: "fd00::1", RsdPort: 1234, closer: func() error { closed++; return nil }}

	if err := tm.UpdateTunnels(context.Background()); err != nil {
		t.Fatalf("UpdateTunnels: %v", err)
	}

	if closed != 1 {
		t.Fatalf("dead tunnel closer called %d times, want 1", closed)
	}
	if got := starter.callsFor("dead-1"); got != 1 {
		t.Fatalf("tunnel restarted %d times, want 1", got)
	}
	rebuilt, ok := tm.tunnels["dead-1"]
	if !ok || rebuilt.RsdPort != 4321 {
		t.Fatalf("expected rebuilt tunnel with RsdPort 4321, got ok=%v tunnel=%+v", ok, rebuilt)
	}
}

// A healthy tunnel record must survive the probe untouched: no teardown, no
// restart.
func TestUpdateTunnelsKeepsHealthyTunnel(t *testing.T) {
	starter := &fakeStarter{rsdPort: 4321}
	prober := &fakeProber{}
	tm := robustnessManager(starter, prober, devEntry("ok-1", "USB"))
	var closed int
	tm.tunnels["ok-1"] = Tunnel{Udid: "ok-1", Address: "fd00::1", RsdPort: 1234, closer: func() error { closed++; return nil }}

	if err := tm.UpdateTunnels(context.Background()); err != nil {
		t.Fatalf("UpdateTunnels: %v", err)
	}

	if len(prober.probed) != 1 || prober.probed[0] != "ok-1" {
		t.Fatalf("probed = %v, want [ok-1]", prober.probed)
	}
	if closed != 0 {
		t.Fatalf("healthy tunnel closer called %d times, want 0", closed)
	}
	if got := starter.callsFor("ok-1"); got != 0 {
		t.Fatalf("healthy tunnel restarted %d times, want 0", got)
	}
	if tun := tm.tunnels["ok-1"]; tun.RsdPort != 1234 {
		t.Fatalf("healthy tunnel record changed: %+v", tun)
	}
}

// A record whose device already vanished from usbmux is not probed — the
// existing disconnect teardown owns that case.
func TestUpdateTunnelsDoesNotProbeDisconnectedDevice(t *testing.T) {
	starter := &fakeStarter{}
	prober := &fakeProber{err: errors.New("dead")}
	tm := robustnessManager(starter, prober) // no devices connected
	tm.tunnels["gone-1"] = Tunnel{Udid: "gone-1", closer: func() error { return nil }}

	if err := tm.UpdateTunnels(context.Background()); err != nil {
		t.Fatalf("UpdateTunnels: %v", err)
	}

	if len(prober.probed) != 0 {
		t.Fatalf("probed = %v, want none for a disconnected device", prober.probed)
	}
	if _, ok := tm.tunnels["gone-1"]; ok {
		t.Fatal("disconnected device's tunnel should have been torn down")
	}
}

// shouldProbe rate-limits probes to one per probeInterval and records attempts.
func TestShouldProbeRespectsInterval(t *testing.T) {
	tm := &TunnelManager{
		pr:            &fakeProber{},
		probeInterval: time.Hour,
		lastProbe:     map[string]time.Time{},
	}
	now := time.Now()
	if !tm.shouldProbe("a", now) {
		t.Fatal("first probe must be due")
	}
	if tm.shouldProbe("a", now.Add(time.Minute)) {
		t.Fatal("probe inside the interval must not be due")
	}
	if !tm.shouldProbe("a", now.Add(2*time.Hour)) {
		t.Fatal("probe after the interval must be due")
	}
	// Managers without a prober (zero value, as older tests construct) never probe.
	if (&TunnelManager{}).shouldProbe("a", now) {
		t.Fatal("manager without prober/interval must not probe")
	}
}

// The default prober cannot judge a userspace tunnel (the local forwarder
// accepts connects regardless of device state), so it must report healthy
// instead of dialing.
func TestRsdProberSkipsUserspaceTunnels(t *testing.T) {
	if err := (rsdProber{}).Probe(Tunnel{Udid: "u-1", UserspaceTUN: true}); err != nil {
		t.Fatalf("userspace tunnel probe = %v, want nil", err)
	}
}

// manualPairingTunnelStart classifies version-based failures with
// ErrUnsupportedVersion so the manager can stop retrying them (issue #523).
func TestManualPairingTunnelStartClassifiesUnsupportedVersions(t *testing.T) {
	starter := manualPairingTunnelStart{}
	_, err := starter.StartTunnel(context.Background(), ios.DeviceEntry{}, PairRecordManager{}, semver.MustParse("16.6.0"), false)
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("iOS 16 error = %v, want ErrUnsupportedVersion", err)
	}
	_, err = starter.StartTunnel(context.Background(), ios.DeviceEntry{}, PairRecordManager{}, semver.MustParse("17.2.0"), true)
	if !errors.Is(err, ErrUnsupportedVersion) {
		t.Fatalf("iOS 17.2 userspace error = %v, want ErrUnsupportedVersion", err)
	}
}

// An unsupported-version failure marks the device permanently skipped: exactly
// one start attempt, no backoff entry, and no further attempts on later cycles
// (previously this warned every single update cycle, issue #523).
func TestUpdateTunnelsSkipsUnsupportedDevicePermanently(t *testing.T) {
	starter := &fakeStarter{err: fmt.Errorf("manualPairingTunnelStart: %w 16.6.0", ErrUnsupportedVersion)}
	tm := robustnessManager(starter, &fakeProber{}, devEntry("old-1", "USB"))

	for i := 0; i < 3; i++ {
		if err := tm.UpdateTunnels(context.Background()); err != nil {
			t.Fatalf("UpdateTunnels cycle %d: %v", i, err)
		}
	}

	if got := starter.callsFor("old-1"); got != 1 {
		t.Fatalf("unsupported device attempted %d times, want exactly 1", got)
	}
	if !tm.unsupportedDevices["old-1"] {
		t.Fatal("device should be marked unsupported")
	}
	if _, ok := tm.failedDevices["old-1"]; ok {
		t.Fatal("unsupported device must not enter the transient-failure backoff")
	}
}

// A transient failure keeps the existing backoff behavior and never lands in
// the permanent unsupported set.
func TestUpdateTunnelsKeepsBackoffForTransientFailures(t *testing.T) {
	starter := &fakeStarter{err: errors.New("pairing handshake failed")}
	tm := robustnessManager(starter, &fakeProber{}, devEntry("flaky-1", "USB"))

	if err := tm.UpdateTunnels(context.Background()); err != nil {
		t.Fatalf("UpdateTunnels: %v", err)
	}

	if got, ok := tm.failedDevices["flaky-1"]; !ok || got.failCount != 1 {
		t.Fatalf("transient failure should be backed off, got ok=%v entry=%+v", ok, got)
	}
	if tm.unsupportedDevices["flaky-1"] {
		t.Fatal("transient failure must not mark the device unsupported")
	}
}

// With a udid filter set, only the matching device gets a tunnel; all others
// are ignored entirely (issues #607, #479). An empty filter manages everything.
func TestUpdateTunnelsUdidFilter(t *testing.T) {
	starter := &fakeStarter{rsdPort: 1111}
	tm := robustnessManager(starter, &fakeProber{}, devEntry("match-1", "USB"), devEntry("other-1", "USB"))
	tm.udidFilter = "match-1"

	if err := tm.UpdateTunnels(context.Background()); err != nil {
		t.Fatalf("UpdateTunnels: %v", err)
	}

	if got := starter.callsFor("match-1"); got != 1 {
		t.Fatalf("filtered device attempted %d times, want 1", got)
	}
	if got := starter.callsFor("other-1"); got != 0 {
		t.Fatalf("non-matching device attempted %d times, want 0", got)
	}
	if _, ok := tm.tunnels["match-1"]; !ok {
		t.Fatal("filtered device should have a tunnel")
	}
	if _, ok := tm.tunnels["other-1"]; ok {
		t.Fatal("non-matching device must not have a tunnel")
	}

	// Empty filter (the default) manages all devices.
	all := &fakeStarter{rsdPort: 2222}
	tmAll := robustnessManager(all, &fakeProber{}, devEntry("match-1", "USB"), devEntry("other-1", "USB"))
	if err := tmAll.UpdateTunnels(context.Background()); err != nil {
		t.Fatalf("UpdateTunnels: %v", err)
	}
	if len(tmAll.tunnels) != 2 {
		t.Fatalf("expected tunnels for both devices, got %v", tmAll.tunnels)
	}
}
