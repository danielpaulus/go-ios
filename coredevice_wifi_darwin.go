//go:build darwin

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"os/exec"
	"strings"

	"github.com/danielpaulus/go-ios/ios/discovery"
)

type coreDeviceListOutput struct {
	Result struct {
		Devices []coreDeviceListDevice `json:"devices"`
	} `json:"result"`
}

type coreDeviceListDevice struct {
	Identifier           string `json:"identifier"`
	ConnectionProperties struct {
		AuthenticationType string   `json:"authenticationType"`
		PotentialHostnames []string `json:"potentialHostnames"`
		TransportType      string   `json:"transportType"`
		TunnelState        string   `json:"tunnelState"`
	} `json:"connectionProperties"`
	DeviceProperties struct {
		Name            string `json:"name"`
		OSVersionNumber string `json:"osVersionNumber"`
	} `json:"deviceProperties"`
	HardwareProperties struct {
		ProductType string `json:"productType"`
		Udid        string `json:"udid"`
	} `json:"hardwareProperties"`
}

func discoverCoreDeviceWifiPairing(ctx context.Context) []discovery.Device {
	out, err := exec.CommandContext(ctx, "xcrun", "devicectl", "list", "devices", "--json-output", "-").CombinedOutput()
	if err != nil {
		slog.Debug("failed to list CoreDevice devices", "err", err)
		return nil
	}
	jsonStart := bytes.IndexByte(out, '{')
	if jsonStart < 0 {
		slog.Debug("CoreDevice device list did not contain JSON output")
		return nil
	}
	var decoded coreDeviceListOutput
	if err := json.Unmarshal(out[jsonStart:], &decoded); err != nil {
		slog.Debug("failed to decode CoreDevice device list", "err", err)
		return nil
	}

	devices := make([]discovery.Device, 0, len(decoded.Result.Devices))
	for _, d := range decoded.Result.Devices {
		if d.HardwareProperties.Udid == "" || !isCoreDeviceWifiCandidate(d) {
			continue
		}
		devices = append(devices, discovery.Device{
			Udid:           d.HardwareProperties.Udid,
			Identifier:     d.Identifier,
			ProductType:    d.HardwareProperties.ProductType,
			ProductVersion: d.DeviceProperties.OSVersionNumber,
			ProductName:    d.DeviceProperties.Name,
			Transports: []discovery.Transport{{
				Type:    "wifi-pairing",
				Source:  "coredevice",
				Address: firstCoreDeviceHostname(d.ConnectionProperties.PotentialHostnames),
			}},
		})
	}
	return devices
}

func isCoreDeviceWifiCandidate(d coreDeviceListDevice) bool {
	if d.ConnectionProperties.TransportType == "localNetwork" {
		return true
	}
	if d.ConnectionProperties.AuthenticationType == "manualPairing" && d.ConnectionProperties.TunnelState != "" {
		return true
	}
	for _, hostname := range d.ConnectionProperties.PotentialHostnames {
		if strings.HasSuffix(hostname, ".coredevice.local") {
			return true
		}
	}
	return false
}

func firstCoreDeviceHostname(hostnames []string) string {
	for _, hostname := range hostnames {
		if hostname != "" {
			return hostname
		}
	}
	return ""
}
