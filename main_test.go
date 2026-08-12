package main_test

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var (
	update = flag.Bool("update", false, "update golden files")
	e2e    = flag.Bool("e2e", false, "test with realdevice")
)

func TestDeviceList(t *testing.T) {
	if !*e2e {
		return
	}
	output, err := exec.Command("go", "run", "./ios.go", "list").Output()
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println(string(output))
}

func TestHelp_GlobalNoArgs(t *testing.T) {
	stdout, stderr, code := runCLI(t)
	if code != 0 {
		t.Fatalf("exit code = %d, stderr = %q", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	assertGolden(t, "global.golden", stdout)
}

func TestHelp_GlobalFlagsEquivalent(t *testing.T) {
	stdoutLong, stderrLong, codeLong := runCLI(t, "--help")
	if codeLong != 0 || stderrLong != "" {
		t.Fatalf("--help failed: code=%d stderr=%q", codeLong, stderrLong)
	}
	stdoutShort, stderrShort, codeShort := runCLI(t, "-h")
	if codeShort != 0 || stderrShort != "" {
		t.Fatalf("-h failed: code=%d stderr=%q", codeShort, stderrShort)
	}
	if eolFormat(stdoutLong) != eolFormat(stdoutShort) {
		t.Fatalf("help outputs differ\n--help:\n%s\n-h:\n%s", stdoutLong, stdoutShort)
	}
}

func TestHelp_SubcommandExplicitImplicitAndGlobalFlagOrdering(t *testing.T) {
	explicit, explicitErr, explicitCode := runCLI(t, "help", "apps")
	if explicitCode != 0 || explicitErr != "" {
		t.Fatalf("help apps failed: code=%d stderr=%q", explicitCode, explicitErr)
	}
	implicit, implicitErr, implicitCode := runCLI(t, "apps", "--help")
	if implicitCode != 0 || implicitErr != "" {
		t.Fatalf("apps --help failed: code=%d stderr=%q", implicitCode, implicitErr)
	}
	implicitShort, implicitShortErr, implicitShortCode := runCLI(t, "apps", "-h")
	if implicitShortCode != 0 || implicitShortErr != "" {
		t.Fatalf("apps -h failed: code=%d stderr=%q", implicitShortCode, implicitShortErr)
	}
	withGlobal, withGlobalErr, withGlobalCode := runCLI(t, "--udid=abc", "help", "apps")
	if withGlobalCode != 0 || withGlobalErr != "" {
		t.Fatalf("--udid=abc help apps failed: code=%d stderr=%q", withGlobalCode, withGlobalErr)
	}
	normalizedExplicit := eolFormat(explicit)
	if normalizedExplicit != eolFormat(implicit) {
		t.Fatalf("explicit and implicit help differ")
	}
	if normalizedExplicit != eolFormat(implicitShort) {
		t.Fatalf("explicit and short implicit help differ")
	}
	if normalizedExplicit != eolFormat(withGlobal) {
		t.Fatalf("explicit and global-flag help differ")
	}
	assertGolden(t, "apps.golden", explicit)
}

func TestHelp_InstrumentsSubcommands(t *testing.T) {
	fps, fpsErr, fpsCode := runCLI(t, "help", "instruments", "fps")
	if fpsCode != 0 || fpsErr != "" {
		t.Fatalf("help instruments fps failed: code=%d stderr=%q", fpsCode, fpsErr)
	}
	assertGolden(t, "instruments_fps.golden", fps)

	network, networkErr, networkCode := runCLI(t, "help", "instruments", "network")
	if networkCode != 0 || networkErr != "" {
		t.Fatalf("help instruments network failed: code=%d stderr=%q", networkCode, networkErr)
	}
	assertGolden(t, "instruments_network.golden", network)
}

func TestHelp_NestedSubcommandAndUnknown(t *testing.T) {
	explicit, explicitErr, explicitCode := runCLI(t, "help", "tunnel", "start")
	if explicitCode != 0 || explicitErr != "" {
		t.Fatalf("help tunnel start failed: code=%d stderr=%q", explicitCode, explicitErr)
	}
	implicit, implicitErr, implicitCode := runCLI(t, "tunnel", "start", "--help")
	if implicitCode != 0 || implicitErr != "" {
		t.Fatalf("tunnel start --help failed: code=%d stderr=%q", implicitCode, implicitErr)
	}
	if eolFormat(explicit) != eolFormat(implicit) {
		t.Fatalf("nested explicit and implicit help differ")
	}

	stdout, stderr, code := runCLI(t, "help", "nope")
	if code == 0 {
		t.Fatalf("expected non-zero exit for unknown topic")
	}
	if stdout != "" {
		t.Fatalf("stdout should be empty for unknown topic, got %q", stdout)
	}
	if !strings.Contains(stderr, "unknown help topic") {
		t.Fatalf("missing unknown topic message: %q", stderr)
	}
}

func runCLI(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"run", "."}, args...)...)
	cmd.Dir = "."
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	code := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			code = exitErr.ExitCode()
		} else {
			t.Fatalf("run cli: %v", err)
		}
	}
	return stdout.String(), stderr.String(), code
}

func assertGolden(t *testing.T, goldenName string, got string) {
	t.Helper()
	path := filepath.Join("testdata", "help", goldenName)
	got = eolFormat(got)
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("mkdir golden dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden: %v", err)
		}
	}
	wantBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s: %v", goldenName, err)
	}
	want := eolFormat(string(wantBytes))
	if got != want {
		t.Fatalf("golden mismatch for %s\n--- got ---\n%s\n--- want ---\n%s", goldenName, got, want)
	}
}

func eolFormat(s string) string {
	return strings.ReplaceAll(s, "\r\n", "\n")
}
