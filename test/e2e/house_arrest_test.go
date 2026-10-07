//go:build e2e

package e2e_test

import (
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/house_arrest"
	"github.com/danielpaulus/go-ios/ios/installationproxy"
)

// logModuleHouseArrest is used as the "module" attr on golog lines this test
// emits (the const in the house_arrest package is unexported; per AGENTS.md an
// external _test package uses the literal string).
const logModuleHouseArrest = "go-ios/test/e2e/house_arrest"

// appClass classifies an installed app for the purpose of choosing a
// house_arrest.New target.
type appClass struct {
	bundleID        string
	applicationType string
	developerSigned bool // get-task-allow entitlement present -> installed with a developer profile
}

// classifyApp inspects an installationproxy AppInfo. Developer-signed apps (run
// from Xcode / a developer profile) carry a get-task-allow entitlement and
// house_arrest's VendContainer succeeds for them. App Store / user apps do not,
// and the device rejects VendContainer with InstallationLookupFailed — that is
// the case #794's VendDocuments fallback exists to handle.
func classifyApp(a installationproxy.AppInfo) appClass {
	c := appClass{bundleID: a.CFBundleIdentifier()}
	if t, ok := a[installationproxy.ApplicationType].(string); ok {
		c.applicationType = t
	}
	if ent, ok := a[installationproxy.Entitlements].(map[string]any); ok {
		if v, ok := ent["get-task-allow"].(bool); ok && v {
			c.developerSigned = true
		}
	}
	return c
}

// listInstalledApps returns the classification of every user app on the device.
func listInstalledApps(t *testing.T, device ios.DeviceEntry) []appClass {
	t.Helper()
	conn, err := installationproxy.New(device)
	if err != nil {
		t.Fatalf("installationproxy.New: %v", err)
	}
	defer conn.Close()
	apps, err := conn.BrowseUserApps()
	if err != nil {
		t.Fatalf("BrowseUserApps: %v", err)
	}
	out := make([]appClass, 0, len(apps))
	for _, a := range apps {
		out = append(out, classifyApp(a))
	}
	return out
}

// TestHouseArrestVendDocumentsFallback verifies PR #794 on a real device: an App
// Store (user, non-developer-signed) app rejects house_arrest's VendContainer
// with InstallationLookupFailed, and house_arrest.New must fall back to
// VendDocuments and still return a working AFC connection. A developer-signed
// app is exercised too, so the VendContainer happy path is proven not to have
// regressed. If the device has no App Store app installed the fallback cannot be
// exercised and the test skips loudly rather than passing silently.
func TestHouseArrestVendDocumentsFallback(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		device, err := ios.GetDevice(udid)
		if err != nil {
			t.Fatalf("GetDevice %s: %v", udid, err)
		}

		apps := listInstalledApps(t, device)

		// Log the full user-app inventory so the run record shows exactly what
		// was on the device and how each app was classified.
		golog.Info("house_arrest e2e: user app inventory",
			"module", logModuleHouseArrest, "udid", udid, "count", len(apps))
		var appStore, developer *appClass
		for i := range apps {
			a := apps[i]
			golog.Info("house_arrest e2e: installed app",
				"module", logModuleHouseArrest, "udid", udid,
				"bundleID", a.bundleID, "applicationType", a.applicationType,
				"developerSigned", a.developerSigned)
			if a.bundleID == "" {
				continue
			}
			// An App Store app: user-installed and NOT developer-signed — the
			// kind that rejects VendContainer.
			if a.applicationType == "User" && !a.developerSigned && appStore == nil {
				appStore = &apps[i]
			}
			// A developer-signed app for the no-regression VendContainer check.
			if a.developerSigned && developer == nil {
				developer = &apps[i]
			}
		}

		// No-regression check: VendContainer still works for a developer-signed
		// app (when one is installed).
		if developer != nil {
			golog.Info("house_arrest e2e: verifying VendContainer path (developer-signed app)",
				"module", logModuleHouseArrest, "udid", udid, "bundleID", developer.bundleID)
			client, err := house_arrest.New(device, developer.bundleID)
			if err != nil {
				t.Fatalf("house_arrest.New for developer-signed app %s failed (VendContainer regression?): %v", developer.bundleID, err)
			}
			entries, listErr := client.List("/")
			_ = client.Close()
			if listErr != nil {
				t.Fatalf("AFC List(/) for developer-signed app %s: %v", developer.bundleID, listErr)
			}
			golog.Info("house_arrest e2e: VendContainer path OK",
				"module", logModuleHouseArrest, "udid", udid,
				"bundleID", developer.bundleID, "entries", len(entries))
		} else {
			golog.Info("house_arrest e2e: no developer-signed app installed; skipping VendContainer no-regression check",
				"module", logModuleHouseArrest, "udid", udid)
		}

		// Core verification: the VendDocuments fallback.
		if appStore == nil {
			t.Skipf("SKIP: no App Store app installed on %s; cannot verify #794 VendDocuments fallback "+
				"(need a User app WITHOUT the get-task-allow entitlement, i.e. installed from the App Store). "+
				"Install any free App Store app on this device to enable this test.", udid)
			return
		}

		golog.Info("house_arrest e2e: verifying VendDocuments fallback (App Store app)",
			"module", logModuleHouseArrest, "udid", udid, "bundleID", appStore.bundleID)
		client, err := house_arrest.New(device, appStore.bundleID)
		if err != nil {
			t.Fatalf("house_arrest.New for App Store app %s failed — #794 VendDocuments fallback did NOT produce a connection: %v",
				appStore.bundleID, err)
		}
		defer client.Close()

		// A working AFC connection proves the fallback succeeded: list the
		// documents root (read-only, non-destructive). The vend is scoped to the
		// app's Documents directory, so "/" here is that Documents root.
		entries, err := client.List("/")
		if err != nil {
			t.Fatalf("house_arrest.New for App Store app %s returned a connection but AFC List(/) failed — fallback connection not usable: %v",
				appStore.bundleID, err)
		}
		golog.Info("house_arrest e2e: VendDocuments fallback OK — AFC connection is usable",
			"module", logModuleHouseArrest, "udid", udid,
			"bundleID", appStore.bundleID, "documentsEntries", len(entries), "entries", entries)
	})
}
