package testmanagerd

import (
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

// Realtime test callbacks and test results
type TestListener struct {
	testFinishedChannel chan struct{}
	err                 error
	logWriter           io.Writer
	debugLogWriter      io.Writer
	testSuite           *TestSuite
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
	ClassName  string
	MethodName string
	Status     TestStatus
	Duration   time.Duration
	// TODO : add attachments from xcActivityRecord
}

const (
	StatusFailed          = "failed"           // Defined by Apple
	StatusPassed          = "passed"           // Defined by Apple
	StatusExpectedFailure = "expected failure" // Defined by Apple
	StatusStalled         = "stalled"          // Defined by us

	// Test suite counter constants
	UnknownCount uint64 = 0
)

type TestStatus struct {
	Status string
	Err    TestError
}

type TestError struct {
	Message string
	File    string
	Line    uint64
}

func NewTestListener(logWriter io.Writer, debugLogWriter io.Writer) *TestListener {
	return &TestListener{
		testFinishedChannel: make(chan struct{}),
		logWriter:           logWriter,
		debugLogWriter:      debugLogWriter,
	}
}

func (t *TestListener) didFinishExecutingTestPlan() {
	close(t.testFinishedChannel)
}

func (t *TestListener) initializationForUITestingDidFailWithError(err nskeyedarchiver.NSError) {
	t.err = err
	close(t.testFinishedChannel)
}

func (t *TestListener) didFailToBootstrapWithError(err nskeyedarchiver.NSError) {
	t.err = err
	close(t.testFinishedChannel)
}

func (t *TestListener) testCaseStalled(testClass string, method string, file string, line uint64) {
	log.Debug("TODO ?")

	testCase := t.testSuite.findTestCase(testClass, method)
	if testCase != nil {
		testCase.Status = TestStatus{
			Status: StatusStalled,
			Err: TestError{
				Message: "Test case stalled",
				File:    file,
				Line:    line,
			},
		}
	}
}

func (t *TestListener) testCaseFinished(testClass string, testMethod string, xcActivityRecord nskeyedarchiver.XCActivityRecord) {
	log.Debug("TODO ?") // Try screenshots
}

func (t *TestListener) testSuiteDidStart(suiteName string, date string) {
	log.Debug("1")

	d, err := time.Parse(time.DateTime+" +0000", date)
	exitIfError("Cannot parse test suite start date", err)

	t.testSuite = &TestSuite{
		Name:      suiteName,
		StartDate: d,
	}
}

func (t *TestListener) testCaseDidStartForClass(testClass string, testMethod string) {
	log.Debug("2")

	t.testSuite.TestCases = append(t.testSuite.TestCases, TestCase{
		ClassName:  testClass,
		MethodName: testMethod,
	})
}

func (t *TestListener) testCaseFailedForClass(testClass string, testMethod string, message string, file string, line uint64) {
	log.Debug("3")

	testCase := t.testSuite.findTestCase(testClass, testMethod)
	if testCase != nil {
		testCase.Status = TestStatus{
			Status: StatusFailed,
			Err: TestError{
				Message: message,
				File:    file,
				Line:    line,
			},
		}
	}
}

func (t *TestListener) testCaseDidFinishForTest(testClass string, testMethod string, status string, duration float64) {
	log.Debug("3.1")

	testCase := t.testSuite.findTestCase(testClass, testMethod)
	if testCase != nil {
		// We override "failed" status for stalled tests with the value "stalled" to be able to distinguish them later
		if testCase.Status.Status == StatusStalled {
			status = StatusStalled
		}

		testCase.Status = TestStatus{
			Status: status,
			Err:    testCase.Status.Err,
		}

		d, err := time.ParseDuration(fmt.Sprintf("%f", duration) + "s")
		exitIfError("Test duration cannot be parsed", err)

		testCase.Duration = d
	}
}

func (t *TestListener) testSuiteFinished(suiteName string, date string, testCount uint64, failures uint64, skip uint64, expectedFailure uint64, unexpectedFailure uint64, uncaughtException uint64, testDuration float64, totalDuration float64) {
	log.Debug("4")

	endDate, err := time.Parse(time.DateTime+" +0000", date)
	exitIfError("Cannot parse test suite start date", err)

	t.testSuite.EndDate = endDate

	d, err := time.ParseDuration(fmt.Sprintf("%f", testDuration) + "s")
	exitIfError("Test duration cannot be parsed", err)
	t.testSuite.TestDuration = d

	d, err = time.ParseDuration(fmt.Sprintf("%f", totalDuration) + "s")
	exitIfError("Test duration cannot be parsed", err)
	t.testSuite.TotalDuration = d
}

func (t *TestListener) LogMessage(msg string) {
	t.logWriter.Write([]byte(msg))
}

func (t *TestListener) LogDebugMessage(msg string) {
	t.debugLogWriter.Write([]byte(msg))
}

func (t *TestListener) TestRunnerKilled() {
	t.err = errors.New("Test runner has been explicitly killed.")
	close(t.testFinishedChannel)
}

func (t *TestListener) Done() <-chan struct{} {
	return t.testFinishedChannel
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

func exitIfError(msg string, err error) {
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Fatalf(msg)
	}
}
