package testmanagerd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// TestListener collects test results from the test execution
type TestListener struct {
	finished             chan struct{}
	finishedOnce         sync.Once
	err                  error
	logWriter            io.Writer
	debugLogWriter       io.Writer
	attachmentsDirectory string
	TestSuites           []TestSuite
	runningTestSuite     *TestSuite
}

type TestSuite struct {
	Name          string
	StartDate     time.Time
	EndDate       time.Time
	TestDuration  time.Duration
	TotalDuration time.Duration
	TestCases     []TestCase
}

type TestCase struct {
	ClassName   string
	MethodName  string
	Status      TestCaseStatus
	Err         TestError
	Duration    time.Duration
	Attachments []TestAttachment
}

type TestCaseStatus string

const (
	StatusFailed          = TestCaseStatus("failed")           // Defined by Apple
	StatusPassed          = TestCaseStatus("passed")           // Defined by Apple
	StatusExpectedFailure = TestCaseStatus("expected failure") // Defined by Apple
	StatusStalled         = TestCaseStatus("stalled")          // Defined by us

	// Test suite counter constants
	unknownCount uint64 = 0
)

type TestError struct {
	Message string
	File    string
	Line    uint64
}

type TestAttachment struct {
	Name                  string
	Path                  string
	Type                  string
	Timestamp             float64
	Activity              string
	UniformTypeIdentifier string
}

func NewTestListener(logWriter io.Writer, debugLogWriter io.Writer, attachmentsDirectory string) *TestListener {
	return &TestListener{
		finished:             make(chan struct{}),
		logWriter:            logWriter,
		debugLogWriter:       debugLogWriter,
		TestSuites:           make([]TestSuite, 0),
		attachmentsDirectory: attachmentsDirectory,
	}
}

func (t *TestListener) didFinishExecutingTestPlan() {
	t.executionFinished()
}

func (t *TestListener) initializationForUITestingDidFailWithError(err nskeyedarchiver.NSError) {
	t.err = err
	t.executionFinished()
}

func (t *TestListener) didFailToBootstrapWithError(err nskeyedarchiver.NSError) {
	t.err = err
	t.executionFinished()
}

func (t *TestListener) testCaseStalled(testClass string, method string, file string, line uint64) {
	testCase := t.findTestCase(testClass, method)
	if testCase != nil {
		testCase.Status = StatusStalled
		testCase.Err = TestError{
			Message: "Test case stalled",
			File:    file,
			Line:    line,
		}
	}
}

func (t *TestListener) testCaseFinished(testClass string, testMethod string, xcActivityRecord nskeyedarchiver.XCActivityRecord) {
	ts := t.findTestSuite(testClass)
	testCase := t.findTestCase(testClass, testMethod)
	if ts == nil || testCase == nil || testClass == "none" || testMethod == "none" {
		// Attachments of activity records are reported under a special test class named "none"
		// That's unfortunately the default behavior defined by Apple.
		// This if block is a safe guard to auto correct the test case information
		ts = t.runningTestSuite
		if len(ts.TestCases) == 0 {
			log.Debug(fmt.Sprintf("Received testCaseFinished for %s:%s without initialization", testClass, testMethod))
			return
		}
		testCase = &ts.TestCases[len(ts.TestCases)-1]
	}

	for _, attachment := range xcActivityRecord.Attachments {
		attachmentsPath := filepath.Join(t.attachmentsDirectory, uuid.New().String())
		file, err := os.Create(attachmentsPath)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "attachment": attachment.Name}).Warn("Received testCaseFinished with activity record but failed writing attachments to disk. Ignoring attachment")
			continue
		}
		defer file.Close()

		file.Write(attachment.Payload)
		testCase.Attachments = append(testCase.Attachments, TestAttachment{
			Name:                  strings.Clone(attachment.Name),
			Timestamp:             attachment.Timestamp,
			Activity:              strings.Clone(xcActivityRecord.Title),
			Path:                  attachmentsPath,
			Type:                  strings.Clone(xcActivityRecord.ActivityType),
			UniformTypeIdentifier: strings.Clone(attachment.UniformTypeIdentifier),
		})
	}
}

func (t *TestListener) testSuiteDidStart(suiteName string, date string) {
	d, err := time.Parse(time.DateTime+" +0000", date)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Cannot parse test suite start date")
		d = time.Now()
	}

	if t.runningTestSuite != nil {
		log.Warn("A new test suite starts running while another one is in progress, finalizing the previous one")
		t.TestSuites = append(t.TestSuites, *t.runningTestSuite)
	}

	t.runningTestSuite = &TestSuite{
		Name:      suiteName,
		StartDate: d,
		TestCases: make([]TestCase, 0),
	}
}

func (t *TestListener) testCaseDidStartForClass(testClass string, testMethod string) {
	// Find the existing test suite or create a new one if not found
	ts := t.findTestSuite(testClass)
	if ts == nil {
		// If no test suite is found and we're not in a running test suite,
		// we should use the runningTestSuite instead of creating a new one.
		// This handles cases where testCaseDidStartForClass is called before
		// testSuiteDidStart, which would otherwise result in a nil pointer dereference.

		if t.runningTestSuite != nil {
			ts = t.runningTestSuite
		} else {
			// Create a new test suite for this class if no running suite exists.
			// We initialize TestCases as an empty slice to avoid potential issues with nil slices.
			d := time.Now()
			newSuite := TestSuite{
				Name:      testClass,
				StartDate: d,
				TestCases: []TestCase{},
			}
			t.TestSuites = append(t.TestSuites, newSuite)
			ts = &t.TestSuites[len(t.TestSuites)-1]
		}
	}

	// Add the test case to the suite
	ts.TestCases = append(ts.TestCases, TestCase{
		ClassName:  testClass,
		MethodName: testMethod,
	})
}

func (t *TestListener) testCaseFailedForClass(testClass string, testMethod string, message string, file string, line uint64) {
	testCase := t.findTestCase(testClass, testMethod)
	if testCase == nil {
		log.Warn("Received failure status for an unknown test, adding it to suite")
		ts := t.findTestSuite(testClass)
		ts.TestCases = append(ts.TestCases, TestCase{
			ClassName:  testClass,
			MethodName: testMethod,
		})
		testCase = &ts.TestCases[len(ts.TestCases)-1]
	}

	testCase.Status = StatusFailed
	testCase.Err = TestError{
		Message: message,
		File:    file,
		Line:    line,
	}
}

func (t *TestListener) testCaseDidFinishForTest(testClass string, testMethod string, status string, duration float64) {
	testCase := t.findTestCase(testClass, testMethod)
	if testCase != nil {
		// We override "failed" status for stalled tests with the value "stalled" to be able to distinguish them later
		if testCase.Status != StatusStalled {
			testCase.Status = TestCaseStatus(status)
		}

		d, err := time.ParseDuration(fmt.Sprintf("%f", duration) + "s")
		if err != nil {
			d = 0
			log.WithFields(log.Fields{"error": err}).Warn("Failed parsing test case duration")
		}

		testCase.Duration = d
	}
}

func (t *TestListener) testSuiteFinished(suiteName string, date string, testCount uint64, failures uint64, skip uint64, expectedFailure uint64, unexpectedFailure uint64, uncaughtException uint64, testDuration float64, totalDuration float64) {
	endDate, err := time.Parse(time.DateTime+" +0000", date)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Cannot parse test suite start date")
		endDate = time.Now()
	}

	ts := t.findTestSuite(suiteName)
	if ts == nil {
		log.Debug(fmt.Sprintf("Received testSuiteFinished for %s without initialization", suiteName))
		return
	}

	ts.EndDate = endDate

	d, err := time.ParseDuration(fmt.Sprintf("%f", testDuration) + "s")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Test duration cannot be parsed")
		d = 0
	}
	ts.TestDuration = d

	d, err = time.ParseDuration(fmt.Sprintf("%f", totalDuration) + "s")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Total duration cannot be parsed")
		d = 0
	}
	ts.TotalDuration = d

	t.TestSuites = append(t.TestSuites, *t.runningTestSuite)
	t.runningTestSuite = nil
}

func (t *TestListener) LogMessage(msg string) {
	t.logWriter.Write([]byte(msg))
}

func (t *TestListener) LogDebugMessage(msg string) {
	t.debugLogWriter.Write([]byte(msg))
}

func (t *TestListener) TestRunnerKilled() {
	t.err = errors.New("Test runner has been explicitly killed.")
	t.executionFinished()
}

func (t *TestListener) FinishWithError(err error) {
	if t.runningTestSuite != nil {
		t.TestSuites = append(t.TestSuites, *t.runningTestSuite)
		t.runningTestSuite = nil
	}
	t.err = err
	t.executionFinished()
}

func (t *TestListener) Done() <-chan struct{} {
	return t.finished
}

func (t *TestListener) findTestCase(className string, methodName string) *TestCase {
	ts := t.findTestSuite(className)

	if ts != nil && len(ts.TestCases) > 0 {
		tc := &ts.TestCases[len(ts.TestCases)-1]
		if tc.ClassName == className && tc.MethodName == methodName {
			return tc
		}
	}

	return nil
}

func (t *TestListener) findTestSuite(className string) *TestSuite {
	if t.runningTestSuite != nil && t.runningTestSuite.Name == className {
		return t.runningTestSuite
	}

	return nil
}

func (t *TestListener) executionFinished() {
	t.finishedOnce.Do(func() {
		close(t.finished)
	})
}

func (t *TestListener) reset() {
	// Reinitialize finished channel to allow signaling again
	t.finished = make(chan struct{})

	// Reset the sync.Once instance so it can be used again
	t.finishedOnce = sync.Once{}

	// Clear error from the previous test run
	t.err = nil

	// Reset test results
	t.TestSuites = nil

	// Clear the reference to the running test suite
	t.runningTestSuite = nil
}
