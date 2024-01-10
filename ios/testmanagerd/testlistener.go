package testmanagerd

import (
	"io"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
)

type TestListener interface {
	LogMessage(msg string)
	LogDebugMessage(msg string)
	DidBeginExecutingTestPlan()
	DidFinishExecutingTestPlan()
	DidFailToBootstrapWithError(err nskeyedarchiver.NSError)
	DidBeginInitializingForUITesting()
	GetProgressForLaunch()
	InitializationForUITestingDidFailWithError(err nskeyedarchiver.NSError)
	TestCaseMethodDidFinishActivity()
	TestCaseWithIdentifierDidFinishActivity(testIdentifier nskeyedarchiver.XCTTestIdentifier, activityRecord nskeyedarchiver.XCActivityRecord)
	TestCaseMethodDidStallOnMainThreadInFileLine()
	TestCaseMethodWillStartActivity()
	TestCaseWithIdentifierWillStartActivity(testIdentifier nskeyedarchiver.XCTTestIdentifier, activityRecord nskeyedarchiver.XCActivityRecord)
	TestCaseDidFailForTestClassMethodWithMessageFileLine()
	TestCaseWithIdentifierDidRecordIssue(testIdentifier nskeyedarchiver.XCTTestIdentifier, issue nskeyedarchiver.XCTIssue)
	TestCaseDidFinishForTestClassMethodWithStatusDuration()
	TestCaseWithIdentifierDidFinishWithStatusDuration(testIdentifier nskeyedarchiver.XCTTestIdentifier, status string, duration float64)
	TestCaseDidStartForTestClassMethod()
	TestCaseDidStartWithIdentifierTestCaseRunConfiguration(testIdentifier nskeyedarchiver.XCTTestIdentifier)
	TestMethodOfClassDidMeasureMetricFileLine()
	TestSuiteDidFinishAtRunCountWithFailuresUnexpectedTestDurationTotalDuration()
	TestSuiteWithIdentifierDidFinishAtRunCountSkipCountFailureCountExpectedFailureCountUncaughtExceptionCountTestDurationTotalDuration(testIdentifier nskeyedarchiver.XCTTestIdentifier, finishAt string, runCount uint64, skipCount uint64, failureCount uint64, expectedFailureCount uint64, uncaughtExceptionCount uint64, testDuration float64, totalDuration float64)
	TestSuiteDidStartAt()
	TestSuiteWithIdentifierDidStartAt(testIdentifier nskeyedarchiver.XCTTestIdentifier, date string)
}

// A concrete implementation of TestListener that writes all test logs to given Writer and ignores all other messages
type TestLogCollector struct {
	Writer *io.Writer
}

func (t TestLogCollector) LogDebugMessage(msg string) {
	(*t.Writer).Write([]byte(msg))
}

func (t TestLogCollector) LogMessage(msg string) {
	(*t.Writer).Write([]byte(msg))
}

func (t TestLogCollector) DidBeginExecutingTestPlan()                                             {}
func (t TestLogCollector) DidFinishExecutingTestPlan()                                            {}
func (t TestLogCollector) DidFailToBootstrapWithError(err nskeyedarchiver.NSError)                {}
func (t TestLogCollector) DidBeginInitializingForUITesting()                                      {}
func (t TestLogCollector) GetProgressForLaunch()                                                  {}
func (t TestLogCollector) InitializationForUITestingDidFailWithError(err nskeyedarchiver.NSError) {}
func (t TestLogCollector) TestCaseMethodDidFinishActivity()                                       {}
func (t TestLogCollector) TestCaseWithIdentifierDidFinishActivity(testIdentifier nskeyedarchiver.XCTTestIdentifier, activityRecord nskeyedarchiver.XCActivityRecord) {
}
func (t TestLogCollector) TestCaseMethodDidStallOnMainThreadInFileLine() {}
func (t TestLogCollector) TestCaseMethodWillStartActivity()              {}
func (t TestLogCollector) TestCaseWithIdentifierWillStartActivity(testIdentifier nskeyedarchiver.XCTTestIdentifier, activityRecord nskeyedarchiver.XCActivityRecord) {
}
func (t TestLogCollector) TestCaseDidFailForTestClassMethodWithMessageFileLine() {}
func (t TestLogCollector) TestCaseWithIdentifierDidRecordIssue(testIdentifier nskeyedarchiver.XCTTestIdentifier, issue nskeyedarchiver.XCTIssue) {
}
func (t TestLogCollector) TestCaseDidFinishForTestClassMethodWithStatusDuration() {}
func (t TestLogCollector) TestCaseWithIdentifierDidFinishWithStatusDuration(testIdentifier nskeyedarchiver.XCTTestIdentifier, status string, duration float64) {
}
func (t TestLogCollector) TestCaseDidStartForTestClassMethod() {}
func (t TestLogCollector) TestCaseDidStartWithIdentifierTestCaseRunConfiguration(testIdentifier nskeyedarchiver.XCTTestIdentifier) {
}
func (t TestLogCollector) TestMethodOfClassDidMeasureMetricFileLine() {}
func (t TestLogCollector) TestSuiteDidFinishAtRunCountWithFailuresUnexpectedTestDurationTotalDuration() {
}
func (t TestLogCollector) TestSuiteWithIdentifierDidFinishAtRunCountSkipCountFailureCountExpectedFailureCountUncaughtExceptionCountTestDurationTotalDuration(testIdentifier nskeyedarchiver.XCTTestIdentifier, finishAt string, runCount uint64, skipCount uint64, failureCount uint64, expectedFailureCount uint64, uncaughtExceptionCount uint64, testDuration float64, totalDuration float64) {
}
func (t TestLogCollector) TestSuiteDidStartAt() {}
func (t TestLogCollector) TestSuiteWithIdentifierDidStartAt(testIdentifier nskeyedarchiver.XCTTestIdentifier, date string) {
}
