package main

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tunnel"
)

// Transport describes one way a discovered device is reachable.
type Transport struct {
	Type                string `json:"type"`   // "usb" | "network" | "tunnel"
	Source              string `json:"source"` // "usbmux" | "tunnel-agent"
	Address             string `json:"address,omitempty"`
	RsdPort             int    `json:"rsdPort,omitempty"`
	UserspaceTunnelPort int    `json:"userspaceTunnelPort,omitempty"`
}

// DiscoveredDevice is a single device merged from all discovery sources, keyed by udid.
type DiscoveredDevice struct {
	Udid           string      `json:"udid"`
	ProductType    string      `json:"productType,omitempty"`
	ProductVersion string      `json:"productVersion,omitempty"`
	ProductName    string      `json:"productName,omitempty"`
	Transports     []Transport `json:"transports"`
}

// mergeDiscoveredDevices merges usbmux entries and tunnel-agent tunnels into one
// DiscoveredDevice per udid, with usbmux transport(s) listed before tunnel ones.
// It is pure (no I/O) and returns devices sorted by udid for deterministic output.
func mergeDiscoveredDevices(usbmux []ios.DeviceEntry, tunnels []tunnel.Tunnel) []DiscoveredDevice {
	byUdid := map[string]*DiscoveredDevice{}
	var order []string

	get := func(udid string) *DiscoveredDevice {
		if d, ok := byUdid[udid]; ok {
			return d
		}
		d := &DiscoveredDevice{Udid: udid}
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
			UserspaceTunnelPort: t.UserspaceTUNPort,
		})
	}

	result := make([]DiscoveredDevice, 0, len(order))
	for _, udid := range order {
		result = append(result, *byUdid[udid])
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Udid < result[j].Udid
	})
	return result
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

// printMergedDeviceList fetches devices from usbmux and the tunnel agent, merges
// them by udid, optionally enriches usb/network devices with lockdown values, and
// prints either a human table (--nojson) or JSON.
func printMergedDeviceList(details bool, cfg tunnelInfoConfig) {
	deviceList, err := ios.ListDevices()
	exitIfError("failed getting device list", err)

	tunnels, err := tunnel.ListRunningTunnels(cfg.Host, cfg.Port)
	if err != nil {
		slog.Debug("failed to list running tunnels, continuing without tunnel sources", "err", err)
		tunnels = nil
	}

	merged := mergeDiscoveredDevices(deviceList.DeviceList, tunnels)

	if details {
		entryByUdid := map[string]ios.DeviceEntry{}
		for _, entry := range deviceList.DeviceList {
			entryByUdid[entry.Properties.SerialNumber] = entry
		}
		for i := range merged {
			if !hasLocalTransport(merged[i]) {
				continue
			}
			entry, ok := entryByUdid[merged[i].Udid]
			if !ok {
				continue
			}
			allValues, err := ios.GetValues(entry)
			if err != nil {
				slog.Debug("failed getting values for device", "udid", merged[i].Udid, "err", err)
				continue
			}
			merged[i].ProductType = allValues.Value.ProductType
			merged[i].ProductVersion = allValues.Value.ProductVersion
			merged[i].ProductName = allValues.Value.ProductName
		}
	}

	if JSONdisabled {
		fmt.Print(renderDeviceTable(merged))
	} else {
		fmt.Println(convertToJSONString(map[string][]DiscoveredDevice{"devices": merged}))
	}
}

// hasLocalTransport reports whether the device is reachable over usb or network
// (i.e. a usbmux source we can query lockdown values from).
func hasLocalTransport(d DiscoveredDevice) bool {
	for _, t := range d.Transports {
		if t.Type == "usb" || t.Type == "network" {
			return true
		}
	}
	return false
}

// renderDeviceTable renders devices as a human-readable, column-aligned table.
// It is pure (no I/O) and returns the table as a string.
func renderDeviceTable(devices []DiscoveredDevice) string {
	var sb strings.Builder
	w := tabwriter.NewWriter(&sb, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "UDID\tMODEL\tOS\tVIA\tADDRESS")
	for _, d := range devices {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			d.Udid,
			dashIfEmpty(d.ProductType),
			dashIfEmpty(d.ProductVersion),
			transportVia(d.Transports),
			transportAddress(d.Transports),
		)
	}
	w.Flush()
	return sb.String()
}

func dashIfEmpty(s string) string {
	if s == "" {
		return "-"
	}
	return s
}

// transportVia joins transport types with commas (e.g. "usb,tunnel").
func transportVia(transports []Transport) string {
	types := make([]string, len(transports))
	for i, t := range transports {
		types[i] = t.Type
	}
	return strings.Join(types, ",")
}

// transportAddress returns "<address>:<rsdPort>" for the first transport that
// carries an address, or "-" if none do.
func transportAddress(transports []Transport) string {
	for _, t := range transports {
		if t.Address != "" {
			return fmt.Sprintf("%s:%d", t.Address, t.RsdPort)
		}
	}
	return "-"
}
