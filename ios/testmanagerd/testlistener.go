package testmanagerd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// TestListener collects test results from the test execution
type TestListener struct {
	executionFinished    chan struct{}
	err                  error
	logWriter            io.Writer
	debugLogWriter       io.Writer
	attachmentsDirectory string
	TestSuite            *TestSuite
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
		executionFinished:    make(chan struct{}),
		logWriter:            logWriter,
		debugLogWriter:       debugLogWriter,
		TestSuite:            &TestSuite{},
		attachmentsDirectory: attachmentsDirectory,
	}
}

func (t *TestListener) didFinishExecutingTestPlan() {
	close(t.executionFinished)
}

func (t *TestListener) initializationForUITestingDidFailWithError(err nskeyedarchiver.NSError) {
	t.err = err
	close(t.executionFinished)
}

func (t *TestListener) didFailToBootstrapWithError(err nskeyedarchiver.NSError) {
	t.err = err
	close(t.executionFinished)
}

func (t *TestListener) testCaseStalled(testClass string, method string, file string, line uint64) {
	testCase := t.TestSuite.findTestCase(testClass, method)
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
	for _, attachment := range xcActivityRecord.Attachments {
		testCase := t.TestSuite.findTestCase(testClass, testMethod)
		if testCase == nil {
			t.TestSuite.TestCases = append(t.TestSuite.TestCases, TestCase{
				ClassName:  testClass,
				MethodName: testMethod,
			})
			testCase = &t.TestSuite.TestCases[len(t.TestSuite.TestCases)-1]
		}

		attachmentsPath := filepath.Join(t.attachmentsDirectory, uuid.New().String())
		file, err := os.Create(attachmentsPath)
		if err != nil {
			log.WithFields(log.Fields{"error": err, "attachment": attachment.Name}).Warn("Received testCaseFinished with activity record but failed writing attachments to disk. Ignoring attachment")
			continue
		}
		defer file.Close()

		file.Write(attachment.Payload)
		testCase.Attachments = append(testCase.Attachments, TestAttachment{
			Name:                  attachment.Name,
			Timestamp:             attachment.Timestamp,
			Activity:              xcActivityRecord.Title,
			Path:                  attachmentsPath,
			Type:                  xcActivityRecord.ActivityType,
			UniformTypeIdentifier: attachment.UniformTypeIdentifier,
		})
	}
}

func (t *TestListener) testSuiteDidStart(suiteName string, date string) {
	d, err := time.Parse(time.DateTime+" +0000", date)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Cannot parse test suite start date")
		d = time.Now()
	}

	t.TestSuite = &TestSuite{
		Name:      suiteName,
		StartDate: d,
	}
}

func (t *TestListener) testCaseDidStartForClass(testClass string, testMethod string) {
	t.TestSuite.TestCases = append(t.TestSuite.TestCases, TestCase{
		ClassName:  testClass,
		MethodName: testMethod,
	})
}

func (t *TestListener) testCaseFailedForClass(testClass string, testMethod string, message string, file string, line uint64) {
	testCase := t.TestSuite.findTestCase(testClass, testMethod)
	if testCase == nil {
		log.Warn("Received failure status for an unknown test, adding it to suite")
		t.TestSuite.TestCases = append(t.TestSuite.TestCases, TestCase{
			ClassName:  testClass,
			MethodName: testMethod,
		})
		testCase = &t.TestSuite.TestCases[len(t.TestSuite.TestCases)-1]
	}

	testCase.Status = StatusFailed
	testCase.Err = TestError{
		Message: message,
		File:    file,
		Line:    line,
	}
}

func (t *TestListener) testCaseDidFinishForTest(testClass string, testMethod string, status string, duration float64) {
	testCase := t.TestSuite.findTestCase(testClass, testMethod)
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

	t.TestSuite.EndDate = endDate

	d, err := time.ParseDuration(fmt.Sprintf("%f", testDuration) + "s")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Test duration cannot be parsed")
		d = 0
	}
	t.TestSuite.TestDuration = d

	d, err = time.ParseDuration(fmt.Sprintf("%f", totalDuration) + "s")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Warn("Total duration cannot be parsed")
		d = 0
	}
	t.TestSuite.TotalDuration = d
}

func (t *TestListener) LogMessage(msg string) {
	t.logWriter.Write([]byte(msg))
}

func (t *TestListener) LogDebugMessage(msg string) {
	t.debugLogWriter.Write([]byte(msg))
}

func (t *TestListener) TestRunnerKilled() {
	t.err = errors.New("Test runner has been explicitly killed.")
	close(t.executionFinished)
}

func (t *TestListener) Done() <-chan struct{} {
	return t.executionFinished
}

func (ts *TestSuite) findTestCase(className string, methodName string) *TestCase {
	for i, _ := range ts.TestCases {
		tc := &ts.TestCases[i]
		if tc.ClassName == className && tc.MethodName == methodName {
			return tc
		}
	}

	return nil
}
