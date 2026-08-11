package tunnel

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestRunCmdHelperProcess is not a real test: TestRunCmdFailureIncludesCommandAndOutput
// re-executes the test binary with GO_IOS_WANT_RUNCMD_HELPER set so this helper acts
// like netsh on Windows, which prints its error message to stdout (not stderr) and
// exits non-zero.
func TestRunCmdHelperProcess(t *testing.T) {
	if os.Getenv("GO_IOS_WANT_RUNCMD_HELPER") != "1" {
		t.Skip("helper process for TestRunCmdFailureIncludesCommandAndOutput")
	}
	fmt.Println("The parameter is incorrect.")
	os.Exit(1)
}

func TestRunCmdFailureIncludesCommandAndOutput(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestRunCmdHelperProcess$", "-test.v")
	cmd.Env = append(os.Environ(), "GO_IOS_WANT_RUNCMD_HELPER=1")

	err := runCmd(cmd)
	if err == nil {
		t.Fatal("expected runCmd to fail")
	}
	msg := err.Error()
	if !strings.Contains(msg, "The parameter is incorrect.") {
		t.Errorf("error does not contain the command's stdout output: %s", msg)
	}
	if !strings.Contains(msg, cmd.String()) {
		t.Errorf("error does not contain the executed command line %q: %s", cmd.String(), msg)
	}
	if !strings.Contains(msg, "exit status 1") {
		t.Errorf("error does not contain the exit status: %s", msg)
	}
}

func TestRunCmdSuccess(t *testing.T) {
	cmd := exec.Command(os.Args[0], "-test.run=TestRunCmdHelperProcess$")
	// Without GO_IOS_WANT_RUNCMD_HELPER the helper test skips, so the command
	// succeeds.
	if err := runCmd(cmd); err != nil {
		t.Fatalf("expected runCmd to succeed, got: %v", err)
	}
}
