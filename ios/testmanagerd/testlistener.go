package testmanagerd

import (
	"io"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
)

// Realtime test callbacks
type TestListener interface {
	DidBeginExecutingTestPlan()
	DidFinishExecutingTestPlan()
	InitializationForUITestingFailed(err nskeyedarchiver.NSError)
	TestCaseStalled(testCase string, method string, file string, line uint64)
	TestCaseFailedForClass(testClass string, method string, message string, file string, line uint64)
	TestCaseDidFinishForTest(testClass string, testMethod string, status string, duration float64)
	TestCaseFinished(testCase string, method string, xcActivityRecord nskeyedarchiver.XCActivityRecord)
	TestSuiteFinished(suiteName string, date string, testCount uint64, failures uint64, unexpected uint64, testDuration float64,
		totalDuration float64)
	TestCaseDidStartForClass(className string, methodName string)
	TestRunnerKilled()
	LogMessage(msg string)
	LogDebugMessage(msg string)
}

// A concrete implementation of TestListener that writes all test logs to given Writer and ignores all other messages
type TestLogCollector struct {
	Writer *io.Writer
}

// Forward logs to writer
func (t TestLogCollector) LogDebugMessage(msg string) {
	(*t.Writer).Write([]byte(msg))
}
func (t TestLogCollector) LogMessage(msg string) {
	(*t.Writer).Write([]byte(msg))
}

// Ignore other messages
func (t TestLogCollector) DidBeginExecutingTestPlan()                                               {}
func (t TestLogCollector) DidFinishExecutingTestPlan()                                              {}
func (t TestLogCollector) InitializationForUITestingFailed(err nskeyedarchiver.NSError)             {}
func (t TestLogCollector) TestCaseStalled(testCase string, method string, file string, line uint64) {}
func (t TestLogCollector) TestCaseFailedForClass(testClass string, method string, message string, file string, line uint64) {
}
func (t TestLogCollector) TestCaseDidFinishForTest(testClass string, testMethod string, status string, duration float64) {
}
func (t TestLogCollector) TestCaseFinished(testCase string, method string, xcActivityRecord nskeyedarchiver.XCActivityRecord) {
}
func (t TestLogCollector) TestSuiteFinished(suiteName string, date string, testCount uint64, failures uint64, unexpected uint64, testDuration float64,
	totalDuration float64) {
}
func (t TestLogCollector) TestCaseDidStartForClass(className string, methodName string) {}
func (t TestLogCollector) TestRunnerKilled()                                            {}
