package tunnel

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

// DefaultHttpApiPort is the port on which we start the HTTP-Server for exposing started tunnels
const DefaultHttpApiPort = 28100

// ServeTunnelInfo starts a simple http serve that exposes the tunnel information about the running tunnel.
// The API has two endpoints:
// 1. GET    localhost:{PORT}/tunnel/{UDID} to get the tunnel info for a specific device
// 2. DELETE localhost:{PORT}/tunnel/{UDID} to stop a device tunnel
// 3. GET    localhost:{PORT}/tunnels       to get a list of all tunnels
func ServeTunnelInfo(tm *TunnelManager, port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/tunnel/", func(writer http.ResponseWriter, request *http.Request) {
		udid := strings.TrimPrefix(request.URL.Path, "/tunnel/")
		if len(udid) == 0 {
			return
		}

		t, err := tm.FindTunnel(udid)
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
		if len(t.Udid) == 0 {
			http.Error(writer, "", http.StatusNotFound)
			return
		}

		if request.Method == "GET" {
			writer.Header().Add("Content-Type", "application/json")
			enc := json.NewEncoder(writer)
			err = enc.Encode(t)
		} else if request.Method == "DELETE" {
			err = tm.stopTunnel(t)
		}
		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
			return
		}
	})
	mux.HandleFunc("/tunnels", func(writer http.ResponseWriter, request *http.Request) {
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
	if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), mux); err != nil {
		return fmt.Errorf("ServeTunnelInfo: failed to start http server: %w", err)
	}
	return nil
}

func TunnelInfoForDevice(udid string, tunnelInfoPort int) (Tunnel, error) {
	c := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := c.Get(fmt.Sprintf("http://127.0.0.1:%d/tunnel/%s", tunnelInfoPort, udid))
	if err != nil {
		return Tunnel{}, fmt.Errorf("TunnelInfoForDevice: failed to get tunnel info: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return Tunnel{}, fmt.Errorf("TunnelInfoForDevice: failed to read body: %w", err)
	}
	var info Tunnel
	err = json.Unmarshal(body, &info)
	if err != nil {
		return Tunnel{}, fmt.Errorf("TunnelInfoForDevice: failed to parse response: %w", err)
	}
	return info, nil
}

func ListRunningTunnels(tunnelInfoPort int) ([]Tunnel, error) {
	c := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := c.Get(fmt.Sprintf("http://127.0.0.1:%d/tunnels", tunnelInfoPort))
	if err != nil {
		return nil, fmt.Errorf("TunnelInfoForDevice: failed to get tunnel info: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("TunnelInfoForDevice: failed to read body: %w", err)
	}
	var info []Tunnel
	err = json.Unmarshal(body, &info)
	if err != nil {
		return nil, fmt.Errorf("TunnelInfoForDevice: failed to parse response: %w", err)
	}
	return info, nil
}

// TunnelManager starts tunnels for devices when needed (if no tunnel is running yet) and stores the information
// how those tunnels are reachable (address and remote service discovery port)
type TunnelManager struct {
	ts                 tunnelStarter
	dl                 deviceLister
	pm                 PairRecordManager
	mux                sync.Mutex
	tunnels            map[string]Tunnel
	startTunnelTimeout time.Duration
}

// NewTunnelManager creates a new TunnelManager instance for setting up device tunnels for all connected devices
func NewTunnelManager(pm PairRecordManager) *TunnelManager {
	return &TunnelManager{
		ts:                 manualPairingTunnelStart{},
		dl:                 deviceList{},
		pm:                 pm,
		tunnels:            map[string]Tunnel{},
		startTunnelTimeout: 10 * time.Second,
	}
}

// UpdateTunnels checks for connected devices and starts a new tunnel if needed
// On device disconnects the tunnel resources get cleaned up
func (m *TunnelManager) UpdateTunnels(ctx context.Context) error {

	m.mux.Lock()
	localTunnels := map[string]Tunnel{}
	maps.Copy(localTunnels, m.tunnels)
	m.mux.Unlock()

	devices, err := m.dl.ListDevices()
	if err != nil {
		return fmt.Errorf("UpdateTunnels: failed to get list of devices: %w", err)
	}
	for _, d := range devices.DeviceList {
		udid := d.Properties.SerialNumber
		if _, exists := localTunnels[udid]; exists {
			continue
		}

		t, err := m.startTunnel(ctx, d)
		if err != nil {
			log.WithField("udid", udid).
				WithError(err).
				Warn("failed to start tunnel")
			continue
		}
		m.mux.Lock()
		localTunnels[udid] = t
		m.tunnels[udid] = t
		m.mux.Unlock()
	}
	for udid, tun := range localTunnels {
		idx := slices.ContainsFunc(devices.DeviceList, func(entry ios.DeviceEntry) bool {
			return entry.Properties.SerialNumber == udid
		})
		if !idx {
			_ = m.stopTunnel(tun)
		}
	}
	return nil
}

func (m *TunnelManager) stopTunnel(t Tunnel) error {
	m.mux.Lock()
	defer m.mux.Unlock()
	log.WithField("udid", t.Udid).Info("stopping tunnel")
	delete(m.tunnels, t.Udid)

	return t.Close()
}

func (m *TunnelManager) startTunnel(ctx context.Context, device ios.DeviceEntry) (Tunnel, error) {
	log.WithField("udid", device.Properties.SerialNumber).Info("start tunnel")
	t1, err := ConnectUserSpaceTunnelLockdown(device)
	return t1, err
	startTunnelCtx, cancel := context.WithTimeout(ctx, m.startTunnelTimeout)
	defer cancel()
	version, err := ios.GetProductVersion(device)
	if err != nil {
		return Tunnel{}, fmt.Errorf("startTunnel: failed to get device version: %w", err)
	}
	t, err := m.ts.StartTunnel(startTunnelCtx, device, m.pm, version)
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
	StartTunnel(ctx context.Context, device ios.DeviceEntry, p PairRecordManager, version *semver.Version) (Tunnel, error)
}

type deviceLister interface {
	ListDevices() (ios.DeviceList, error)
}

type manualPairingTunnelStart struct {
}

func (m manualPairingTunnelStart) StartTunnel(ctx context.Context, device ios.DeviceEntry, p PairRecordManager, version *semver.Version) (Tunnel, error) {

	if version.Major() >= 17 && version.Minor() >= 4 {
		return ConnectTunnelLockdown(device)
	}
	if version.Major() >= 17 {
		return ManualPairAndConnectToTunnel(ctx, device, p)
	}
	return Tunnel{}, fmt.Errorf("manualPairingTunnelStart: unsupported iOS version %s", version.String())
}

type deviceList struct {
}

func (d deviceList) ListDevices() (ios.DeviceList, error) {
	return ios.ListDevices()
}
