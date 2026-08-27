package main

import (
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/ios/discovery"
)

func TestRenderDeviceTable(t *testing.T) {
	devices := []discovery.Device{
		{
			Udid:           "usb-only",
			ProductType:    "iPhone14,2",
			ProductVersion: "17.0",
			Transports:     []discovery.Transport{{Type: "usb", Source: "usbmux"}},
		},
		{
			Udid: "usb-tunnel",
			Transports: []discovery.Transport{
				{Type: "usb", Source: "usbmux"},
				{Type: "tunnel", Source: "tunnel-agent", Address: "fd6f::1", RsdPort: 62173},
			},
		},
	}

	out := renderDeviceTable(devices)

	for _, want := range []string{
		"UDID/ID", "MODEL", "OS", "VIA", "ADDRESS",
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

func TestRenderDeviceTableWifiPairingCandidate(t *testing.T) {
	devices := []discovery.Device{
		{
			Identifier:  "remote-id",
			ProductType: "iPhone15,2",
			ProductName: "iPhone",
			Transports:  []discovery.Transport{{Type: "wifi-pairing", Source: "bonjour", Address: "host.local.:1234"}},
		},
	}

	out := renderDeviceTable(devices)

	if strings.Contains(out, "host.local.:1234:0") {
		t.Fatalf("renderDeviceTable() appended an RSD port to a non-RSD address\n--- output ---\n%s", out)
	}
	for _, want := range []string{"remote-id", "iPhone15,2", "wifi-pairing", "host.local.:1234"} {
		if !strings.Contains(out, want) {
			t.Errorf("renderDeviceTable() output missing %q\n--- output ---\n%s", want, out)
		}
	}
}
