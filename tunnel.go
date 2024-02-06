package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tunnel"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/exp/slices"
)

const tunnelInfoDefaultHttpPort = 28100

// serveTunnelInfo starts a simple http serve that exposes the tunnel information about the running tunnel
func serveTunnelInfo(tm tunnelManager, port int) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/tunnel/", func(writer http.ResponseWriter, request *http.Request) {
		udid := strings.TrimPrefix(request.URL.Path, "/tunnel/")
		if len(udid) == 0 {
			return
		}
		tunnels, err := tm.ListTunnels()
		idx := slices.IndexFunc(tunnels, func(info tunnelInfo) bool {
			return info.Udid == udid
		})
		if idx < 0 {
			http.Error(writer, "", http.StatusNotFound)
			return
		}
		t := tunnels[idx]

		writer.Header().Add("Content-Type", "application/json")
		enc := json.NewEncoder(writer)
		err = enc.Encode(t)
		if err != nil {
			return
		}
	})
	mux.HandleFunc("/tunnels", func(writer http.ResponseWriter, request *http.Request) {
		tunnels, err := tm.ListTunnels()

		writer.Header().Add("Content-Type", "application/json")
		enc := json.NewEncoder(writer)
		err = enc.Encode(tunnels)
		if err != nil {
			return
		}
	})
	if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), mux); err != nil {
		return fmt.Errorf("serveTunnelInfo: failed to start http server: %w", err)
	}
	return nil
}

type TunnelClient struct {
	h    http.Client
	port int
}

func tunnelInfoForDevice(udid string, tunnelInfoPort int) (tunnelInfo, error) {
	c := http.Client{
		Timeout: 5 * time.Second,
	}
	res, err := c.Get(fmt.Sprintf("http://127.0.0.1:%d/tunnel/%s", tunnelInfoPort, udid))
	if err != nil {
		return tunnelInfo{}, fmt.Errorf("Tunnel: failed to get tunnel info: %w", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return tunnelInfo{}, fmt.Errorf("Tunnel: failed to read body: %w", err)
	}
	var info tunnelInfo
	err = json.Unmarshal(body, &info)
	if err != nil {
		return tunnelInfo{}, fmt.Errorf("Tunnel: faile to parse response: %w", err)
	}
	return info, nil
}

type tunnelInfo struct {
	Address string `json:"address"`
	RsdPort int    `json:"rsdPort"`
	Udid    string `json:"udid"`
	closer  io.Closer
}

type tunnelManager struct {
	ts tunnelStarter
	dl deviceLister
	pm tunnel.PairRecordManager

	tunnels map[string]tunnelInfo
}

func (m *tunnelManager) UpdateTunnels() error {
	devices, err := m.dl.ListDevices()
	if err != nil {
		return fmt.Errorf("UpdateTunnels: failed to get list of devices: %w", err)
	}
	for _, d := range devices.DeviceList {
		udid := d.Properties.SerialNumber
		if _, exists := m.tunnels[udid]; exists {
			continue
		}
		log.WithField("udid", udid).Info("start tunnel")
		t, err := m.ts.StartTunnel(context.TODO(), d, m.pm)
		if err != nil {
			continue
		}
		m.tunnels[udid] = tunnelInfo{
			Address: t.Address,
			RsdPort: t.RsdPort,
			Udid:    udid,
			closer:  t.Closer,
		}
	}
	for udid, tun := range m.tunnels {
		idx := slices.ContainsFunc(devices.DeviceList, func(entry ios.DeviceEntry) bool {
			return entry.Properties.SerialNumber == udid
		})
		if !idx {
			log.WithField("udid", udid).Info("stopping tunnel")
			_ = tun.closer.Close()
			delete(m.tunnels, udid)
		}
	}
	return nil
}

func (m *tunnelManager) ListTunnels() ([]tunnelInfo, error) {
	return maps.Values(m.tunnels), nil
}

type tunnelStarter interface {
	StartTunnel(ctx context.Context, device ios.DeviceEntry, p tunnel.PairRecordManager) (tunnel.Tunnel, error)
}

type deviceLister interface {
	ListDevices() (ios.DeviceList, error)
}

type manualPairingTunnelStart struct {
}

func (m manualPairingTunnelStart) StartTunnel(ctx context.Context, device ios.DeviceEntry, p tunnel.PairRecordManager) (tunnel.Tunnel, error) {
	return tunnel.ManualPairAndConnectToTunnel(ctx, device, p)
}

type deviceList struct {
}

func (d deviceList) ListDevices() (ios.DeviceList, error) {
	return ios.ListDevices()
}
