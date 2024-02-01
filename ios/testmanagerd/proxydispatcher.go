package testmanagerd

import (
	"strings"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

type proxyDispatcher struct {
	testBundleReadyChannel          chan dtx.Message
	testRunnerReadyWithCapabilities dtx.MethodWithResponse
	dtxConnection                   *dtx.Connection
	id                              string
	testListener                    *TestListener
}

func (p proxyDispatcher) Dispatch(m dtx.Message) {
	shouldAck := true
	if len(m.Payload) == 1 {
		method := m.Payload[0].(string)

		if !strings.Contains(method, "logDebugMessage") {
			log.Debug("Method: " + method)
		}

		switch method {
		case "_XCT_testBundleReadyWithProtocolVersion:minimumVersion:":
			p.testBundleReadyChannel <- m
			return
		case "_XCT_testRunnerReadyWithCapabilities:":
			shouldAck = false
			log.Debug("received testRunnerReadyWithCapabilities")
			resp, _ := p.testRunnerReadyWithCapabilities(m)
			payload, _ := nskeyedarchiver.ArchiveBin(resp)
			messageBytes, _ := dtx.Encode(m.Identifier, 1, m.ChannelCode, false, dtx.ResponseWithReturnValueInPayload, payload, dtx.NewPrimitiveDictionary())
			log.Debug("sending response for capabs")
			p.dtxConnection.Send(messageBytes)
		case "_XCT_didBeginExecutingTestPlan":
			log.Debug("_XCT_didBeginExecutingTestPlan received. Executing test.")
		case "_XCT_didFinishExecutingTestPlan":
			log.Debug("_XCT_didFinishExecutingTestPlan received. Closing test.")

			p.testListener.didFinishExecutingTestPlan()
		case "_XCT_initializationForUITestingDidFailWithError:":
			err := extractNSErrorArg(m, 0)

			p.testListener.initializationForUITestingDidFailWithError(err)
		case "_XCT_didFailToBootstrapWithError:":
			err := extractNSErrorArg(m, 0)
			p.testListener.didFailToBootstrapWithError(err)
		case "_XCT_logMessage:":
			mbytes := m.Auxiliary.GetArguments()[0].([]byte)
			data, _ := nskeyedarchiver.Unarchive(mbytes)

			p.testListener.LogDebugMessage(data[0].(string))
		case "_XCT_logDebugMessage:":
			mbytes := m.Auxiliary.GetArguments()[0].([]byte)
			data, _ := nskeyedarchiver.Unarchive(mbytes)

			p.testListener.LogMessage(data[0].(string))
		case "_XCT_didBeginInitializingForUITesting":
			log.Debug("_XCT_didBeginInitializingForUITesting received. ")
		case "_XCT_getProgressForLaunch:":
			log.Debug("_XCT_getProgressForLaunch received. ")
		case "_XCT_testCase:method:didFinishActivity:":
			testCase := extractStringArg(m, 0)
			testMethod := extractStringArg(m, 1)
			activityRecord := extractActivityRecordArg(m, 2)

			p.testListener.testCaseFinished(testCase, testMethod, activityRecord)
		case "_XCT_testCaseWithIdentifier:didFinishActivity:":
			testIdentifier := extractTestIdentifierArg(m, 0)
			activityRecord := extractActivityRecordArg(m, 1)

			p.testListener.testCaseFinished(testIdentifier.C[0], testIdentifier.C[1], activityRecord)
		case "_XCT_testCase:method:didStallOnMainThreadInFile:line:":
			testCase := extractStringArg(m, 0)
			testMethod := extractStringArg(m, 1)
			file := extractStringArg(m, 2)
			line := extractUint64Arg(m, 3)

			p.testListener.testCaseStalled(testCase, testMethod, file, line)
		case "_XCT_testCase:method:willStartActivity:":
			log.Debug("_XCT_testCase:method:willStartActivity: received.")
		case "_XCT_testCaseWithIdentifier:willStartActivity:":
			log.Debug("_XCT_testCaseWithIdentifier:willStartActivity: received.")
		case "_XCT_testCaseDidFailForTestClass:method:withMessage:file:line:":
			testCase := extractStringArg(m, 0)
			testMethod := extractStringArg(m, 1)
			message := extractStringArg(m, 2)
			file := extractStringArg(m, 3)
			line := extractUint64Arg(m, 4)

			p.testListener.testCaseFailedForClass(testCase, testMethod, message, file, line)
		case "_XCT_testCaseWithIdentifier:didRecordIssue:":
			testIdentifier := extractTestIdentifierArg(m, 0)
			issue := extractIssueArg(m, 1)
			p.testListener.testCaseFailedForClass(testIdentifier.C[0], testIdentifier.C[1], issue.CompactDescription, issue.SourceCodeContext.Location.FileUrl.Path, issue.SourceCodeContext.Location.LineNumber)
		case "_XCT_testCaseDidFinishForTestClass:method:withStatus:duration:":
			testCase := extractStringArg(m, 0)
			testMethod := extractStringArg(m, 1)
			status := extractStringArg(m, 2)
			duration := extractFloat64Arg(m, 3)

			p.testListener.testCaseDidFinishForTest(testCase, testMethod, status, duration)
		case "_XCT_testCaseWithIdentifier:didFinishWithStatus:duration:":
			testIdentifier := extractTestIdentifierArg(m, 0)
			status := extractStringArg(m, 1)
			duration := extractFloat64Arg(m, 2)

			p.testListener.testCaseDidFinishForTest(testIdentifier.C[0], testIdentifier.C[1], status, duration)
		case "_XCT_testCaseDidStartForTestClass:method:":
			testClass := extractStringArg(m, 0)
			testMethod := extractStringArg(m, 1)

			p.testListener.testCaseDidStartForClass(testClass, testMethod)
		case "_XCT_testCaseDidStartWithIdentifier:testCaseRunConfiguration:":
			testIdentifier := extractTestIdentifierArg(m, 0)

			p.testListener.testCaseDidStartForClass(testIdentifier.C[0], testIdentifier.C[1])
		case "_XCT_testMethod:ofClass:didMeasureMetric:file:line:":
			log.Debug("_XCT_testMethod:ofClass:didMeasureMetric:file:line: received.")
		case "_XCT_testSuite:didFinishAt:runCount:withFailures:unexpected:testDuration:totalDuration:":
			testSuite := extractStringArg(m, 0)
			finishAt := extractStringArg(m, 1)
			runCount := extractUint64Arg(m, 2)
			failures := extractUint64Arg(m, 3)
			unexpectedFailureCount := extractUint64Arg(m, 4)
			testDuration := extractFloat64Arg(m, 5)
			totalDuration := extractFloat64Arg(m, 6)

			p.testListener.testSuiteFinished(
				testSuite,
				finishAt,
				runCount,
				failures,
				unknownCount, // skip count
				unknownCount, // expected failure count
				unexpectedFailureCount,
				unknownCount, // uncaught exception count
				testDuration,
				totalDuration,
			)
		case "_XCT_testSuiteWithIdentifier:didFinishAt:runCount:skipCount:failureCount:expectedFailureCount:uncaughtExceptionCount:testDuration:totalDuration:":
			testIdentifier := extractTestIdentifierArg(m, 0)
			finishAt := extractStringArg(m, 1)
			runCount := extractUint64Arg(m, 2)
			skipCount := extractUint64Arg(m, 3)
			failureCount := extractUint64Arg(m, 4)
			expectedFailureCount := extractUint64Arg(m, 5)
			uncaughtExceptionCount := extractUint64Arg(m, 6)
			testDuration := extractFloat64Arg(m, 7)
			totalDuration := extractFloat64Arg(m, 8)

			if len(testIdentifier.C) > 0 {
				p.testListener.testSuiteFinished(
					testIdentifier.C[0],
					finishAt,
					runCount,
					failureCount,
					skipCount,
					expectedFailureCount,
					unknownCount, // unexpected failure count
					uncaughtExceptionCount,
					testDuration,
					totalDuration,
				)
			}
		case "_XCT_testSuite:didStartAt:":
			testSuite := extractStringArg(m, 0)
			date := extractStringArg(m, 1)

			p.testListener.testSuiteDidStart(testSuite, date)
		case "_XCT_testSuiteWithIdentifier:didStartAt:":
			testIdentifier := extractTestIdentifierArg(m, 0)

			if len(testIdentifier.C) > 0 && testIdentifier.C[0] != "All tests" {
				date := extractStringArg(m, 1)
				p.testListener.testSuiteDidStart(testIdentifier.C[0], date)
			}
		default:
			log.WithFields(log.Fields{"sel": method}).Infof("device called local method")
		}
	}
	if shouldAck {
		dtx.SendAckIfNeeded(p.dtxConnection, m)
	}
	log.Tracef("dispatcher received: %s", m.String())
}

func extractStringArg(m dtx.Message, index int) string {
	mbytes := m.Auxiliary.GetArguments()[index].([]byte)
	data, _ := nskeyedarchiver.Unarchive(mbytes)
	return data[0].(string)
}

func extractFloat64Arg(m dtx.Message, index int) float64 {
	mbytes := m.Auxiliary.GetArguments()[index].([]byte)
	data, _ := nskeyedarchiver.Unarchive(mbytes)
	return data[0].(float64)
}

func extractUint64Arg(m dtx.Message, index int) uint64 {
	mbytes := m.Auxiliary.GetArguments()[index].([]byte)
	data, _ := nskeyedarchiver.Unarchive(mbytes)
	return data[0].(uint64)
}

func extractNSErrorArg(m dtx.Message, index int) nskeyedarchiver.NSError {
	mbytes := m.Auxiliary.GetArguments()[index].([]byte)
	data, _ := nskeyedarchiver.Unarchive(mbytes)
	return data[0].(nskeyedarchiver.NSError)
}
func extractTestIdentifierArg(m dtx.Message, index int) nskeyedarchiver.XCTTestIdentifier {
	mbytes := m.Auxiliary.GetArguments()[index].([]byte)
	data, _ := nskeyedarchiver.Unarchive(mbytes)
	return data[0].(nskeyedarchiver.XCTTestIdentifier)
}

func extractIssueArg(m dtx.Message, index int) nskeyedarchiver.XCTIssue {
	mbytes := m.Auxiliary.GetArguments()[index].([]byte)
	data, _ := nskeyedarchiver.Unarchive(mbytes)
	return data[0].(nskeyedarchiver.XCTIssue)
}

func extractActivityRecordArg(m dtx.Message, index int) nskeyedarchiver.XCActivityRecord {
	mbytes := m.Auxiliary.GetArguments()[index].([]byte)
	data, _ := nskeyedarchiver.Unarchive(mbytes)
	return data[0].(nskeyedarchiver.XCActivityRecord)
}
