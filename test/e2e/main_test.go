//go:build e2e

package e2e_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var iosBin string

func TestMain(m *testing.M) {
	root, err := repoRoot()
	if err != nil {
		panic(err)
	}

	dir, err := os.MkdirTemp("", "ios-e2e-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	iosBin = filepath.Join(dir, "ios")
	cmd := exec.Command("go", "build", "-o", iosBin, ".")
	cmd.Dir = root
	if out, err := cmd.CombinedOutput(); err != nil {
		panic("build failed: " + err.Error() + "\n" + string(out))
	}

	os.Exit(m.Run())
}

func repoRoot() (string, error) {
	out, err := exec.Command("go", "env", "GOMOD").Output()
	if err != nil {
		return "", err
	}
	return filepath.Dir(strings.TrimSpace(string(out))), nil
}

// runIOS executes the ios binary with the given args and returns stdout.
// On non-zero exit it fails the test with stderr + stdout for debugging.
func runIOS(t *testing.T, args ...string) []byte {
	t.Helper()
	var stderr bytes.Buffer
	cmd := exec.Command(iosBin, args...)
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("ios %v: %v\nstderr: %s\nstdout: %s", args, err, stderr.String(), out)
	}
	return out
}
