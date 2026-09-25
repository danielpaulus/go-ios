package discovery

import (
	"reflect"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
)

func usbEntry(udid, connType string) ios.DeviceEntry {
	return ios.DeviceEntry{
		Properties: ios.DeviceProperties{
			SerialNumber:   udid,
			ConnectionType: connType,
		},
	}
}

func TestMergeDevices(t *testing.T) {
	tests := []struct {
		name    string
		usbmux  []ios.DeviceEntry
		bonjour []ios.RemotedDevice
		tunnels []TunnelInfo
		want    []Device
	}{
		{
			name: "empty inputs",
			want: []Device{},
		},
		{
			name: "usb and network usbmux devices",
			usbmux: []ios.DeviceEntry{
				usbEntry("bbb-usb", "USB"),
				usbEntry("aaa-net", "Network"),
			},
			want: []Device{
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
			name: "bonjour-only device",
			bonjour: []ios.RemotedDevice{
				{Udid: "bonjour-only", Address: "fd6f::9%en0", RsdPort: 50000},
			},
			want: []Device{
				{
					Udid: "bonjour-only",
					Transports: []Transport{{
						Type:    "tunnel",
						Source:  "bonjour",
						Address: "fd6f::9%en0",
						RsdPort: 50000,
					}},
				},
			},
		},
		{
			name: "tunnel-only device",
			tunnels: []TunnelInfo{
				{Udid: "tunnel-only", Address: "fd6f::1", RsdPort: 62173, UserspaceTunnelPort: 5000},
			},
			want: []Device{
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
			name:    "device in all three sources merges to one entry with three transports",
			usbmux:  []ios.DeviceEntry{usbEntry("triple", "USB")},
			bonjour: []ios.RemotedDevice{{Udid: "triple", Address: "fd6f::3%en0", RsdPort: 50}},
			tunnels: []TunnelInfo{
				{Udid: "triple", Address: "fd6f::2", RsdPort: 100, UserspaceTunnelPort: 200},
			},
			want: []Device{
				{
					Udid: "triple",
					Transports: []Transport{
						{Type: "usb", Source: "usbmux"},
						{Type: "tunnel", Source: "bonjour", Address: "fd6f::3%en0", RsdPort: 50},
						{Type: "tunnel", Source: "tunnel-agent", Address: "fd6f::2", RsdPort: 100, UserspaceTunnelPort: 200},
					},
				},
			},
		},
		{
			name:    "empty udids skipped",
			usbmux:  []ios.DeviceEntry{usbEntry("", "USB")},
			bonjour: []ios.RemotedDevice{{Udid: ""}},
			tunnels: []TunnelInfo{{Udid: ""}},
			want:    []Device{},
		},
		{
			name:   "empty connection type defaults to usb",
			usbmux: []ios.DeviceEntry{usbEntry("zzz", "")},
			want: []Device{
				{Udid: "zzz", Transports: []Transport{{Type: "usb", Source: "usbmux"}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeDevices(tt.usbmux, tt.bonjour, tt.tunnels)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("MergeDevices() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestMergeWifiPairing(t *testing.T) {
	devices := []Device{
		{Udid: "usb", Transports: []Transport{{Type: "usb", Source: "usbmux"}}},
		{
			Udid:        "coredevice-udid",
			Identifier:  "coredevice-id",
			ProductType: "iPhone15,2",
			ProductName: "iPhone",
			Transports:  []Transport{{Type: "wifi-pairing", Source: "coredevice"}},
		},
	}
	wifi := []ios.WifiPairingDevice{
		{Identifier: "pairable", Name: "iPhone", Model: "iPhone15,2", Address: "host.local.:1234"},
		{Identifier: "other-pairable", Name: "Other", Model: "iPhone15,3", Address: "other.local.:1234"},
		{Identifier: "pairable", Name: "iPhone", Model: "iPhone15,2", Address: "host.local.:1234"},
		{Identifier: ""},
	}

	got := MergeWifiPairing(devices, wifi)
	want := []Device{
		{
			Udid:        "coredevice-udid",
			Identifier:  "coredevice-id",
			ProductType: "iPhone15,2",
			ProductName: "iPhone",
			Transports:  []Transport{{Type: "wifi-pairing", Source: "coredevice"}},
		},
		{
			Identifier:  "other-pairable",
			ProductType: "iPhone15,3",
			ProductName: "Other",
			Transports:  []Transport{{Type: "wifi-pairing", Source: "bonjour", Address: "other.local.:1234"}},
		},
		{
			Udid:       "usb",
			Transports: []Transport{{Type: "usb", Source: "usbmux"}},
		},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MergeWifiPairing() = %+v, want %+v", got, want)
	}
}

func TestMergeCandidates(t *testing.T) {
	devices := []Device{{Udid: "existing"}}
	candidates := []Device{
		{Udid: "candidate", Identifier: "candidate-id"},
		{Udid: "existing", Identifier: "duplicate-udid"},
		{Identifier: "candidate-id"},
		{},
	}

	got := MergeCandidates(devices, candidates)
	want := []Device{
		{Udid: "candidate", Identifier: "candidate-id"},
		{Udid: "existing"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("MergeCandidates() = %+v, want %+v", got, want)
	}
}
