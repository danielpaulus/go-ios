package tunnel

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

var netClient = &http.Client{
	Timeout: time.Millisecond * 200,
}

var ErrTunnelNotFound = errors.New("tunnel not found")

// ErrUnsupportedVersion is returned when a device's iOS version can never use
// go-ios managed tunnels (tunnels only exist for iOS 17+). The TunnelManager
// treats it as permanent and stops retrying such devices.
var ErrUnsupportedVersion = errors.New("unsupported iOS version")

func CloseAgent() error {
	_, err := netClient.Get(fmt.Sprintf("http://%s:%d/shutdown", ios.HttpApiHost(), ios.HttpApiPort()))
	if err != nil {
		return fmt.Errorf("CloseAgent: failed to send shutdown request: %w", err)
	}
	return nil
}

func IsAgentRunning() bool {
	resp, err := netClient.Get(fmt.Sprintf("http://%s:%d/health", ios.HttpApiHost(), ios.HttpApiPort()))
	if err != nil {
		return false
	}
	return resp.StatusCode == http.StatusOK
}
func WaitUntilAgentReady() bool {
	for {
		golog.Info("Waiting for go-ios agent to be ready...", "module", logModule)
		resp, err := netClient.Get(fmt.Sprintf("http://%s:%d/ready", ios.HttpApiHost(), ios.HttpApiPort()))
		if err != nil {
			return false
		}
		ready := resp.StatusCode == http.StatusOK
		resp.Body.Close()
		if ready {
			golog.Info("Go-iOS Agent is ready", "module", logModule)
			return true
		}
		// Not ready yet (agent up but first tunnel update not finished): back off
		// a little instead of hammering /ready in a tight loop, and close the body
		// each iteration so responses don't leak.
		time.Sleep(200 * time.Millisecond)
	}
}

func RunAgent(mode string, args ...string) error {
	if IsAgentRunning() {
		return nil
	}
	golog.Info("Go-iOS Agent not running, starting it on port", "module", logModule, "port", ios.HttpApiPort())
	ex, err := os.Executable()
	if err != nil {
		return fmt.Errorf("RunAgent: failed to get executable path: %w", err)
	}

	var cmd *exec.Cmd
	switch mode {
	case "kernel":
		cmd = exec.Command(ex, append([]string{"tunnel", "start"}, args...)...)
	case "user":
		cmd = exec.Command(ex, append([]string{"tunnel", "start", "--userspace"}, args...)...)
	default:
		return fmt.Errorf("RunAgent: unknown mode: %s. Only 'kernel' and 'user' are supported", mode)
	}

	// OS specific SysProcAttr assignment
	cmd.SysProcAttr = createSysProcAttr()

	err = cmd.Start()

	if err != nil {
		return fmt.Errorf("RunAgent: failed to start agent: %w", err)
	}
	err = cmd.Process.Release()
	if err != nil {
		return fmt.Errorf("RunAgent: failed to release process: %w", err)
	}
	WaitUntilAgentReady()
	return nil
}

// ServeTunnelInfo starts a simple http serve that exposes the tunnel information about the running tunnel.
// The API has two endpoints:
// 1. GET    {HOST}:{PORT}/tunnel/{UDID} to get the tunnel info for a specific device
// 2. DELETE {HOST}:{PORT}/tunnel/{UDID} to stop a device tunnel
// 3. GET    {HOST}:{PORT}/tunnels       to get a list of all tunnels
// The host defaults to 127.0.0.1 and can be changed with --tunnel-info-host / GO_IOS_AGENT_HOST.
func ServeTunnelInfo(tm *TunnelManager, host string, port int) error {
	if err := http.ListenAndServe(net.JoinHostPort(host, fmt.Sprintf("%d", port)), tunnelInfoMux(tm)); err != nil {
		return fmt.Errorf("ServeTunnelInfo: failed to start http server: %w", err)
	}
	return nil
}

func tunnelInfoMux(tm *TunnelManager) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		writer.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/ready", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if tm.FirstUpdateCompleted() {
			writer.WriteHeader(http.StatusOK)
		} else {
			writer.WriteHeader(http.StatusServiceUnavailable)
		}
	})
	mux.HandleFunc("/shutdown", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		err := tm.Close()
		if err != nil {
			golog.Error("failed to close tunnel manager", "module", logModule, "error", err)
		}
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("shutting down in 1 second..."))
		go func() {
			time.Sleep(1 * time.Second)
			os.Exit(0)
		}()
	})
	mux.HandleFunc("/tunnel/", func(writer http.ResponseWriter, request *http.Request) {
		udid := strings.TrimPrefix(request.URL.Path, "/tunnel/")
		if len(udid) == 0 {
			http.Error(writer, "missing udid", http.StatusBadRequest)
			return
		}

		if request.Method == "GET" {
			t, err := tm.FindTunnel(udid)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			if len(t.Udid) == 0 {
				http.Error(writer, ErrTunnelNotFound.Error(), http.StatusNotFound)
				return
			}
			writer.Header().Add("Content-Type", "application/json")
			// The header/status are already committed, so an encode failure here
			// can only be logged, not turned into an error response.
			if err := json.NewEncoder(writer).Encode(t); err != nil {
				golog.Error("failed to encode tunnel info", "module", logModule, "udid", udid, "error", err)
			}
		} else if request.Method == "DELETE" {
			err := tm.RemoveTunnel(request.Context(), udid)
			if errors.Is(err, ErrTunnelNotFound) {
				http.Error(writer, err.Error(), http.StatusNotFound)
				return
			}
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}
			writer.WriteHeader(http.StatusNoContent)
		} else {
			writer.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/tunnels", func(writer http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			writer.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		tunnels, err := tm.ListTunnels()
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}

		writer.Header().Add("Content-Type", "application/json")
		enc := json.NewEncoder(writer)
		err = enc.Encode(tunnels)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	return mux
}

func TunnelInfoForDevice(udid string, tunnelInfoHost string, tunnelInfoPort int) (Tunnel, error) {
	c := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := c.Get(fmt.Sprintf("http://%s/tunnel/%s", net.JoinHostPort(tunnelInfoHost, fmt.Sprintf("%d", tunnelInfoPort)), udid))
	if err != nil {
		return Tunnel{}, fmt.Errorf("TunnelInfoForDevice: failed to get tunnel info: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Tunnel{}, fmt.Errorf("TunnelInfoForDevice: failed to read body: %w", err)
	}
	if res.StatusCode == http.StatusNotFound {
		return Tunnel{}, ErrTunnelNotFound
	}
	if res.StatusCode != http.StatusOK {
		return Tunnel{}, fmt.Errorf("TunnelInfoForDevice: unexpected status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	var info Tunnel
	err = json.Unmarshal(body, &info)
	if err != nil {
		return Tunnel{}, fmt.Errorf("TunnelInfoForDevice: failed to parse response: %w", err)
	}
	return info, nil
}

func ListRunningTunnels(tunnelInfoHost string, tunnelInfoPort int) ([]Tunnel, error) {
	c := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := c.Get(fmt.Sprintf("http://%s/tunnels", net.JoinHostPort(tunnelInfoHost, fmt.Sprintf("%d", tunnelInfoPort))))
	if err != nil {
		return nil, fmt.Errorf("TunnelInfoForDevice: failed to get tunnel info: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("TunnelInfoForDevice: failed to read body: %w", err)
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("TunnelInfoForDevice: unexpected status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	var info []Tunnel
	err = json.Unmarshal(body, &info)
	if err != nil {
		return nil, fmt.Errorf("TunnelInfoForDevice: failed to parse response: %w", err)
	}
	return info, nil
}

func StopTunnelForDevice(udid string, tunnelInfoHost string, tunnelInfoPort int) error {
	if udid == "" {
		return fmt.Errorf("StopTunnelForDevice: udid is required")
	}
	c := http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("http://%s/tunnel/%s", net.JoinHostPort(tunnelInfoHost, fmt.Sprintf("%d", tunnelInfoPort)), udid), nil)
	if err != nil {
		return fmt.Errorf("StopTunnelForDevice: failed to create request: %w", err)
	}
	res, err := c.Do(req)
	if err != nil {
		return fmt.Errorf("StopTunnelForDevice: failed to stop tunnel: %w", err)
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("StopTunnelForDevice: failed to read body: %w", err)
	}
	if res.StatusCode == http.StatusNotFound {
		return ErrTunnelNotFound
	}
	if res.StatusCode != http.StatusNoContent {
		return fmt.Errorf("StopTunnelForDevice: unexpected status %d: %s", res.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func RefreshTunnelForDevice(udid string, tunnelInfoHost string, tunnelInfoPort int, timeout time.Duration) (Tunnel, error) {
	if err := StopTunnelForDevice(udid, tunnelInfoHost, tunnelInfoPort); err != nil && !errors.Is(err, ErrTunnelNotFound) {
		return Tunnel{}, err
	}
	deadline := time.Now().Add(timeout)
	for {
		t, err := TunnelInfoForDevice(udid, tunnelInfoHost, tunnelInfoPort)
		if err == nil {
			return t, nil
		}
		if !errors.Is(err, ErrTunnelNotFound) {
			return Tunnel{}, err
		}
		if time.Now().After(deadline) {
			return Tunnel{}, fmt.Errorf("RefreshTunnelForDevice: timed out waiting for tunnel %s", udid)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// TunnelManager starts tunnels for devices when needed (if no tunnel is running yet) and stores the information
// how those tunnels are reachable (address and remote service discovery port)
// failedDevice tracks a device whose tunnel failed to start, so retries can be
// backed off instead of attempted every UpdateTunnels cycle (each attempt opens
// a usbmux socket, so hammering a device that always fails leaks sockets).
type failedDevice struct {
	lastAttempt time.Time
	failCount   int
}

type TunnelManager struct {
	ts      tunnelStarter
	dl      deviceLister
	pr      tunnelProber
	pm      PairRecordManager
	mux     sync.Mutex
	tunnels map[string]Tunnel
	// failedDevices tracks devices whose tunnel start failed (keyed by udid) so
	// UpdateTunnels can back off before retrying them.
	failedDevices map[string]failedDevice
	// unsupportedDevices tracks devices whose iOS version can never tunnel
	// (ErrUnsupportedVersion). They are skipped for the rest of the process so
	// the same warning is not logged every update cycle.
	unsupportedDevices map[string]bool
	// lastProbe tracks when each tunnel was last liveness-probed so probes run
	// at most once per probeInterval.
	lastProbe map[string]time.Time
	// probeFailures counts consecutive failed liveness probes per udid. A tunnel
	// is only torn down after probeFailureThreshold consecutive failures so a
	// single transient probe error (a momentary route hiccup or load spike)
	// doesn't destroy an otherwise healthy tunnel. A successful probe resets it.
	probeFailures        map[string]int
	probeInterval        time.Duration
	startTunnelTimeout   time.Duration
	firstUpdateCompleted bool
	userspaceTUN         bool
	closeOnce            sync.Once
	portOffset           int
	// productVersion resolves a device's iOS version; a seam so tests can run
	// without a device. Defaults to ios.GetProductVersion.
	productVersion func(ios.DeviceEntry) (*semver.Version, error)
	// udidFilter, when non-empty, restricts the manager to a single device so
	// you can run one isolated tunnel agent per device.
	udidFilter string
	// basePort is the base for derived userspace listener ports (the agent's own
	// tunnel-info port), so per-device agents on different ports don't collide.
	basePort int
}

// NewTunnelManager creates a new TunnelManager instance for setting up device tunnels for all connected devices
// If userspaceTUN is set to true, the network stack will run in user space.
func NewTunnelManager(pm PairRecordManager, userspaceTUN bool) *TunnelManager {
	return newTunnelManager(pm, userspaceTUN, "", ios.HttpApiPort())
}

// NewTunnelManagerForDevice creates a TunnelManager bound to a specific
// tunnel-info port. If udid is non-empty the manager only manages that one
// device (ignoring all others), so you can run one isolated agent per device; an
// empty udid manages all connected devices. basePort is the agent's tunnel-info
// port: userspace listener ports are derived from it so multiple agents on
// different ports (e.g. a general agent plus per-device agents) never clash.
func NewTunnelManagerForDevice(pm PairRecordManager, userspaceTUN bool, udid string, basePort int) *TunnelManager {
	return newTunnelManager(pm, userspaceTUN, udid, basePort)
}

func newTunnelManager(pm PairRecordManager, userspaceTUN bool, udidFilter string, basePort int) *TunnelManager {
	if basePort == 0 {
		basePort = ios.HttpApiPort()
	}
	return &TunnelManager{
		ts:                 manualPairingTunnelStart{},
		dl:                 deviceList{},
		pr:                 rsdProber{},
		pm:                 pm,
		tunnels:            map[string]Tunnel{},
		failedDevices:      map[string]failedDevice{},
		unsupportedDevices: map[string]bool{},
		lastProbe:          map[string]time.Time{},
		probeFailures:      map[string]int{},
		probeInterval:      defaultProbeInterval,
		startTunnelTimeout: 10 * time.Second,
		userspaceTUN:       userspaceTUN,
		udidFilter:         udidFilter,
		basePort:           basePort,
		portOffset:         1,
		productVersion:     ios.GetProductVersion,
	}
}

func (m *TunnelManager) Close() error {
	var baseErr error
	m.closeOnce.Do(func() {
		tunnels, err := m.ListTunnels()
		if err != nil {
			golog.Error("failed to list tunnels", "module", logModule, "error", err)
		}
		for _, t := range tunnels {
			err := t.Close()
			baseErr = errors.Join(baseErr, err)
			if err != nil {
				golog.Error("failed to stop tunnel", "module", logModule, "udid", t.Udid, "error", err)
			}
		}
	})
	return baseErr
}

// FirstUpdateCompleted returns true if the first update completed,
// use it to prevent race conditions when trying to use go-ios agent for the first time
func (m *TunnelManager) FirstUpdateCompleted() bool {
	m.mux.Lock()
	defer m.mux.Unlock()
	return m.firstUpdateCompleted
}

// UpdateTunnels checks for connected devices and starts a new tunnel if needed
// On device disconnects the tunnel resources get cleaned up
func (m *TunnelManager) UpdateTunnels(ctx context.Context) error {

	m.mux.Lock()
	localTunnels := map[string]Tunnel{}
	maps.Copy(localTunnels, m.tunnels)
	localFailed := map[string]failedDevice{}
	maps.Copy(localFailed, m.failedDevices)
	localUnsupported := map[string]bool{}
	maps.Copy(localUnsupported, m.unsupportedDevices)
	m.mux.Unlock()

	devices, err := m.dl.ListDevices()
	if err != nil {
		// ListDevices can fail transiently while a device disconnects/reconnects
		// (notably on Linux). Tear down the local tunnels so a stale one can't
		// survive and block the device's next connection; they get recreated on
		// the next successful update.
		for udid, tun := range localTunnels {
			golog.Info("stopping tunnel due to device list error", "module", logModule, "udid", udid)
			_ = m.stopTunnel(tun)
		}
		return fmt.Errorf("UpdateTunnels: failed to get list of devices: %w", err)
	}

	// currentUDIDs holds every connected device (built before the udidFilter
	// check) so stale failedDevices entries for now-disconnected devices can be
	// pruned below, letting a reconnect retry immediately.
	currentUDIDs := make(map[string]bool, len(devices.DeviceList))
	for _, d := range devices.DeviceList {
		currentUDIDs[d.Properties.SerialNumber] = true
	}

	// Liveness-probe existing tunnel records of still-connected devices: a
	// tunnel can die without a usbmux disconnect (quick reboot, transport
	// death), leaving a stale record that would otherwise be served forever. A
	// tunnel that fails probeFailureThreshold consecutive probes is torn down
	// here so the create loop below rebuilds it in the same cycle. Records of
	// vanished devices are handled by the disconnect teardown at the end.
	//
	// Probes run concurrently (bounded) because a dead-but-still-routed tunnel
	// blackholes the SYN and burns the full probeDialTimeout; probing serially
	// would let N dead tunnels delay tunnel creation for newly connected devices
	// by up to probeDialTimeout*N.
	now := time.Now()
	for udid := range m.probeTunnels(localTunnels, currentUDIDs, now) {
		_ = m.stopTunnel(localTunnels[udid])
		delete(localTunnels, udid)
	}

	for _, d := range devices.DeviceList {
		udid := d.Properties.SerialNumber
		if m.udidFilter != "" && udid != m.udidFilter {
			continue
		}
		if localUnsupported[udid] {
			continue
		}
		if _, exists := localTunnels[udid]; exists {
			continue
		}
		// Skip network devices (they can't tunnel) and devices still inside their
		// failure backoff window. Either way, attempting a tunnel here would open
		// a usbmux socket (via GetProductVersion) that, for a device that always
		// fails, accumulates as a leaked socket every cycle.
		if shouldSkipDevice(d, localFailed, time.Now()) {
			continue
		}
		if m.userspaceTUN && d.UserspaceTUNPort == 0 {
			m.mux.Lock()
			d.UserspaceTUNPort = m.basePort + m.portOffset
			m.portOffset++
			m.mux.Unlock()
		}
		t, err := m.startTunnel(ctx, d)
		if err != nil {
			if errors.Is(err, ErrUnsupportedVersion) {
				// The device can never tunnel on its current iOS version, so
				// retrying would only repeat the same warning every cycle. Log
				// once at info and skip the device for the rest of the process.
				golog.Info("device iOS version does not support tunnels, skipping it from now on", "module", logModule, "udid", udid, "error", err)
				m.mux.Lock()
				m.unsupportedDevices[udid] = true
				m.mux.Unlock()
				continue
			}
			golog.Warn("failed to start tunnel", "module", logModule, "udid", udid, "error", err)
			m.mux.Lock()
			m.failedDevices[udid] = failedDevice{lastAttempt: time.Now(), failCount: m.failedDevices[udid].failCount + 1}
			m.mux.Unlock()
			continue
		}
		m.mux.Lock()
		delete(m.failedDevices, udid)
		if m.lastProbe != nil {
			// A fresh tunnel is known-alive; defer its first probe by a full interval.
			m.lastProbe[udid] = time.Now()
		}
		localTunnels[udid] = t
		m.tunnels[udid] = t
		m.mux.Unlock()
	}
	for udid, tun := range localTunnels {
		stillConnected := slices.ContainsFunc(devices.DeviceList, func(entry ios.DeviceEntry) bool {
			return entry.Properties.SerialNumber == udid
		})
		if !stillConnected {
			_ = m.stopTunnel(tun)
		}
	}
	m.mux.Lock()
	for udid := range m.failedDevices {
		if !currentUDIDs[udid] {
			delete(m.failedDevices, udid)
		}
	}
	// Prune per-device bookkeeping for devices that are no longer connected so
	// these maps stay bounded by the current device count instead of growing
	// with every udid ever seen. Pruning unsupportedDevices on disconnect also
	// lets a device that was unsupported, then upgraded to a tunnel-capable iOS
	// and reconnected, be retried instead of skipped for the rest of the process.
	for udid := range m.unsupportedDevices {
		if !currentUDIDs[udid] {
			delete(m.unsupportedDevices, udid)
		}
	}
	for udid := range m.probeFailures {
		if !currentUDIDs[udid] {
			delete(m.probeFailures, udid)
		}
	}
	m.firstUpdateCompleted = true
	m.mux.Unlock()
	return nil
}

// shouldSkipDevice reports whether UpdateTunnels should not attempt a tunnel for
// d on this cycle: network-connected devices can never establish a tunnel, and a
// device that recently failed is held off until its backoff window elapses.
func shouldSkipDevice(d ios.DeviceEntry, failed map[string]failedDevice, now time.Time) bool {
	if d.Properties.ConnectionType == "Network" {
		return true
	}
	if f, ok := failed[d.Properties.SerialNumber]; ok && now.Sub(f.lastAttempt) < failedDeviceBackoff(f.failCount) {
		return true
	}
	return false
}

// shouldProbe reports whether the tunnel for udid is due for a liveness probe
// and, if so, records the attempt so probes run at most once per probeInterval.
func (m *TunnelManager) shouldProbe(udid string, now time.Time) bool {
	if m.pr == nil || m.probeInterval <= 0 {
		return false
	}
	m.mux.Lock()
	defer m.mux.Unlock()
	if last, ok := m.lastProbe[udid]; ok && now.Sub(last) < m.probeInterval {
		return false
	}
	m.lastProbe[udid] = now
	return true
}

// probeTunnels liveness-probes every due tunnel of a still-connected device and
// returns the set of udids that should be torn down. Probes run concurrently
// with at most maxConcurrentProbes in flight so a batch of dead-but-routed
// tunnels (each of which blackholes the SYN for the full probeDialTimeout)
// can't serialize into a multi-second stall of the update loop. A tunnel is only
// reported dead after probeFailureThreshold consecutive failures; a success
// resets its counter, so a single transient probe error is absorbed.
func (m *TunnelManager) probeTunnels(localTunnels map[string]Tunnel, currentUDIDs map[string]bool, now time.Time) map[string]bool {
	type probeResult struct {
		udid string
		err  error
	}
	results := make(chan probeResult)
	sem := make(chan struct{}, maxConcurrentProbes)
	var wg sync.WaitGroup
	for udid, tun := range localTunnels {
		if !currentUDIDs[udid] || !m.shouldProbe(udid, now) {
			continue
		}
		wg.Add(1)
		go func(udid string, tun Tunnel) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results <- probeResult{udid: udid, err: m.pr.Probe(tun)}
		}(udid, tun)
	}
	go func() {
		wg.Wait()
		close(results)
	}()

	dead := map[string]bool{}
	m.mux.Lock()
	defer m.mux.Unlock()
	for r := range results {
		if r.err == nil {
			delete(m.probeFailures, r.udid)
			continue
		}
		m.probeFailures[r.udid]++
		fails := m.probeFailures[r.udid]
		if fails < probeFailureThreshold {
			golog.Warn("tunnel failed liveness probe, will retry before restarting it",
				"module", logModule, "udid", r.udid, "consecutiveFailures", fails, "threshold", probeFailureThreshold, "error", r.err)
			continue
		}
		golog.Warn("tunnel failed liveness probe repeatedly, restarting it",
			"module", logModule, "udid", r.udid, "consecutiveFailures", fails, "error", r.err)
		delete(m.probeFailures, r.udid)
		dead[r.udid] = true
	}
	return dead
}

// failedDeviceBackoff returns how long to wait before retrying a device after
// failCount consecutive failures: 30s, 60s, 120s, 240s, capped at 5 minutes.
func failedDeviceBackoff(failCount int) time.Duration {
	shift := failCount - 1
	if shift < 0 {
		shift = 0
	}
	if shift > 4 {
		shift = 4
	}
	seconds := 30 * (1 << shift)
	if seconds > 300 {
		seconds = 300
	}
	return time.Duration(seconds) * time.Second
}

func (m *TunnelManager) RemoveTunnel(ctx context.Context, serialNumber string) error {
	m.mux.Lock()
	tun, exists := m.tunnels[serialNumber]
	if exists {
		delete(m.tunnels, serialNumber)
	}
	delete(m.lastProbe, serialNumber)
	delete(m.probeFailures, serialNumber)
	m.mux.Unlock()

	if !exists {
		return ErrTunnelNotFound
	}
	golog.Info("stopping tunnel", "module", logModule, "udid", tun.Udid)
	return tun.Close()
}

func (m *TunnelManager) stopTunnel(t Tunnel) error {
	m.mux.Lock()
	golog.Info("stopping tunnel", "module", logModule, "udid", t.Udid)
	delete(m.tunnels, t.Udid)
	delete(m.lastProbe, t.Udid)
	delete(m.probeFailures, t.Udid)
	m.mux.Unlock()

	return t.Close()
}

func (m *TunnelManager) startTunnel(ctx context.Context, device ios.DeviceEntry) (Tunnel, error) {
	golog.Info("start tunnel", "module", logModule, "udid", device.Properties.SerialNumber)
	startTunnelCtx, cancel := context.WithTimeout(ctx, m.startTunnelTimeout)
	defer cancel()
	versionOf := m.productVersion
	if versionOf == nil {
		versionOf = ios.GetProductVersion
	}
	version, err := versionOf(device)
	if err != nil {
		return Tunnel{}, fmt.Errorf("startTunnel: failed to get device version: %w", err)
	}
	t, err := m.ts.StartTunnel(startTunnelCtx, device, m.pm, version, m.userspaceTUN)
	if err != nil {
		return Tunnel{}, err
	}
	return t, nil
}

// ListTunnels provides all currently running device tunnels
func (m *TunnelManager) ListTunnels() ([]Tunnel, error) {
	m.mux.Lock()
	defer m.mux.Unlock()
	return maps.Values(m.tunnels), nil
}

func (m *TunnelManager) FindTunnel(udid string) (Tunnel, error) {
	tunnels, err := m.ListTunnels()
	if err != nil {
		return Tunnel{}, err
	}

	for _, t := range tunnels {
		if t.Udid == udid {
			return t, nil
		}
	}

	return Tunnel{}, nil
}

type tunnelStarter interface {
	StartTunnel(ctx context.Context, device ios.DeviceEntry, p PairRecordManager, version *semver.Version, userspaceTUN bool) (Tunnel, error)
}

type deviceLister interface {
	ListDevices() (ios.DeviceList, error)
}

// tunnelProber checks whether an existing tunnel record still reaches the
// device, so the TunnelManager can tear down and rebuild dead tunnels.
type tunnelProber interface {
	Probe(t Tunnel) error
}

// defaultProbeInterval is how often the TunnelManager liveness-probes each
// tunnel record. Combined with the probe's short dial timeout and the
// probeFailureThreshold consecutive-failure requirement, a persistently dead
// tunnel is detected and rebuilt within roughly a minute instead of never,
// while a single transient probe error never tears down a healthy tunnel.
const defaultProbeInterval = 30 * time.Second

// probeDialTimeout bounds the liveness probe's TCP connect. A healthy tunnel
// answers in milliseconds; a dead-but-still-routed one swallows the SYN, so a
// short timeout is enough to tell them apart without stalling UpdateTunnels.
const probeDialTimeout = 5 * time.Second

// maxConcurrentProbes caps how many liveness probes run at once. Probes are
// concurrent so a batch of dead-but-routed tunnels can't serialize into a
// probeDialTimeout*N stall, but bounded so a large fleet can't open an
// unbounded number of dial sockets in one cycle.
const maxConcurrentProbes = 8

// probeFailureThreshold is the number of consecutive failed liveness probes
// required before a tunnel is torn down and rebuilt. Requiring two failures
// keeps a single transient probe error (a momentary route hiccup) from
// destroying a healthy tunnel.
const probeFailureThreshold = 2

// rsdProber is the default tunnelProber: a short-timeout TCP dial of the
// tunnel's RSD endpoint over the kernel TUN route.
type rsdProber struct {
}

func (rsdProber) Probe(t Tunnel) error {
	if t.UserspaceTUN {
		// The userspace forwarder listens on localhost and accepts connects no
		// matter what state the device is in, so a dial proves nothing here.
		return nil
	}
	conn, err := ios.DialTunnelTCPWithTimeout(fmt.Sprintf("[%s]:%d", t.Address, t.RsdPort), probeDialTimeout)
	if err != nil {
		return err
	}
	return conn.Close()
}

type manualPairingTunnelStart struct {
}

func (m manualPairingTunnelStart) StartTunnel(ctx context.Context, device ios.DeviceEntry, p PairRecordManager, version *semver.Version, userspaceTUN bool) (Tunnel, error) {

	if version.GreaterThan(semver.MustParse("17.4.0")) {
		if userspaceTUN {
			tun, err := ConnectUserSpaceTunnelLockdown(device, device.UserspaceTUNPort)
			tun.UserspaceTUN = true
			tun.UserspaceTUNPort = device.UserspaceTUNPort
			return tun, err
		}
		return ConnectTunnelLockdown(device)
	}
	if version.Major() >= 17 {
		if userspaceTUN {
			return Tunnel{}, fmt.Errorf("manualPairingTunnelStart: userspaceTUN not supported for iOS >=17 and < 17.4: %w %s", ErrUnsupportedVersion, version.String())
		}
		return ManualPairAndConnectToTunnel(ctx, device, p)
	}
	return Tunnel{}, fmt.Errorf("manualPairingTunnelStart: %w %s", ErrUnsupportedVersion, version.String())
}

type deviceList struct {
}

func (d deviceList) ListDevices() (ios.DeviceList, error) {
	return ios.ListDevices()
}
