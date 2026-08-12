//go:build e2e

package e2e_test

// This test verifies the `ios runtest --junit-output` feature (PR #799 /
// feat/issue-286-junit-report) against a REAL XCTest run on the device farm.
//
// It installs the checked-in kitchen-sink runner (testdata/kitchensink-runner.zip,
// built from testdata/kitchensink/), runs it via `ios runtest --junit-output`,
// dumps the produced report.xml verbatim into the test log (so the real XML is
// visible in CI), and then asserts the JUnit mapping is correct: two suites, the
// right per-test outcomes (pass / fail+message+file:line / expected-failure /
// skip), sane aggregate counts, nonzero durations, and an ISO8601 suite
// timestamp.
//
// Like the WDA/DeviceKit tests it needs the shared signing identity
// (GO_IOS_E2E_SIGNING_P12_B64 + GO_IOS_E2E_SIGNING_CERT_ID) plus the App Store
// Connect credentials to mint a per-bundle profile; provisionSigningAssets skips
// the test when they are absent.

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/test/e2e/harness"
)

const (
	kitchenSinkRunnerBundleID = "com.deviceboxhq.goios.kitchensink.xctrunner"
	kitchenSinkXctestConfig   = "KitchenSinkUITests.xctest"
	kitchenSinkRunnerApp      = "KitchenSinkUITests-Runner.app"
	kitchenSinkRunnerZip      = "testdata/kitchensink-runner.zip"
)

// Local structs mirroring ios/junit's XML shape, so the assertions parse exactly
// what the CLI wrote without importing the (unexported) internal types.
type ksTestSuites struct {
	XMLName  xml.Name      `xml:"testsuites"`
	Tests    int           `xml:"tests,attr"`
	Failures int           `xml:"failures,attr"`
	Errors   int           `xml:"errors,attr"`
	Skipped  int           `xml:"skipped,attr"`
	Time     string        `xml:"time,attr"`
	Suites   []ksTestSuite `xml:"testsuite"`
}

type ksTestSuite struct {
	Name      string       `xml:"name,attr"`
	Tests     int          `xml:"tests,attr"`
	Failures  int          `xml:"failures,attr"`
	Errors    int          `xml:"errors,attr"`
	Skipped   int          `xml:"skipped,attr"`
	Time      string       `xml:"time,attr"`
	Timestamp string       `xml:"timestamp,attr"`
	Cases     []ksTestCase `xml:"testcase"`
}

type ksTestCase struct {
	ClassName string    `xml:"classname,attr"`
	Name      string    `xml:"name,attr"`
	Time      string    `xml:"time,attr"`
	Failure   *ksResult `xml:"failure"`
	Error     *ksResult `xml:"error"`
	Skipped   *ksResult `xml:"skipped"`
}

type ksResult struct {
	Message string `xml:"message,attr"`
	Content string `xml:",chardata"`
}

func TestKitchenSinkJUnitOutput(t *testing.T) {
	forEachDevice(t, func(t *testing.T, udid string) {
		// Serialize with WDA/DeviceKit: only one XCUITest runner may drive a
		// device at a time. Reuse the WDA mutex so this never races them.
		wdaMu.Lock()
		defer wdaMu.Unlock()

		runnerApp := prepareKitchenSinkRunner(t)

		// Provision + re-sign + install the unsigned runner, exactly like the
		// DeviceKit sign-and-install e2e. The shared cert + EnsureBundleID
		// auto-registration handle the runner bundle id.
		p12Path, profilePath := provisionSigningAssets(t, udid, kitchenSinkRunnerBundleID, "KitchenSink JUnit E2E")
		signedApp := filepath.Join(t.TempDir(), "kitchensink-runner-signed.app")
		runIOSForDevice(t, udid,
			"sign", "app",
			"--path="+runnerApp,
			"--output="+signedApp,
			"--bundleid="+kitchenSinkRunnerBundleID,
			"--p12file="+p12Path,
			"--profile="+profilePath,
			"--p12password=go-ios-e2e",
			"--install",
		)

		reportPath := filepath.Join(t.TempDir(), "report.xml")
		// No --bundle-id: the .xctest is packaged inside the runner, so the runner
		// is its own test host and there is no separate app-under-test to install.
		// (Passing --bundle-id here would make testmanagerd look up an installed
		// app-under-test and fail, producing an empty report.) The kitchen-sink
		// tests are pure asserts and never drive XCUIApplication, so no target app
		// is needed.
		runIOSForDevice(t, udid,
			"runtest",
			"--test-runner-bundle-id="+kitchenSinkRunnerBundleID,
			"--xctest-config="+kitchenSinkXctestConfig,
			"--junit-output="+reportPath,
		)

		raw, err := os.ReadFile(reportPath)
		if err != nil {
			t.Fatalf("read junit report %s: %v", reportPath, err)
		}
		// Dump the ENTIRE produced report so the real XCTest's XML is captured in
		// the CI logs. This is the primary artifact of this test.
		t.Logf("=== BEGIN ios runtest --junit-output report.xml (%d bytes) ===\n%s\n=== END report.xml ===", len(raw), raw)

		assertKitchenSinkReport(t, raw)
	})
}

// prepareKitchenSinkRunner unzips the checked-in runner and returns the path to
// the KitchenSinkUITests-Runner.app inside it.
func prepareKitchenSinkRunner(t *testing.T) string {
	t.Helper()
	zipPath := filepath.Join(harness.RepoRoot(t), kitchenSinkRunnerZip)
	if _, err := os.Stat(zipPath); err != nil {
		t.Fatalf("kitchen-sink runner zip missing at %s: %v", zipPath, err)
	}
	dir := filepath.Join(t.TempDir(), "kitchensink")
	unzip(t, zipPath, dir)
	appPath := findFirstApp(t, dir)
	if appPath == "" {
		t.Fatalf("no .app found in %s", zipPath)
	}
	if base := filepath.Base(appPath); base != kitchenSinkRunnerApp {
		t.Logf("note: runner app is named %q (expected %q)", base, kitchenSinkRunnerApp)
	}
	return appPath
}

func assertKitchenSinkReport(t *testing.T, raw []byte) {
	t.Helper()

	// Well-formed XML with the <?xml prolog.
	if !strings.HasPrefix(strings.TrimSpace(string(raw)), "<?xml") {
		t.Errorf("report does not start with the <?xml prolog: %q", firstBytes(raw, 40))
	}
	var report ksTestSuites
	if err := xml.Unmarshal(raw, &report); err != nil {
		t.Fatalf("report is not well-formed JUnit XML: %v", err)
	}

	// Two suites: OutcomeTests + SecondSuiteTests.
	suiteByName := map[string]ksTestSuite{}
	for _, s := range report.Suites {
		suiteByName[s.Name] = s
	}
	outcome, ok := suiteByName["OutcomeTests"]
	if !ok {
		t.Fatalf("missing <testsuite name=\"OutcomeTests\">; got suites %v", suiteNames(report))
	}
	second, ok := suiteByName["SecondSuiteTests"]
	if !ok {
		t.Fatalf("missing <testsuite name=\"SecondSuiteTests\">; got suites %v", suiteNames(report))
	}

	caseByName := func(s ksTestSuite) map[string]ksTestCase {
		m := map[string]ksTestCase{}
		for _, c := range s.Cases {
			m[c.Name] = c
		}
		return m
	}
	outcomeCases := caseByName(outcome)
	secondCases := caseByName(second)

	// testPasses / testAlsoPasses => plain testcase, no child element.
	assertPassed(t, outcomeCases, "testPasses")
	assertPassed(t, secondCases, "testAlsoPasses")

	// testFails => <failure> with the message and a *.swift:<line> chardata.
	if c, ok := outcomeCases["testFails"]; !ok {
		t.Errorf("OutcomeTests has no testFails case")
	} else if c.Failure == nil {
		t.Errorf("testFails: expected a <failure> child, got failure=nil (skipped=%v error=%v)", c.Skipped, c.Error)
	} else {
		if !strings.Contains(c.Failure.Message, "kitchen sink failure message") {
			t.Errorf("testFails: <failure message> missing marker; got %q", c.Failure.Message)
		}
		if !regexp.MustCompile(`\.swift:\d+`).MatchString(c.Failure.Content) {
			t.Errorf("testFails: <failure> chardata has no *.swift:<line>; got %q", c.Failure.Content)
		}
	}

	// testExpectedFailure => <skipped message="expected failure...">.
	if c, ok := outcomeCases["testExpectedFailure"]; !ok {
		t.Errorf("OutcomeTests has no testExpectedFailure case")
	} else if c.Skipped == nil {
		t.Errorf("testExpectedFailure: expected a <skipped> child, got skipped=nil (failure=%v error=%v)", c.Failure, c.Error)
	} else if !strings.HasPrefix(strings.ToLower(c.Skipped.Message), "expected failure") {
		t.Errorf("testExpectedFailure: <skipped message> should start with \"expected failure\"; got %q", c.Skipped.Message)
	}

	// testSkipped (throw XCTSkip("skipped on purpose")) => <skipped>.
	//
	// Real-device finding: testmanagerd delivers XCTSkip'd tests with the bare
	// status "skipped" and does NOT propagate the skip reason string, so the real
	// report is <skipped message="skipped"> — the "skipped on purpose" text never
	// leaves the device. (Contrast expected failures above, whose "expected
	// failure" status DOES come through.) go-ios's junit mapping is correct given
	// the data: its default branch falls back to the status word "skipped" as the
	// message. So we assert the element + a "skipped" marker, not the reason.
	if c, ok := outcomeCases["testSkipped"]; !ok {
		t.Errorf("OutcomeTests has no testSkipped case")
	} else if c.Skipped == nil {
		t.Errorf("testSkipped: expected a <skipped> child, got skipped=nil (failure=%v error=%v)", c.Failure, c.Error)
	} else if !strings.Contains(strings.ToLower(c.Skipped.Message+c.Skipped.Content), "skip") {
		t.Errorf("testSkipped: <skipped> should carry a skip marker; got message=%q content=%q", c.Skipped.Message, c.Skipped.Content)
	}

	// testWithAttachment => passed, report intact (attachments must not break it).
	assertPassed(t, outcomeCases, "testWithAttachment")

	// Aggregates: tests == sum over suites, at least one failure, at least two
	// skipped (testExpectedFailure + testSkipped, plus testStalls which is
	// XCTSkip'd while disabled).
	sumTests := 0
	for _, s := range report.Suites {
		sumTests += len(s.Cases)
	}
	if report.Tests != sumTests {
		t.Errorf("<testsuites tests=%d> != sum of testcases %d", report.Tests, sumTests)
	}
	if report.Failures < 1 {
		t.Errorf("expected >=1 failure across the report, got %d", report.Failures)
	}
	if report.Skipped < 2 {
		t.Errorf("expected >=2 skipped across the report, got %d", report.Skipped)
	}

	// At least one passed testcase reports a nonzero, float-parseable duration.
	if !anyPassedNonzeroDuration(report) {
		t.Errorf("no passed testcase had a nonzero, float-parseable time= attribute")
	}

	// Suite timestamp is ISO8601-ish (YYYY-MM-DDTHH:MM:SS).
	tsRe := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`)
	if !tsRe.MatchString(outcome.Timestamp) {
		t.Errorf("OutcomeTests timestamp %q does not match ISO8601 %s", outcome.Timestamp, tsRe)
	}
}

func assertPassed(t *testing.T, cases map[string]ksTestCase, name string) {
	t.Helper()
	c, ok := cases[name]
	if !ok {
		t.Errorf("missing passing testcase %q", name)
		return
	}
	if c.Failure != nil || c.Error != nil || c.Skipped != nil {
		t.Errorf("%s: expected a plain (passed) testcase, got failure=%v error=%v skipped=%v", name, c.Failure, c.Error, c.Skipped)
	}
}

func anyPassedNonzeroDuration(report ksTestSuites) bool {
	for _, s := range report.Suites {
		for _, c := range s.Cases {
			if c.Failure != nil || c.Error != nil || c.Skipped != nil {
				continue
			}
			if v, err := strconv.ParseFloat(strings.TrimSpace(c.Time), 64); err == nil && v > 0 {
				return true
			}
		}
	}
	return false
}

func suiteNames(report ksTestSuites) []string {
	names := make([]string, 0, len(report.Suites))
	for _, s := range report.Suites {
		names = append(names, s.Name)
	}
	return names
}

func firstBytes(b []byte, n int) string {
	if len(b) > n {
		return string(b[:n])
	}
	return string(b)
}
