package testmanagerd

import (
	"errors"
	"io"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

// Realtime test callbacks
type TestListener struct {
	testFinishedChannel chan struct{}
	err                 error

	logWriter      io.Writer
	debugLogWriter io.Writer
	testCase       TestCase
}

type TestSuite struct {
	Name            string
	StartDate       string
	EndDate         string
	TestCount       uint64
	FailureCount    uint64
	UnexpectedCount uint64
	TestDuration    float64
	TotalDuration   float64
	TestCases       []TestCase
}

type TestCase struct {
	ClassName  string
	MethodName string
	Status     TestStatus
	Duration   float64
	// TODO : add attachments from xcActivityRecord
}

const (
	// Status
	TestRunning   = "RUNNING"
	TestCompleted = "COMPLETED"

	// Test suite counter constants
	UnknownCount uint64 = 0
)

type TestStatus struct {
	Status  string
	Stalled bool
	Err     *TestError
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

func (t *TestListener) testCaseStalled(testCase string, method string, file string, line uint64) {
	log.Debug("TODO ?")
}

func (t *TestListener) testCaseFinished(testCase string, method string, xcActivityRecord nskeyedarchiver.XCActivityRecord) {
	log.Debug("TODO ?")
}

func (t *TestListener) testSuiteDidStart(suiteName string, date string) {
	log.Debug("TODO 1")
}

func (t *TestListener) testCaseDidStartForClass(className string, methodName string) {
	log.Debug("TODO 2")
}

func (t *TestListener) testCaseFailedForClass(testClass string, method string, message string, file string, line uint64) {
	log.Debug("TODO 3")
}

func (t *TestListener) testCaseDidFinishForTest(testClass string, testMethod string, status string, duration float64) {
	log.Debug("TODO 3.1")
}

func (t *TestListener) testSuiteFinished(suiteName string, date string, testCount uint64, failures uint64, skip uint64, expectedFailure uint64, unexpectedFailure uint64, uncaughtException uint64, testDuration float64, totalDuration float64) {
	log.Debug("TODO 4")
}

func (t *TestListener) LogMessage(msg string) {
	if t.logWriter != nil {
		t.logWriter.Write([]byte(msg))
	}
}

func (t *TestListener) LogDebugMessage(msg string) {
	if t.debugLogWriter != nil {
		t.debugLogWriter.Write([]byte(msg))
	}
}

func (t *TestListener) TestRunnerKilled() {
	t.err = errors.New("Test runner has been explicitly killed.")
	close(t.testFinishedChannel)
}

func (t *TestListener) Done() <-chan struct{} {
	return t.testFinishedChannel
}
