//go:build e2e

package e2e_test

import "testing"

func TestDevmodeGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "devmode", "get") })
}

func TestLang(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "lang") })
}

func TestDiagnosticsList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smokeJSON(t, udid, "diagnostics", "list") })
}
