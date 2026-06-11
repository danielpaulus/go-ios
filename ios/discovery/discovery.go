// Package discovery is the shared device/transport model for go-ios device
// discovery. It merges devices from several sources (usbmux, Bonjour/remoted,
// and running tunnels) into one Device per udid, each carrying the transports
// over which it is reachable.
//
// It imports only the ios package (plus context and the stdlib). In particular
// it must NOT import ios/tunnel: ios/tunnel (the agent) imports this package to
// serve a warm registry, so the reverse would be an import cycle. Tunnel data is
// folded in by callers via the plain TunnelInfo input, not fetched here.
package discovery

import (
	"context"
	"sort"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
)

const logModule = "go-ios/discovery"

// Transport describes one way a discovered device is reachable.
type Transport struct {
	Type                string `json:"type"`   // "usb" | "network" | "tunnel"
	Source              string `json:"source"` // "usbmux" | "tunnel-agent" | "bonjour"
	Address             string `json:"address,omitempty"`
	RsdPort             int    `json:"rsdPort,omitempty"`
	UserspaceTunnelPort int    `json:"userspaceTunnelPort,omitempty"`
}

// Device is a single device merged from all discovery sources, keyed by udid.
type Device struct {
	Udid           string      `json:"udid"`
	ProductType    string      `json:"productType,omitempty"`
	ProductVersion string      `json:"productVersion,omitempty"`
	ProductName    string      `json:"productName,omitempty"`
	Transports     []Transport `json:"transports"`
}

// TunnelInfo is plain input describing a running tunnel, so discovery needn't
// import ios/tunnel. Callers map their own tunnel representation into this.
type TunnelInfo struct {
	Udid                string
	Address             string
	RsdPort             int
	UserspaceTunnelPort int
}

// MergeDevices merges usbmux entries, Bonjour/remoted devices and running
// tunnels into one Device per udid. It is pure (no I/O). Transports are appended
// in source order (usbmux, then bonjour, then tunnels). Empty udids are skipped.
// Output is sorted by udid for deterministic results.
func MergeDevices(usbmux []ios.DeviceEntry, bonjour []ios.RemotedDevice, tunnels []TunnelInfo) []Device {
	byUdid := map[string]*Device{}
	var order []string

	get := func(udid string) *Device {
		if d, ok := byUdid[udid]; ok {
			return d
		}
		d := &Device{Udid: udid}
		byUdid[udid] = d
		order = append(order, udid)
		return d
	}

	for _, entry := range usbmux {
		udid := entry.Properties.SerialNumber
		if udid == "" {
			continue
		}
		d := get(udid)
		d.Transports = append(d.Transports, Transport{
			Type:   usbmuxTransportType(entry),
			Source: "usbmux",
		})
	}

	for _, b := range bonjour {
		if b.Udid == "" {
			continue
		}
		d := get(b.Udid)
		d.Transports = append(d.Transports, Transport{
			Type:    "tunnel",
			Source:  "bonjour",
			Address: b.Address,
			RsdPort: b.RsdPort,
		})
	}

	for _, t := range tunnels {
		if t.Udid == "" {
			continue
		}
		d := get(t.Udid)
		d.Transports = append(d.Transports, Transport{
			Type:                "tunnel",
			Source:              "tunnel-agent",
			Address:             t.Address,
			RsdPort:             t.RsdPort,
			UserspaceTunnelPort: t.UserspaceTunnelPort,
		})
	}

	result := make([]Device, 0, len(order))
	for _, udid := range order {
		result = append(result, *byUdid[udid])
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Udid < result[j].Udid
	})
	return result
}

// Discover gathers devices from usbmux and (optionally) Bonjour/remoted, both
// best-effort, and merges them with the supplied running tunnels. Each source
// failure is logged and skipped; Discover never returns an error that aborts the
// listing — it returns whatever it managed to merge. The Bonjour browse is
// bounded by ctx.
//
// browseBonjour MUST be false for any caller that already holds an RSD/tunnel
// session to the devices (notably the running tunnel agent): a device allows
// only one RemoteXPC/RSD session at a time, so a second handshake from the
// Bonjour browse is RST'd by the device (see issue #724) and, run repeatedly,
// can destabilise the agent's own tunnels. The ad-hoc CLI path sets it true
// because it only runs when no agent holds a session.
func Discover(ctx context.Context, tunnels []TunnelInfo, browseBonjour bool) []Device {
	var usbmux []ios.DeviceEntry
	deviceList, err := ios.ListDevices()
	if err != nil {
		golog.Debug("failed to list usbmux devices, continuing", "module", logModule, "err", err)
	} else {
		usbmux = deviceList.DeviceList
	}

	var bonjour []ios.RemotedDevice
	if browseBonjour {
		bonjour, err = ios.BrowseRemoted(ctx)
		if err != nil {
			golog.Debug("failed to browse remoted devices, continuing", "module", logModule, "err", err)
			bonjour = nil
		}
	}

	return MergeDevices(usbmux, bonjour, tunnels)
}

// usbmuxTransportType maps a usbmux device's connection type to a transport type.
// usbmuxd's default (empty connection type) is a USB connection.
func usbmuxTransportType(entry ios.DeviceEntry) string {
	switch entry.Properties.ConnectionType {
	case "USB":
		return "usb"
	case "Network":
		return "network"
	case "":
		return "usb"
	default:
		return strings.ToLower(entry.ConnectionTypeLabel())
	}
}
