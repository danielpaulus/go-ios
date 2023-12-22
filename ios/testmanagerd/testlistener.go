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
	TestCaseWithIdentifierDidRecordIssue()
	TestCaseDidFinishForTestClassMethodWithStatusDuration()
	TestCaseWithIdentifierDidFinishWithStatusDuration()
	TestCaseDidStartForTestClassMethod()
	TestCaseDidStartWithIdentifierTestCaseRunConfiguration()
	TestMethodOfClassDidMeasureMetricFileLine()
	TestSuiteDidFinishAtRunCountWithFailuresUnexpectedTestDurationTotalDuration()
	TestSuiteWithIdentifierDidFinishAtRunCountSkipCountFailureCountExpectedFailureCountUncaughtExceptionCountTestDurationTotalDuration()
	TestSuiteDidStartAt()
	TestSuiteWithIdentifierDidStartAt(testIdentifier nskeyedarchiver.XCTTestIdentifier, date string)
}

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
func (t TestLogCollector) TestCaseDidFailForTestClassMethodWithMessageFileLine()   {}
func (t TestLogCollector) TestCaseWithIdentifierDidRecordIssue()                   {}
func (t TestLogCollector) TestCaseDidFinishForTestClassMethodWithStatusDuration()  {}
func (t TestLogCollector) TestCaseWithIdentifierDidFinishWithStatusDuration()      {}
func (t TestLogCollector) TestCaseDidStartForTestClassMethod()                     {}
func (t TestLogCollector) TestCaseDidStartWithIdentifierTestCaseRunConfiguration() {}
func (t TestLogCollector) TestMethodOfClassDidMeasureMetricFileLine()              {}
func (t TestLogCollector) TestSuiteDidFinishAtRunCountWithFailuresUnexpectedTestDurationTotalDuration() {
}
func (t TestLogCollector) TestSuiteWithIdentifierDidFinishAtRunCountSkipCountFailureCountExpectedFailureCountUncaughtExceptionCountTestDurationTotalDuration() {
}
func (t TestLogCollector) TestSuiteDidStartAt() {}
func (t TestLogCollector) TestSuiteWithIdentifierDidStartAt(testIdentifier nskeyedarchiver.XCTTestIdentifier, date string) {
}
