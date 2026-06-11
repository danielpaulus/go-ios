package main

import (
	"reflect"
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tunnel"
)

func usbEntry(udid, connType string) ios.DeviceEntry {
	return ios.DeviceEntry{
		Properties: ios.DeviceProperties{
			SerialNumber:   udid,
			ConnectionType: connType,
		},
	}
}

func TestMergeDiscoveredDevices(t *testing.T) {
	tests := []struct {
		name    string
		usbmux  []ios.DeviceEntry
		tunnels []tunnel.Tunnel
		want    []DiscoveredDevice
	}{
		{
			name:    "empty inputs",
			usbmux:  nil,
			tunnels: nil,
			want:    []DiscoveredDevice{},
		},
		{
			name: "usb and network usbmux devices",
			usbmux: []ios.DeviceEntry{
				usbEntry("bbb-usb", "USB"),
				usbEntry("aaa-net", "Network"),
			},
			want: []DiscoveredDevice{
				{
					Udid:       "aaa-net",
					Transports: []Transport{{Type: "network", Source: "usbmux"}},
				},
				{
					Udid:       "bbb-usb",
					Transports: []Transport{{Type: "usb", Source: "usbmux"}},
				},
			},
		},
		{
			name: "tunnel-only device",
			tunnels: []tunnel.Tunnel{
				{
					Udid:             "tunnel-only",
					Address:          "fd6f::1",
					RsdPort:          62173,
					UserspaceTUNPort: 5000,
				},
			},
			want: []DiscoveredDevice{
				{
					Udid: "tunnel-only",
					Transports: []Transport{{
						Type:                "tunnel",
						Source:              "tunnel-agent",
						Address:             "fd6f::1",
						RsdPort:             62173,
						UserspaceTunnelPort: 5000,
					}},
				},
			},
		},
		{
			name:   "device in both sources merges to one entry",
			usbmux: []ios.DeviceEntry{usbEntry("dual", "USB")},
			tunnels: []tunnel.Tunnel{
				{Udid: "dual", Address: "fd6f::2", RsdPort: 100, UserspaceTUNPort: 200},
			},
			want: []DiscoveredDevice{
				{
					Udid: "dual",
					Transports: []Transport{
						{Type: "usb", Source: "usbmux"},
						{Type: "tunnel", Source: "tunnel-agent", Address: "fd6f::2", RsdPort: 100, UserspaceTunnelPort: 200},
					},
				},
			},
		},
		{
			name:    "empty udids skipped",
			usbmux:  []ios.DeviceEntry{usbEntry("", "USB")},
			tunnels: []tunnel.Tunnel{{Udid: ""}},
			want:    []DiscoveredDevice{},
		},
		{
			name:   "empty connection type defaults to usb",
			usbmux: []ios.DeviceEntry{usbEntry("zzz", "")},
			want: []DiscoveredDevice{
				{Udid: "zzz", Transports: []Transport{{Type: "usb", Source: "usbmux"}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeDiscoveredDevices(tt.usbmux, tt.tunnels)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("mergeDiscoveredDevices() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestRenderDeviceTable(t *testing.T) {
	devices := []DiscoveredDevice{
		{
			Udid:           "usb-only",
			ProductType:    "iPhone14,2",
			ProductVersion: "17.0",
			Transports:     []Transport{{Type: "usb", Source: "usbmux"}},
		},
		{
			Udid: "usb-tunnel",
			Transports: []Transport{
				{Type: "usb", Source: "usbmux"},
				{Type: "tunnel", Source: "tunnel-agent", Address: "fd6f::1", RsdPort: 62173},
			},
		},
	}

	out := renderDeviceTable(devices)

	for _, want := range []string{
		"UDID", "MODEL", "OS", "VIA", "ADDRESS",
		"usb-only", "iPhone14,2", "17.0",
		"usb,tunnel",
		"fd6f::1:62173",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("renderDeviceTable() output missing %q\n--- output ---\n%s", want, out)
		}
	}

	// usb-only device has no address transport -> "-"
	if !strings.Contains(out, "-") {
		t.Errorf("expected a dash placeholder for missing fields in:\n%s", out)
	}
}
