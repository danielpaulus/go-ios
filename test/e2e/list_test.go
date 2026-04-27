//go:build e2e

package e2e_test

import (
	"encoding/json"
	"testing"
)

func TestList(t *testing.T) {
	out := runIOS(t, "list")

	var v struct {
		DeviceList []string `json:"deviceList"`
	}
	if err := json.Unmarshal(out, &v); err != nil {
		t.Fatalf("parse: %v\n%s", err, out)
	}
	if len(v.DeviceList) == 0 {
		t.Fatal("no devices found")
	}
	t.Logf("devices: %v", v.DeviceList)
}
