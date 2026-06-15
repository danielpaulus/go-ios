package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/discovery"
	"github.com/danielpaulus/go-ios/ios/tunnel"
)

// printMergedDeviceList lists all devices merged across discovery sources. It
// prefers the warm path: a running `ios tunnel start` agent serving GET /devices
// is queried first (short timeout). If the agent is unavailable, it falls back
// to discovering ad-hoc itself, folding in any running tunnels reported by the
// agent's /tunnels endpoint. Wi-Fi remote-pairing candidates are added so `list`
// reflects devices Xcode can see even before go-ios has a usable transport. With
// --details, usb/network devices are enriched with lockdown values. Output is a
// human table (--nojson) or JSON.
func printMergedDeviceList(details bool, cfg tunnelInfoConfig) {
	var merged []discovery.Device

	if devices, ok := devicesFromAgent(cfg); ok {
		merged = devices
	}

	if merged == nil {
		var tunnels []discovery.TunnelInfo
		running, err := tunnel.ListRunningTunnels(cfg.Host, cfg.Port)
		if err != nil {
			slog.Debug("failed to list running tunnels, continuing without tunnel sources", "err", err)
		} else {
			tunnels = tunnelInfosFromTunnels(running)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		merged = discovery.Discover(ctx, tunnels, true)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	merged = discovery.MergeCandidates(merged, discoverCoreDeviceWifiPairing(ctx))
	merged = discovery.MergeWifiPairing(merged, discovery.DiscoverWifiPairing(ctx))

	if details {
		enrichLocalDevices(merged)
	}

	if JSONdisabled {
		fmt.Print(renderDeviceTable(merged))
	} else {
		fmt.Println(convertToJSONString(map[string][]discovery.Device{"devices": merged}))
	}
}

// devicesFromAgent fetches the warm device list from a running tunnel agent's
// GET /devices endpoint with a short timeout. It returns (devices, true) on a
// successful 200 + JSON response, and (nil, false) if the agent is unavailable
// or returns an error, signalling the caller to fall back to ad-hoc discovery.
func devicesFromAgent(cfg tunnelInfoConfig) ([]discovery.Device, bool) {
	c := http.Client{Timeout: 1 * time.Second}
	resp, err := c.Get(fmt.Sprintf("http://%s:%d/devices", cfg.Host, cfg.Port))
	if err != nil {
		slog.Debug("tunnel agent /devices unavailable, falling back to ad-hoc discovery", "err", err)
		return nil, false
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		slog.Debug("tunnel agent /devices returned non-200, falling back to ad-hoc discovery", "status", resp.StatusCode)
		return nil, false
	}
	var devices []discovery.Device
	if err := json.NewDecoder(resp.Body).Decode(&devices); err != nil {
		slog.Debug("failed to decode tunnel agent /devices response, falling back to ad-hoc discovery", "err", err)
		return nil, false
	}
	return devices, true
}

// tunnelInfosFromTunnels maps tunnel-agent tunnels to the plain TunnelInfo input
// the discovery package consumes.
func tunnelInfosFromTunnels(tunnels []tunnel.Tunnel) []discovery.TunnelInfo {
	infos := make([]discovery.TunnelInfo, 0, len(tunnels))
	for _, t := range tunnels {
		infos = append(infos, discovery.TunnelInfo{
			Udid:                t.Udid,
			Address:             t.Address,
			RsdPort:             t.RsdPort,
			UserspaceTunnelPort: t.UserspaceTUNPort,
		})
	}
	return infos
}

// enrichLocalDevices fills in product info for devices reachable over usb or
// network by querying lockdown values (best-effort).
func enrichLocalDevices(devices []discovery.Device) {
	deviceList, err := ios.ListDevices()
	if err != nil {
		slog.Debug("failed getting device list for enrichment", "err", err)
		return
	}
	entryByUdid := map[string]ios.DeviceEntry{}
	for _, entry := range deviceList.DeviceList {
		entryByUdid[entry.Properties.SerialNumber] = entry
	}
	for i := range devices {
		if !hasLocalTransport(devices[i]) {
			continue
		}
		entry, ok := entryByUdid[devices[i].Udid]
		if !ok {
			continue
		}
		allValues, err := ios.GetValues(entry)
		if err != nil {
			slog.Debug("failed getting values for device", "udid", devices[i].Udid, "err", err)
			continue
		}
		devices[i].ProductType = allValues.Value.ProductType
		devices[i].ProductVersion = allValues.Value.ProductVersion
		devices[i].ProductName = allValues.Value.ProductName
	}
}

// hasLocalTransport reports whether the device is reachable over usb or network
// (i.e. a usbmux source we can query lockdown values from).
func hasLocalTransport(d discovery.Device) bool {
	for _, t := range d.Transports {
		if t.Type == "usb" || t.Type == "network" {
			return true
		}
	}
	return false
}

// renderDeviceTable renders devices as a human-readable, column-aligned table.
// It is pure (no I/O) and returns the table as a string.
func renderDeviceTable(devices []discovery.Device) string {
	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "UDID/ID\tMODEL\tOS\tVIA\tADDRESS")
	for _, d := range devices {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			deviceDisplayID(d),
			dashIfEmpty(d.ProductType),
			dashIfEmpty(d.ProductVersion),
			transportVia(d.Transports),
			transportAddress(d.Transports),
		)
	}
	w.Flush()
	return sb.String()
}

func deviceDisplayID(d discovery.Device) string {
	if d.Udid != "" {
		return d.Udid
	}
	return d.Identifier
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// transportVia joins transport types with commas (e.g. "usb,tunnel").
func transportVia(transports []discovery.Transport) string {
	types := make([]string, len(transports))
	for i, t := range transports {
		types[i] = t.Type
	}
	return strings.Join(types, ",")
}

// transportAddress returns "<address>:<rsdPort>" for the first transport that
// carries an address, or "-" if none do.
func transportAddress(transports []discovery.Transport) string {
	for _, t := range transports {
		if t.Address != "" {
			if t.RsdPort == 0 {
				return t.Address
			}
			return fmt.Sprintf("%s:%d", t.Address, t.RsdPort)
		}
	}
	return "-"
}
