package forward

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"syscall"
)

// Entry describes one active port forward. Each running `ios forward` process
// registers its forwards so `ios forward --list` can enumerate them, similar
// to `adb forward --list`.
type Entry struct {
	Udid       string `json:"udid"`
	HostPort   uint16 `json:"hostPort"`
	DevicePort uint16 `json:"devicePort"`
	Pid        int    `json:"pid"`
}

// Registry is a file-backed store of active forwards. Every forward is one
// JSON file named <pid>-<hostPort>.json in dir, owned by the forwarding
// process. Listing prunes entries whose owning process is gone, so forwards
// that died without cleaning up (SIGKILL, crash) disappear on the next list
// instead of lingering forever.
type Registry struct {
	dir string
}

// NewRegistry returns a Registry backed by dir. The dir is created lazily on
// the first Register.
func NewRegistry(dir string) *Registry {
	return &Registry{dir: dir}
}

// DefaultRegistryDir is the per-user state dir the CLI uses:
// <user config dir>/go-ios/forwards.
func DefaultRegistryDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("forward: could not determine user config dir: %w", err)
	}
	return filepath.Join(base, "go-ios", "forwards"), nil
}

func (r *Registry) entryPath(e Entry) string {
	return filepath.Join(r.dir, fmt.Sprintf("%d-%d.json", e.Pid, e.HostPort))
}

// Register persists e in the registry dir so other processes can list it.
func (r *Registry) Register(e Entry) error {
	if err := os.MkdirAll(r.dir, 0o755); err != nil {
		return fmt.Errorf("forward: could not create registry dir: %w", err)
	}
	data, err := json.Marshal(e)
	if err != nil {
		return fmt.Errorf("forward: could not marshal registry entry: %w", err)
	}
	if err := os.WriteFile(r.entryPath(e), data, 0o644); err != nil {
		return fmt.Errorf("forward: could not write registry entry: %w", err)
	}
	return nil
}

// Unregister removes the entry for e. A missing entry is not an error, so
// Unregister and prune-on-list can race harmlessly.
func (r *Registry) Unregister(e Entry) error {
	err := os.Remove(r.entryPath(e))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("forward: could not remove registry entry: %w", err)
	}
	return nil
}

// List returns all registered forwards whose owning process is still alive,
// sorted by udid, host port, pid. Entries whose process is gone (and files
// that are not valid entries) are pruned from the dir as a side effect.
func (r *Registry) List() ([]Entry, error) {
	files, err := os.ReadDir(r.dir)
	if errors.Is(err, os.ErrNotExist) {
		return []Entry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("forward: could not read registry dir: %w", err)
	}
	entries := make([]Entry, 0, len(files))
	for _, f := range files {
		if f.IsDir() || filepath.Ext(f.Name()) != ".json" {
			continue
		}
		path := filepath.Join(r.dir, f.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var e Entry
		if err := json.Unmarshal(data, &e); err != nil || e.Pid <= 0 {
			os.Remove(path)
			continue
		}
		if !pidAlive(e.Pid) {
			os.Remove(path)
			continue
		}
		entries = append(entries, e)
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Udid != entries[j].Udid {
			return entries[i].Udid < entries[j].Udid
		}
		if entries[i].HostPort != entries[j].HostPort {
			return entries[i].HostPort < entries[j].HostPort
		}
		return entries[i].Pid < entries[j].Pid
	})
	return entries, nil
}

// pidAlive reports whether a process with the given pid exists. On unix,
// signal 0 probes for existence without affecting the process — EPERM still
// means it exists. On Windows os.FindProcess itself fails for gone pids and
// Signal is not usable for probing, so FindProcess succeeding is the answer.
func pidAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return true
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil || errors.Is(err, syscall.EPERM)
}
