package main

import (
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tunnel"
)

// fakeTunnelAgent serves the tunnel agent's /tunnels endpoint on 127.0.0.1 and
// returns the matching tunnelInfoConfig.
func fakeTunnelAgent(t *testing.T, tunnels []tunnel.Tunnel) tunnelInfoConfig {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/tunnels" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(tunnels); err != nil {
			t.Errorf("failed to encode tunnels: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	host, portString, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		t.Fatalf("failed to parse test server address: %v", err)
	}
	port, err := strconv.Atoi(portString)
	if err != nil {
		t.Fatalf("failed to parse test server port: %v", err)
	}
	return tunnelInfoConfig{Host: host, Port: port}
}

func TestTunnelBackedDevicesMergeIntoEmptyUsbmuxList(t *testing.T) {
	tunnelInfo := fakeTunnelAgent(t, []tunnel.Tunnel{
		{Udid: "00008120-001905DE2E9B401E", Address: "fd6f:7edb:be57::1", RsdPort: 62173, UserspaceTUN: true, UserspaceTUNPort: 61246},
		{Udid: "00008030-000E4C523C40802E", Address: "fd7b:15e6:a8ae::1", RsdPort: 55555},
	})

	devices, err := tunnelBackedDevices(tunnelInfo)
	if err != nil {
		t.Fatalf("tunnelBackedDevices failed: %v", err)
	}
	merged := mergeTunnelDevices(ios.DeviceList{}, devices)
	if len(merged.DeviceList) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(merged.DeviceList))
	}

	userspaceDevice := merged.DeviceList[0]
	if userspaceDevice.Properties.SerialNumber != "00008120-001905DE2E9B401E" {
		t.Errorf("unexpected udid: %s", userspaceDevice.Properties.SerialNumber)
	}
	if userspaceDevice.Properties.ConnectionType != connectionTypeUserspaceTunnel {
		t.Errorf("expected connectionType %q, got %q", connectionTypeUserspaceTunnel, userspaceDevice.Properties.ConnectionType)
	}
	if userspaceDevice.Address != "fd6f:7edb:be57::1" {
		t.Errorf("unexpected address: %s", userspaceDevice.Address)
	}
	if !userspaceDevice.UserspaceTUN || userspaceDevice.UserspaceTUNPort != 61246 {
		t.Errorf("userspace TUN fields not propagated: %+v", userspaceDevice)
	}
	if userspaceDevice.UserspaceTUNHost != tunnelInfo.Host {
		t.Errorf("expected userspace TUN host %q, got %q", tunnelInfo.Host, userspaceDevice.UserspaceTUNHost)
	}

	kernelDevice := merged.DeviceList[1]
	if kernelDevice.Properties.ConnectionType != connectionTypeTunnel {
		t.Errorf("expected connectionType %q, got %q", connectionTypeTunnel, kernelDevice.Properties.ConnectionType)
	}
	if !isTunnelOnlyDevice(userspaceDevice) || !isTunnelOnlyDevice(kernelDevice) {
		t.Error("tunnel-backed entries must be detected as tunnel-only")
	}
}

func TestMergeTunnelDevicesPrefersUsbmuxEntries(t *testing.T) {
	usbmux := ios.DeviceList{DeviceList: []ios.DeviceEntry{
		{DeviceID: 7, Properties: ios.DeviceProperties{SerialNumber: "shared-udid", ConnectionType: "USB"}},
	}}
	tunnelDevices := []ios.DeviceEntry{
		{Properties: ios.DeviceProperties{SerialNumber: "shared-udid", ConnectionType: connectionTypeUserspaceTunnel}},
		{Properties: ios.DeviceProperties{SerialNumber: "tunnel-only-udid", ConnectionType: connectionTypeTunnel}},
	}

	merged := mergeTunnelDevices(usbmux, tunnelDevices)
	if len(merged.DeviceList) != 2 {
		t.Fatalf("expected 2 devices, got %d", len(merged.DeviceList))
	}
	if merged.DeviceList[0].Properties.ConnectionType != "USB" || merged.DeviceList[0].DeviceID != 7 {
		t.Errorf("usbmuxd entry must win for devices known to both sources: %+v", merged.DeviceList[0])
	}
	if merged.DeviceList[1].Properties.SerialNumber != "tunnel-only-udid" {
		t.Errorf("tunnel-only device missing from merge: %+v", merged.DeviceList[1])
	}
	if isTunnelOnlyDevice(merged.DeviceList[0]) {
		t.Error("usbmuxd entry must not be marked tunnel-only")
	}
}

// TestDirectTargetDeviceBypassesUsbmuxd covers the --address/--rsd-port path:
// resolution must produce a usable entry with the udid as identity metadata
// even when usbmuxd does not know the device.
func TestDirectTargetDeviceBypassesUsbmuxd(t *testing.T) {
	device := directTargetDevice("00008120-001905DE2E9B401E", "fd6f:7edb:be57::1", true)
	if device.Properties.SerialNumber != "00008120-001905DE2E9B401E" {
		t.Errorf("udid must be kept as identity metadata, got %q", device.Properties.SerialNumber)
	}
	if device.Address != "fd6f:7edb:be57::1" {
		t.Errorf("unexpected address: %s", device.Address)
	}
	if device.Properties.ConnectionType != connectionTypeUserspaceTunnel {
		t.Errorf("expected connectionType %q, got %q", connectionTypeUserspaceTunnel, device.Properties.ConnectionType)
	}

	kernelTarget := directTargetDevice("", "fd6f:7edb:be57::1", false)
	if kernelTarget.Properties.ConnectionType != connectionTypeTunnel {
		t.Errorf("expected connectionType %q, got %q", connectionTypeTunnel, kernelTarget.Properties.ConnectionType)
	}
	if kernelTarget.Properties.SerialNumber != "" {
		t.Errorf("empty udid must stay empty until the RSD handshake provides one, got %q", kernelTarget.Properties.SerialNumber)
	}
}

// TestPrintDeviceListIncludesTunnelDevices asserts the merged `ios list`
// output contains devices only known to the tunnel agent, regardless of
// whether a local usbmuxd is running.
func TestPrintDeviceListIncludesTunnelDevices(t *testing.T) {
	tunnelInfo := fakeTunnelAgent(t, []tunnel.Tunnel{
		{Udid: "tunnel-only-test-udid", Address: "fd6f:7edb:be57::1", RsdPort: 62173, UserspaceTUN: true, UserspaceTUNPort: 61246},
	})

	output := captureStdout(t, func() {
		printDeviceList(false, tunnelInfo)
	})
	if !strings.Contains(output, "tunnel-only-test-udid") {
		t.Errorf("list output must contain the tunnel-backed device, got: %s", output)
	}
}

// TestDetailedListMarksTunnelTransport asserts the --details output labels
// tunnel-backed devices with their transport instead of failing on the
// unreachable lockdown connection.
func TestDetailedListMarksTunnelTransport(t *testing.T) {
	deviceList := ios.DeviceList{DeviceList: []ios.DeviceEntry{
		{
			Properties:       ios.DeviceProperties{SerialNumber: "tunnel-only-test-udid", ConnectionType: connectionTypeUserspaceTunnel},
			Address:          "fd6f:7edb:be57::1",
			UserspaceTUN:     true,
			UserspaceTUNPort: 61246,
		},
	}}

	output := captureStdout(t, func() {
		outputDetailedList(deviceList)
	})

	var parsed struct {
		DeviceList []detailsEntry `json:"deviceList"`
	}
	if err := json.Unmarshal([]byte(output), &parsed); err != nil {
		t.Fatalf("details output is not valid JSON: %v, output: %s", err, output)
	}
	if len(parsed.DeviceList) != 1 {
		t.Fatalf("expected 1 device, got %d", len(parsed.DeviceList))
	}
	if parsed.DeviceList[0].Udid != "tunnel-only-test-udid" {
		t.Errorf("unexpected udid: %s", parsed.DeviceList[0].Udid)
	}
	if parsed.DeviceList[0].ConnectionType != connectionTypeUserspaceTunnel {
		t.Errorf("expected connectionType %q, got %q", connectionTypeUserspaceTunnel, parsed.DeviceList[0].ConnectionType)
	}
}

func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	original := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = original }()

	done := make(chan string)
	go func() {
		data, _ := io.ReadAll(reader)
		done <- string(data)
	}()

	f()
	writer.Close()
	return <-done
}
