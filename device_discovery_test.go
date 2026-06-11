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
