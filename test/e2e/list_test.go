//go:build e2e

package e2e_test

import (
	"encoding/json"
	"slices"
	"testing"
)

func TestList(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		out := runIOS(t, "list")

		var v struct {
			DeviceList []string `json:"deviceList"`
		}
		if err := json.Unmarshal(out, &v); err != nil {
			t.Fatalf("parse: %v\n%s", err, out)
		}
		if !slices.Contains(v.DeviceList, udid) {
			t.Fatalf("device %s not present in list: %v", udid, v.DeviceList)
		}
	})
}
