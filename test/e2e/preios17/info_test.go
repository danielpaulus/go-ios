//go:build e2e

package preios17_test

import (
	"encoding/json"
	"slices"
	"testing"
	"time"
)

func TestInfo(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{
			"ProductType", "ProductVersion", "BuildVersion", "DeviceClass",
			"CPUArchitecture", "HardwareModel", "ProductName", "ModelNumber",
			"SerialNumber", "UniqueDeviceID", "DeviceName",
		}, "info")
		assertSnapshot(t, udid, m)
	})
}

func TestInfoLockdown(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{
			"BuildVersion", "CPUArchitecture", "DeviceClass", "DeviceName",
			"ProductType", "ProductVersion", "UniqueDeviceID",
		}, "info", "lockdown")
		assertSnapshot(t, udid, m)
	})
}

// TestLockdownGet dumps all lockdown values and asserts the recorded identity
// fields match the snapshot.
func TestLockdownGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{
			"BuildVersion", "CPUArchitecture", "DeviceClass", "DeviceName",
			"ProductType", "ProductVersion", "UniqueDeviceID",
		}, "lockdown", "get")
		if exp, ok := expectedDevice(udid); ok {
			for _, key := range []string{"ProductType", "ProductVersion", "BuildVersion"} {
				if got, _ := m[key].(string); got != exp[key] {
					t.Fatalf("lockdown get %s = %q, want %q", key, got, exp[key])
				}
			}
		}
	})
}

func TestDevicename(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"devicename"}, "devicename")
		if name, _ := m["devicename"].(string); name == "" {
			t.Fatalf("devicename: empty")
		}
	})
}

// TestDate reads the device clock and asserts formatedDate is a well-formed
// RFC850 date (e.g. "Thursday, 20-Mar-25 17:56:04 CET") in the device's own
// timezone, reported as a loadable IANA name.
func TestDate(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"TimeIntervalSince1970", "formatedDate", "TimeZone"}, "date")
		fd, _ := m["formatedDate"].(string)
		if _, err := time.Parse(time.RFC850, fd); err != nil {
			t.Fatalf("date: formatedDate %q is not a parseable date: %v", fd, err)
		}
		tz, _ := m["TimeZone"].(string)
		if _, err := time.LoadLocation(tz); err != nil {
			t.Fatalf("date: TimeZone %q is not a valid IANA timezone: %v", tz, err)
		}
	})
}

func TestLang(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"Language", "Locale", "SupportedLocales"}, "lang")
		if lang, _ := m["Language"].(string); lang == "" {
			t.Fatalf("lang: empty Language")
		}
	})
}

// TestMobilegestalt queries a key via the diagnostics relay. Unlike current iOS
// (where MobileGestalt is deprecated and returns no value), pre-iOS17 devices
// still serve it: assert the relay succeeds and returns the real value.
func TestMobilegestalt(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"Diagnostics", "Status"}, "mobilegestalt", "ProductVersion")
		if status, _ := m["Status"].(string); status != "Success" {
			t.Fatalf("mobilegestalt: Status = %q, want Success", status)
		}
		diag, _ := m["Diagnostics"].(map[string]any)
		mg, _ := diag["MobileGestalt"].(map[string]any)
		if s, _ := mg["Status"].(string); s != "Success" {
			t.Fatalf("mobilegestalt: MobileGestalt.Status = %q, want Success", s)
		}
		if v, _ := mg["ProductVersion"].(string); v == "" {
			t.Fatalf("mobilegestalt: MobileGestalt.ProductVersion is empty")
		}
	})
}

// TestDiagnosticsList asserts the diagnostics relay reports overall success.
func TestDiagnosticsList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		m := smokeObj(t, udid, []string{"Diagnostics", "Status"}, "diagnostics", "list")
		if status, _ := m["Status"].(string); status != "Success" {
			t.Fatalf("diagnostics list: Status = %q, want Success", status)
		}
	})
}

// TestVersion does not target a device: it verifies the built binary runs and
// prints a JSON object with a non-empty version string.
func TestVersion(t *testing.T) {
	var m map[string]any
	if err := json.Unmarshal(runIOS(t, "version"), &m); err != nil {
		t.Fatalf("version: not JSON: %v", err)
	}
	if v, _ := m["version"].(string); v == "" {
		t.Fatalf("version: missing/empty version field in %v", m)
	}
}

func TestList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		var v struct {
			DeviceList []string `json:"deviceList"`
		}
		if err := json.Unmarshal(runIOS(t, "list"), &v); err != nil {
			t.Fatalf("parse: %v", err)
		}
		if !slices.Contains(v.DeviceList, udid) {
			t.Fatalf("device %s not present in list: %v", udid, v.DeviceList)
		}
	})
}

// TestListDetails asserts the device appears in `list --details` with a real
// ConnectionType and identity fields matching the snapshot.
func TestListDetails(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		var v struct {
			DeviceList []map[string]any `json:"deviceList"`
		}
		if err := json.Unmarshal(runIOS(t, "list", "--details"), &v); err != nil {
			t.Fatalf("parse: %v", err)
		}
		exp, known := expectedDevice(udid)
		found := false
		for _, d := range v.DeviceList {
			if u, _ := d["Udid"].(string); u != udid {
				continue
			}
			found = true
			if ct, _ := d["ConnectionType"].(string); ct != "USB" && ct != "Network" {
				t.Fatalf("device %s ConnectionType = %q, want USB or Network", udid, ct)
			}
			if known {
				for _, key := range []string{"ProductName", "ProductType", "ProductVersion"} {
					if got, _ := d[key].(string); got != exp[key] {
						t.Fatalf("%s = %q, want %q (testdata/devices.json)", key, got, exp[key])
					}
				}
			}
		}
		if !found {
			t.Fatalf("device %s not present in list --details", udid)
		}
	})
}
