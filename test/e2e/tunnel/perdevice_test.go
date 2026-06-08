//go:build e2e

package tunnel_test

import (
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

// tunnelInfo mirrors the tunnel-info HTTP API / `ios tunnel ls` JSON.
type tunnelInfo struct {
	Address          string `json:"address"`
	RsdPort          int    `json:"rsdPort"`
	Udid             string `json:"udid"`
	UserspaceTUN     bool   `json:"userspaceTun"`
	UserspaceTUNPort int    `json:"userspaceTunPort"`
}

// tunnelTransports is the transport matrix the agent supports. The kernel-TUN
// variant needs root (CAP_NET_ADMIN) and is skipped otherwise; it is also
// skipped on macOS, where go-ios sticks to userspace tunnels.
var tunnelTransports = []struct {
	name      string
	userspace bool
}{
	{"userspace", true},
	{"kernel", false},
}

// TestTunnelAgent exhaustively exercises a per-device tunnel agent over both
// transports: start + establish, ls, the --udid filter, the stop/recreate and
// refresh lifecycle, negative cases (bogus/missing udid), and HTTP shutdown.
// Each case runs its own isolated agent on a free port so it never touches a
// shared tunnel agent the host may already be running.
func TestTunnelAgent(t *testing.T) {
	for _, tr := range tunnelTransports {
		tr := tr
		t.Run(tr.name, func(t *testing.T) {
			if !tr.userspace {
				if runtime.GOOS == "darwin" {
					t.Skip("kernel TUN tunnel is not exercised on macOS — go-ios uses userspace tunnels there")
				}
				if os.Geteuid() != 0 {
					t.Skip("kernel TUN tunnel needs root (CAP_NET_ADMIN); run the suite as root to cover it")
				}
			}
			forEachDevice(t, func(t *testing.T, udid string) { perDeviceAgentSuite(t, udid, tr.userspace) })
		})
	}
}

func perDeviceAgentSuite(t *testing.T, udid string, userspace bool) {
	t.Helper()
	port := freeTunnelPort(t)
	portArg := fmt.Sprintf("--tunnel-info-port=%d", port)
	agentOut, stop := startBackground(t, udid, syscall.SIGINT, tunnelStartArgs(userspace, portArg)...)
	defer stop()

	// start + establish
	info := waitForTunnel(t, port, udid, 90*time.Second, agentOut)
	if info.UserspaceTUN != userspace {
		t.Fatalf("expected userspaceTun=%v, got %+v", userspace, info)
	}
	if userspace && info.UserspaceTUNPort == 0 {
		t.Fatalf("expected a userspace listener port, got %+v", info)
	}
	if info.RsdPort == 0 || info.Address == "" {
		t.Fatalf("tunnel missing address/rsdPort: %+v", info)
	}

	// ls + --udid filter: the agent must manage exactly this device.
	tunnels := tunnelLs(t, udid, portArg)
	if len(tunnels) != 1 || tunnels[0].Udid != udid {
		t.Fatalf("--udid filter: agent should manage exactly %s, got %+v", udid, tunnels)
	}

	// The freshly established tunnel must actually carry traffic.
	assertTunnelCarriesTraffic(t, udid, portArg)

	// negative cases — these must fail, not panic or hang.
	bogus := "00000000-000000000000000B"
	assertFails(t, "stop bogus udid", "tunnel", "stop", "--udid="+bogus, portArg)
	assertFails(t, "refresh bogus udid", "tunnel", "refresh", "--udid="+bogus, portArg)
	assertFails(t, "stop without --udid", "tunnel", "stop", portArg)

	// stop this device's tunnel, then confirm the agent recreates it.
	var stopped struct {
		Udid, Status string
	}
	if err := json.Unmarshal(smokeJSON(t, udid, "tunnel", "stop", portArg), &stopped); err != nil {
		t.Fatalf("tunnel stop: invalid JSON: %v", err)
	}
	if stopped.Status != "stopped" || stopped.Udid != udid {
		t.Fatalf("tunnel stop: expected stopped %s, got %+v", udid, stopped)
	}
	recreated := waitForTunnel(t, port, udid, 60*time.Second, agentOut)
	if recreated.Udid != udid {
		t.Fatalf("tunnel not recreated after stop: %+v", recreated)
	}

	// refresh tears it down and waits for a fresh, working tunnel.
	var refreshed tunnelInfo
	if err := json.Unmarshal(smokeJSON(t, udid, "tunnel", "refresh", portArg), &refreshed); err != nil {
		t.Fatalf("tunnel refresh: invalid JSON: %v", err)
	}
	if refreshed.Udid != udid || refreshed.RsdPort == 0 || refreshed.UserspaceTUN != userspace {
		t.Fatalf("tunnel refresh: expected a fresh %s tunnel for %s, got %+v", transportName(userspace), udid, refreshed)
	}
	// refresh must produce a genuinely NEW tunnel, not hand back the same stale
	// one (the #712 case is recovering from a stale tunnel after a reboot).
	if refreshed.Address == recreated.Address && refreshed.RsdPort == recreated.RsdPort {
		t.Fatalf("tunnel refresh: returned the same tunnel (%s rsd %d), expected a fresh one", refreshed.Address, refreshed.RsdPort)
	}
	// and the refreshed tunnel must actually work.
	assertTunnelCarriesTraffic(t, udid, portArg)
	t.Logf("per-device %s tunnel ok: refreshed %s rsd %d -> %s rsd %d",
		transportName(userspace), recreated.Address, recreated.RsdPort, refreshed.Address, refreshed.RsdPort)

	// HTTP shutdown (what `tunnel stopagent` does, but targeted at this agent's
	// port): the agent stops and its API becomes unreachable.
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/shutdown", port))
	if err == nil {
		resp.Body.Close()
	}
	if !waitUntil(10*time.Second, func() bool {
		_, e := http.Get(fmt.Sprintf("http://127.0.0.1:%d/tunnels", port))
		return e != nil // connection refused once the agent is gone
	}) {
		t.Fatalf("agent did not shut down after /shutdown on port %d", port)
	}
}

// TestTunnelAgentAllDevices verifies the non-filtered agent: started without
// --udid it tunnels every connected (tunnel-capable) device, and stopping one
// device's tunnel leaves the agent — and the other devices — running.
func TestTunnelAgentAllDevices(t *testing.T) {
	port := freeTunnelPort(t)
	portArg := fmt.Sprintf("--tunnel-info-port=%d", port)
	// Empty udid => no --udid filter => manage all devices.
	agentOut, stop := startBackground(t, "", syscall.SIGINT, "tunnel", "start", "--userspace", portArg)
	defer stop()

	tunnels := waitForAnyTunnel(t, port, 90*time.Second, agentOut)
	t.Logf("non-filtered agent established %d tunnel(s)", len(tunnels))
	target := tunnels[0].Udid

	// The non-filtered agent's tunnel must actually carry traffic.
	assertTunnelCarriesTraffic(t, target, portArg)

	// Stop one device's tunnel; the agent must stay up and keep serving.
	var stopped struct{ Status string }
	if err := json.Unmarshal(smokeJSON(t, target, "tunnel", "stop", portArg), &stopped); err != nil || stopped.Status != "stopped" {
		t.Fatalf("tunnel stop %s: status=%q err=%v", target, stopped.Status, err)
	}
	if _, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/tunnels", port)); err != nil {
		t.Fatalf("agent went down after a per-device stop (should stay up): %v\n%s", err, agentOut())
	}
}

// TestTunnelAgentMixed runs a general (all-devices) agent and a per-device agent
// at the same time on different tunnel-info ports and proves they don't conflict:
// they bind cleanly, each gets its own independent tunnel to the device with its
// own userspace listener port (derived from its own tunnel-info port), and
// stopping a device's tunnel on one agent never affects the other agent's tunnel.
// People run go-ios like this in production, so it is verified, not assumed.
func TestTunnelAgentMixed(t *testing.T) {
	devs := harness.Devices()
	if len(devs) == 0 {
		t.Skip("GO_IOS_E2E_DEVICES not set")
	}
	udid := devs[0]

	genPort := freeTunnelPort(t)
	perPort := freeTunnelPort(t)
	genArg := fmt.Sprintf("--tunnel-info-port=%d", genPort)
	perArg := fmt.Sprintf("--tunnel-info-port=%d", perPort)

	// General agent: empty udid => manage all devices.
	genOut, stopGen := startBackground(t, "", syscall.SIGINT, "tunnel", "start", "--userspace", genArg)
	defer stopGen()
	// Per-device agent: only this device.
	perOut, stopPer := startBackground(t, udid, syscall.SIGINT, "tunnel", "start", "--userspace", perArg)
	defer stopPer()

	gen := waitForTunnel(t, genPort, udid, 90*time.Second, genOut)
	per := waitForTunnel(t, perPort, udid, 90*time.Second, perOut)

	// Independent tunnels: distinct device addresses and distinct listener ports.
	if gen.Address == per.Address {
		t.Fatalf("expected two independent tunnels, both have address %s", gen.Address)
	}
	if gen.UserspaceTUNPort == per.UserspaceTUNPort {
		t.Fatalf("userspace listener port collision: both agents on %d", gen.UserspaceTUNPort)
	}
	// Each agent's userspace port is derived from its OWN tunnel-info port.
	if gen.UserspaceTUNPort <= genPort || gen.UserspaceTUNPort > genPort+1000 {
		t.Fatalf("general agent userspace port %d not derived from its tunnel-info port %d", gen.UserspaceTUNPort, genPort)
	}
	if per.UserspaceTUNPort != perPort+1 {
		t.Fatalf("per-device agent userspace port = %d, want %d", per.UserspaceTUNPort, perPort+1)
	}

	// Both tunnels to the same device must actually carry traffic simultaneously.
	assertTunnelCarriesTraffic(t, udid, genArg)
	assertTunnelCarriesTraffic(t, udid, perArg)

	// Isolation 1: stopping the device on the general agent must not touch the
	// per-device agent's tunnel.
	smokeJSON(t, udid, "tunnel", "stop", genArg)
	if still := waitForTunnel(t, perPort, udid, 5*time.Second, perOut); still.Address != per.Address {
		t.Fatalf("per-device tunnel changed after stopping the device on the general agent: %s -> %s", per.Address, still.Address)
	}

	// Isolation 2: the reverse. The general agent recreated its own tunnel after
	// the stop above; capture it, then stop on the per-device agent and confirm
	// the general agent's tunnel is untouched.
	genNow := waitForTunnel(t, genPort, udid, 30*time.Second, genOut)
	smokeJSON(t, udid, "tunnel", "stop", perArg)
	if after := waitForTunnel(t, genPort, udid, 5*time.Second, genOut); after.Address != genNow.Address {
		t.Fatalf("general tunnel changed after stopping the device on the per-device agent: %s -> %s", genNow.Address, after.Address)
	}
	t.Logf("mixed agents ok: general %s (uPort %d) + per-device %s (uPort %d) coexist and are isolated",
		gen.Address, gen.UserspaceTUNPort, per.Address, per.UserspaceTUNPort)
}

func tunnelStartArgs(userspace bool, portArg string) []string {
	if userspace {
		return []string{"tunnel", "start", "--userspace", portArg}
	}
	return []string{"tunnel", "start", portArg}
}

func transportName(userspace bool) string {
	if userspace {
		return "userspace"
	}
	return "kernel"
}

// assertFails runs an ios command (raw, no --udid appended) and fails the test
// unless the command exits non-zero.
func assertFails(t *testing.T, what string, args ...string) {
	t.Helper()
	stdout, stderr, err := harness.TryRun(t, args...)
	if err == nil {
		t.Fatalf("%s: expected failure, but command succeeded\nstdout: %s\nstderr: %s", what, stdout, stderr)
	}
}

// assertTunnelCarriesTraffic proves the agent's tunnel actually works — that it
// carries RemoteXPC traffic — not merely that the tunnel-info API lists it. A
// zombie tunnel still lists fine while its data plane is dead, so the only
// honest check is to drive a real developer-service command through it: capture
// a screenshot through this agent's tunnel and verify it decodes as a PNG.
func assertTunnelCarriesTraffic(t *testing.T, udid, portArg string) {
	t.Helper()
	out := filepath.Join(t.TempDir(), "shot.png")
	runIOSForDevice(t, udid, "screenshot", "--output="+out, portArg)
	f, err := os.Open(out)
	if err != nil {
		t.Fatalf("screenshot through tunnel: output not created: %v", err)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatalf("screenshot through tunnel: not a valid PNG (tunnel data plane likely broken): %v", err)
	}
	if b := img.Bounds(); b.Dx() <= 0 || b.Dy() <= 0 {
		t.Fatalf("screenshot through tunnel: invalid dimensions %dx%d", b.Dx(), b.Dy())
	}
}

// tunnelLs runs `ios tunnel ls` against the given agent and returns the tunnels.
func tunnelLs(t *testing.T, udid string, portArg string) []tunnelInfo {
	t.Helper()
	out := runIOSForDevice(t, udid, "tunnel", "ls", portArg)
	var tunnels []tunnelInfo
	if err := json.Unmarshal(out, &tunnels); err != nil {
		t.Fatalf("tunnel ls: invalid JSON %q: %v", string(out), err)
	}
	return tunnels
}

// waitForTunnel polls the agent's tunnel-info API until a tunnel for udid exists.
func waitForTunnel(t *testing.T, port int, udid string, timeout time.Duration, agentOut func() string) tunnelInfo {
	t.Helper()
	url := fmt.Sprintf("http://127.0.0.1:%d/tunnel/%s", port, udid)
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(url)
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				var info tunnelInfo
				if json.Unmarshal(body, &info) == nil && info.Udid == udid {
					return info
				}
			}
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("tunnel for %s not ready within %s on port %d; agent output:\n%s", udid, timeout, port, agentOut())
	return tunnelInfo{}
}

// waitForAnyTunnel polls /tunnels until the agent reports at least one tunnel.
func waitForAnyTunnel(t *testing.T, port int, timeout time.Duration, agentOut func() string) []tunnelInfo {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/tunnels", port))
		if err == nil {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			var tunnels []tunnelInfo
			if resp.StatusCode == http.StatusOK && json.Unmarshal(body, &tunnels) == nil && len(tunnels) > 0 {
				return tunnels
			}
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatalf("no tunnels within %s on port %d; agent output:\n%s", timeout, port, agentOut())
	return nil
}

func waitUntil(timeout time.Duration, cond func() bool) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if cond() {
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false
}

// freeTunnelPort returns a free TCP port whose successor (used by the userspace
// listener) is also free, so the per-device agent can bind both.
func freeTunnelPort(t *testing.T) int {
	t.Helper()
	for attempts := 0; attempts < 30; attempts++ {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			continue
		}
		port := l.Addr().(*net.TCPAddr).Port
		l.Close()
		next, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port+1))
		if err != nil {
			continue
		}
		next.Close()
		return port
	}
	t.Fatal("could not find a free consecutive port pair")
	return 0
}
