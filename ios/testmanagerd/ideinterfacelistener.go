package testmanagerd

import "github.com/danielpaulus/go-ios/ios/nskeyedarchiver"

// Raw ide interface message listener. It forwards converts low level
// testmanagerd selectors to unified callbacks defined under TestListener
type IdeInterfaceListener struct {
	testListener TestListener
}

func (t IdeInterfaceListener) LogMessage(msg string) {
	t.testListener.LogMessage(msg)
}
func (t IdeInterfaceListener) LogDebugMessage(msg string) {
	t.testListener.LogDebugMessage(msg)
}
func (t IdeInterfaceListener) DidBeginExecutingTestPlan() {
	t.testListener.DidBeginExecutingTestPlan()
}
func (t IdeInterfaceListener) DidFinishExecutingTestPlan() {
	t.testListener.DidFinishExecutingTestPlan()
}
func (t IdeInterfaceListener) DidFailToBootstrapWithError(err nskeyedarchiver.NSError) {
	// intentionally ignored
}
func (t IdeInterfaceListener) DidBeginInitializingForUITesting() {
	// intentionally ignored
}
func (t IdeInterfaceListener) GetProgressForLaunch() {
	// intentionally ignored
}
func (t IdeInterfaceListener) InitializationForUITestingDidFailWithError(err nskeyedarchiver.NSError) {
	t.testListener.InitializationForUITestingFailed(err)
}
func (t IdeInterfaceListener) TestCaseMethodDidFinishActivity(testCase string, testMethod string, activityRecord nskeyedarchiver.XCActivityRecord) {
	t.testListener.TestCaseFinished(testCase, testMethod, activityRecord)
}
func (t IdeInterfaceListener) TestCaseWithIdentifierDidFinishActivity(testIdentifier nskeyedarchiver.XCTTestIdentifier, activityRecord nskeyedarchiver.XCActivityRecord) {
	t.testListener.TestCaseFinished(testIdentifier.C[0], testIdentifier.C[1], activityRecord)
}
func (t IdeInterfaceListener) TestCaseMethodDidStallOnMainThreadInFileLine(testCase string, testMethod string, file string, line uint64) {
	t.testListener.TestCaseStalled(testCase, testMethod, file, line)
}
func (t IdeInterfaceListener) TestCaseMethodWillStartActivity() {
	// intentionally ignored
}
func (t IdeInterfaceListener) TestCaseWithIdentifierWillStartActivity(testIdentifier nskeyedarchiver.XCTTestIdentifier, activityRecord nskeyedarchiver.XCActivityRecord) {
	// intentionally ignored
}
func (t IdeInterfaceListener) TestCaseDidFailForTestClassMethodWithMessageFileLine(testCase string, testMethod string, message string, file string, line uint64) {
	t.testListener.TestCaseFailedForClass(testCase, testMethod, message, file, line)
}
func (t IdeInterfaceListener) TestCaseWithIdentifierDidRecordIssue(testIdentifier nskeyedarchiver.XCTTestIdentifier, issue nskeyedarchiver.XCTIssue) {
	t.testListener.TestCaseFailedForClass(testIdentifier.C[0], testIdentifier.C[1], issue.CompactDescription, issue.SourceCodeContext.Location.FileUrl.Path, issue.SourceCodeContext.Location.LineNumber)
}
func (t IdeInterfaceListener) TestCaseDidFinishForTestClassMethodWithStatusDuration(testCase string, testMethod string, status string, duration float64) {
	t.testListener.TestCaseDidFinishForTest(testCase, testMethod, status, duration)
}
func (t IdeInterfaceListener) TestCaseWithIdentifierDidFinishWithStatusDuration(testIdentifier nskeyedarchiver.XCTTestIdentifier, status string, duration float64) {
	t.testListener.TestCaseDidFinishForTest(testIdentifier.C[0], testIdentifier.C[1], status, duration)
}
func (t IdeInterfaceListener) TestCaseDidStartForTestClassMethod(testClass string, testMethod string) {
	t.testListener.TestCaseDidStartForClass(testClass, testMethod)
}
func (t IdeInterfaceListener) TestCaseDidStartWithIdentifierTestCaseRunConfiguration(testIdentifier nskeyedarchiver.XCTTestIdentifier) {
	t.testListener.TestCaseDidStartForClass(testIdentifier.C[0], testIdentifier.C[1])
}
func (t IdeInterfaceListener) TestMethodOfClassDidMeasureMetricFileLine() {
	// intentionally ignored
}
func (t IdeInterfaceListener) TestSuiteDidFinishAtRunCountWithFailuresUnexpectedTestDurationTotalDuration(testSuite string, finishAt string, runCount uint64, failures uint64, unexpectedFailureCount uint64, testDuration float64, totalDuration float64) {
	t.testListener.TestSuiteFinished(testSuite, finishAt, runCount, failures, unexpectedFailureCount, testDuration, totalDuration)
}
func (t IdeInterfaceListener) TestSuiteWithIdentifierDidFinishAtRunCountSkipCountFailureCountExpectedFailureCountUncaughtExceptionCountTestDurationTotalDuration(testIdentifier nskeyedarchiver.XCTTestIdentifier, finishAt string, runCount uint64, skipCount uint64, failureCount uint64, expectedFailureCount uint64, uncaughtExceptionCount uint64, testDuration float64, totalDuration float64) {
	t.testListener.TestSuiteFinished(testIdentifier.C[0], finishAt, runCount, failureCount, expectedFailureCount, testDuration, totalDuration)
}
func (t IdeInterfaceListener) TestSuiteDidStartAt() {
	// intentionally ignored
}
func (t IdeInterfaceListener) TestSuiteWithIdentifierDidStartAt(testIdentifier nskeyedarchiver.XCTTestIdentifier, date string) {
	// intentionally ignored
}
