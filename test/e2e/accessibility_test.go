//go:build e2e

package e2e_test

import "testing"

func TestAssistivetouchGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "assistivetouch", "get") })
}

func TestVoiceoverGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "voiceover", "get") })
}

func TestZoomGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "zoom", "get") })
}

func TestTimeformatGet(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) { smoke(t, udid, "timeformat", "get") })
}
