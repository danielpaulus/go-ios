package forward

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

// deadPid returns a pid that is guaranteed not to be running anymore: it
// spawns the test binary with a pattern matching no tests, waits for it to
// exit and returns its pid.
func deadPid(t *testing.T) int {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestNothingMatchesThisName$")
	if err := cmd.Run(); err != nil {
		t.Fatalf("spawn helper process: %v", err)
	}
	return cmd.Process.Pid
}

func TestRegistryRegisterListUnregister(t *testing.T) {
	r := NewRegistry(t.TempDir())

	e1 := Entry{Udid: "udid-a", HostPort: 8100, DevicePort: 9100, Pid: os.Getpid()}
	e2 := Entry{Udid: "udid-b", HostPort: 8200, DevicePort: 9200, Pid: os.Getpid()}
	// Register in reverse order to verify List sorts by udid/hostPort.
	if err := r.Register(e2); err != nil {
		t.Fatalf("register e2: %v", err)
	}
	if err := r.Register(e1); err != nil {
		t.Fatalf("register e1: %v", err)
	}

	got, err := r.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if want := []Entry{e1, e2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("list: got %+v want %+v", got, want)
	}

	if err := r.Unregister(e1); err != nil {
		t.Fatalf("unregister e1: %v", err)
	}
	got, err = r.List()
	if err != nil {
		t.Fatalf("list after unregister: %v", err)
	}
	if want := []Entry{e2}; !reflect.DeepEqual(got, want) {
		t.Fatalf("list after unregister: got %+v want %+v", got, want)
	}

	// Unregistering an already-removed entry must not error (prune races).
	if err := r.Unregister(e1); err != nil {
		t.Fatalf("unregister e1 twice: %v", err)
	}
}

func TestRegistryPrunesDeadPids(t *testing.T) {
	dir := t.TempDir()
	r := NewRegistry(dir)

	alive := Entry{Udid: "udid-alive", HostPort: 8100, DevicePort: 9100, Pid: os.Getpid()}
	dead := Entry{Udid: "udid-dead", HostPort: 8200, DevicePort: 9200, Pid: deadPid(t)}
	if err := r.Register(alive); err != nil {
		t.Fatalf("register alive: %v", err)
	}
	if err := r.Register(dead); err != nil {
		t.Fatalf("register dead: %v", err)
	}

	got, err := r.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if want := []Entry{alive}; !reflect.DeepEqual(got, want) {
		t.Fatalf("list: got %+v want %+v", got, want)
	}

	// The dead entry's file must be pruned from the state dir, not just hidden.
	if _, err := os.Stat(r.entryPath(dead)); !os.IsNotExist(err) {
		t.Fatalf("dead entry file not pruned, stat err: %v", err)
	}
	if _, err := os.Stat(r.entryPath(alive)); err != nil {
		t.Fatalf("alive entry file must survive prune: %v", err)
	}
}

func TestRegistryListMissingDirAndGarbage(t *testing.T) {
	dir := t.TempDir()
	// Listing a registry whose dir was never created returns empty, no error.
	r := NewRegistry(filepath.Join(dir, "never-created"))
	got, err := r.List()
	if err != nil {
		t.Fatalf("list missing dir: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("list missing dir: got %+v want empty", got)
	}

	// Non-entry files are ignored, invalid JSON entries are pruned.
	r = NewRegistry(dir)
	if err := os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	garbage := filepath.Join(dir, "1234-1.json")
	if err := os.WriteFile(garbage, []byte("{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err = r.List()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("list: got %+v want empty", got)
	}
	if _, err := os.Stat(garbage); !os.IsNotExist(err) {
		t.Fatalf("garbage entry not pruned, stat err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "notes.txt")); err != nil {
		t.Fatalf("non-json file must be left alone: %v", err)
	}
}
