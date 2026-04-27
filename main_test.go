//go:build e2e

package main_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var iosBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "ios-e2e-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	iosBin = filepath.Join(dir, "ios")
	if out, err := exec.Command("go", "build", "-o", iosBin, ".").CombinedOutput(); err != nil {
		panic("build failed: " + err.Error() + "\n" + string(out))
	}

	os.Exit(m.Run())
}

func TestList(t *testing.T) {
	out, err := exec.Command(iosBin, "list").Output()
	if err != nil {
		t.Fatalf("ios list: %v", err)
	}

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
