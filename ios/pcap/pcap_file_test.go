package pcap

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreatePcapUsesPrivateExclusiveFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "capture.pcap")
	file, err := createPcap(path)
	if err != nil {
		t.Fatalf("createPcap failed: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("close capture: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat capture: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("capture mode = %o, want 600", got)
	}
	if _, err := createPcap(path); err == nil {
		t.Fatal("createPcap unexpectedly replaced an existing file")
	}
}

func TestCreatePcapDoesNotFollowSymlink(t *testing.T) {
	directory := t.TempDir()
	victim := filepath.Join(directory, "victim")
	if err := os.WriteFile(victim, []byte("keep"), 0o600); err != nil {
		t.Fatalf("write victim: %v", err)
	}
	path := filepath.Join(directory, "capture.pcap")
	if err := os.Symlink(victim, path); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	if _, err := createPcap(path); err == nil {
		t.Fatal("createPcap unexpectedly followed a symlink")
	}
	content, err := os.ReadFile(victim)
	if err != nil {
		t.Fatalf("read victim: %v", err)
	}
	if string(content) != "keep" {
		t.Fatalf("victim content = %q, want keep", content)
	}
}

func TestCreateTemporaryPcapUsesUniqueNames(t *testing.T) {
	directory := t.TempDir()
	first, err := createTemporaryPcap(directory)
	if err != nil {
		t.Fatalf("first temporary capture: %v", err)
	}
	defer first.Close()
	second, err := createTemporaryPcap(directory)
	if err != nil {
		t.Fatalf("second temporary capture: %v", err)
	}
	defer second.Close()
	if first.Name() == second.Name() {
		t.Fatalf("temporary captures used the same path %q", first.Name())
	}
}
