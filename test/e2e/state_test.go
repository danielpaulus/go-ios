//go:build e2e

package e2e_test

import "testing"

func TestDevmodeGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "devmode", "get") })
}

func TestLang(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "lang") })
}

func TestDiagnosticsList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "diagnostics", "list") })
}
