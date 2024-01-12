package testmanagerd

import (
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

// Handles dispatching method calls received from the dtx connection to ide interface listeners
type ideInterfaceDtxMessageHandler struct {
	dtxConnection                   *dtx.Connection
	ideInterfaceListener            *IdeInterfaceListener
	testBundleReadyChannel          chan dtx.Message
	testRunnerReadyWithCapabilities dtx.MethodWithResponse
}

func (d ideInterfaceDtxMessageHandler) handleDtxMessage(m dtx.Message) bool {
	shouldClose := false
	shouldAck := true
	if len(m.Payload) == 1 {
		method := m.Payload[0].(string)
		switch method {
		case "_XCT_testBundleReadyWithProtocolVersion:minimumVersion:":
			d.testBundleReadyChannel <- m
			return shouldClose
		case "_XCT_testRunnerReadyWithCapabilities:":
			shouldAck = false
			log.Debug("received testRunnerReadyWithCapabilities")
			resp, _ := d.testRunnerReadyWithCapabilities(m)
			payload, _ := nskeyedarchiver.ArchiveBin(resp)
			messageBytes, _ := dtx.Encode(m.Identifier, 1, m.ChannelCode, false, dtx.ResponseWithReturnValueInPayload, payload, dtx.NewPrimitiveDictionary())
			log.Debug("sending response for capabs")
			d.dtxConnection.Send(messageBytes)
		case "_XCT_logMessage:":
			mbytes := m.Auxiliary.GetArguments()[0].([]byte)
			data, _ := nskeyedarchiver.Unarchive(mbytes)
			if d.ideInterfaceListener != nil {
				(*d.ideInterfaceListener).LogDebugMessage(data[0].(string))
			} else {
				log.Debug(data)
			}
		case "_XCT_logDebugMessage:":
			mbytes := m.Auxiliary.GetArguments()[0].([]byte)
			data, _ := nskeyedarchiver.Unarchive(mbytes)
			if d.ideInterfaceListener != nil {
				(*d.ideInterfaceListener).LogMessage(data[0].(string))
			} else {
				log.Debug(data)
			}
		case "_XCT_didBeginExecutingTestPlan":
			if d.ideInterfaceListener != nil {
				(*d.ideInterfaceListener).DidBeginExecutingTestPlan()
			}
		case "_XCT_didFinishExecutingTestPlan":
			if d.ideInterfaceListener != nil {
				(*d.ideInterfaceListener).DidFinishExecutingTestPlan()
			}
			log.Info("_XCT_didFinishExecutingTestPlan received. Closing test.")
			shouldClose = true
		case "_XCT_didFailToBootstrapWithError:":
			if d.ideInterfaceListener != nil {
				(*d.ideInterfaceListener).DidFailToBootstrapWithError(extractNSErrorArg(m, 0))
			}
			log.Info("_XCT_didFailToBootstrapWithError received. Closing test.")
			shouldClose = true
		case "_XCT_didBeginInitializingForUITesting":
			if d.ideInterfaceListener != nil {
				(*d.ideInterfaceListener).DidBeginInitializingForUITesting()
			}
		case "_XCT_getProgressForLaunch:":
			if d.ideInterfaceListener != nil {
				(*d.ideInterfaceListener).GetProgressForLaunch()
			}
		case "_XCT_initializationForUITestingDidFailWithError:":
			if d.ideInterfaceListener != nil {
				(*d.ideInterfaceListener).InitializationForUITestingDidFailWithError(extractNSErrorArg(m, 0))
			}
		case "_XCT_testCase:method:didFinishActivity:":
			if d.ideInterfaceListener != nil {
				testCase := extractStringArg(m, 0)
				testMethod := extractStringArg(m, 1)
				activityRecord := extractActivityRecordArg(m, 2)
				(*d.ideInterfaceListener).TestCaseMethodDidFinishActivity(testCase, testMethod, activityRecord)
			}
		case "_XCT_testCaseWithIdentifier:didFinishActivity:":
			if d.ideInterfaceListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				activityRecord := extractActivityRecordArg(m, 1)
				(*d.ideInterfaceListener).TestCaseWithIdentifierDidFinishActivity(testIdentifier, activityRecord)
			}
		case "_XCT_testCase:method:didStallOnMainThreadInFile:line:":
			if d.ideInterfaceListener != nil {
				testCase := extractStringArg(m, 0)
				testMethod := extractStringArg(m, 1)
				file := extractStringArg(m, 2)
				line := extractUint64Arg(m, 3)
				(*d.ideInterfaceListener).TestCaseMethodDidStallOnMainThreadInFileLine(testCase, testMethod, file, line)
			}
		case "_XCT_testCase:method:willStartActivity:":
			if d.ideInterfaceListener != nil {
				(*d.ideInterfaceListener).TestCaseMethodWillStartActivity()
			}
		case "_XCT_testCaseWithIdentifier:willStartActivity:":
			if d.ideInterfaceListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				activityRecord := extractActivityRecordArg(m, 1)
				(*d.ideInterfaceListener).TestCaseWithIdentifierWillStartActivity(testIdentifier, activityRecord)
			}
		case "_XCT_testCaseDidFailForTestClass:method:withMessage:file:line:":
			if d.ideInterfaceListener != nil {
				testCase := extractStringArg(m, 0)
				testMethod := extractStringArg(m, 1)
				message := extractStringArg(m, 2)
				file := extractStringArg(m, 3)
				line := extractUint64Arg(m, 4)
				(*d.ideInterfaceListener).TestCaseDidFailForTestClassMethodWithMessageFileLine(testCase, testMethod, message, file, line)
			}
		case "_XCT_testCaseWithIdentifier:didRecordIssue:":
			if d.ideInterfaceListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				issue := extractIssueArg(m, 1)
				(*d.ideInterfaceListener).TestCaseWithIdentifierDidRecordIssue(testIdentifier, issue)
			}
		case "_XCT_testCaseDidFinishForTestClass:method:withStatus:duration:":
			if d.ideInterfaceListener != nil {
				testCase := extractStringArg(m, 0)
				testMethod := extractStringArg(m, 1)
				status := extractStringArg(m, 2)
				duration := extractFloat64Arg(m, 3)

				(*d.ideInterfaceListener).TestCaseDidFinishForTestClassMethodWithStatusDuration(testCase, testMethod, status, duration)
			}
		case "_XCT_testCaseWithIdentifier:didFinishWithStatus:duration:":
			if d.ideInterfaceListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				status := extractStringArg(m, 1)
				duration := extractFloat64Arg(m, 2)

				(*d.ideInterfaceListener).TestCaseWithIdentifierDidFinishWithStatusDuration(testIdentifier, status, duration)
			}
		case "_XCT_testCaseDidStartForTestClass:method:":
			if d.ideInterfaceListener != nil {
				testClass := extractStringArg(m, 0)
				testMethod := extractStringArg(m, 1)
				(*d.ideInterfaceListener).TestCaseDidStartForTestClassMethod(testClass, testMethod)
			}
		case "_XCT_testCaseDidStartWithIdentifier:testCaseRunConfiguration:":
			if d.ideInterfaceListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				(*d.ideInterfaceListener).TestCaseDidStartWithIdentifierTestCaseRunConfiguration(testIdentifier)
			}
		case "_XCT_testMethod:ofClass:didMeasureMetric:file:line:":
			if d.ideInterfaceListener != nil {
				(*d.ideInterfaceListener).TestMethodOfClassDidMeasureMetricFileLine()
			}
		case "_XCT_testSuite:didFinishAt:runCount:withFailures:unexpected:testDuration:totalDuration:":
			if d.ideInterfaceListener != nil {
				testSuite := extractStringArg(m, 0)
				finishAt := extractStringArg(m, 1)
				runCount := extractUint64Arg(m, 2)
				failures := extractUint64Arg(m, 3)
				unexpectedFailureCount := extractUint64Arg(m, 4)
				testDuration := extractFloat64Arg(m, 5)
				totalDuration := extractFloat64Arg(m, 6)

				(*d.ideInterfaceListener).TestSuiteDidFinishAtRunCountWithFailuresUnexpectedTestDurationTotalDuration(testSuite, finishAt, runCount, failures, unexpectedFailureCount, testDuration, totalDuration)
			}
		case "_XCT_testSuiteWithIdentifier:didFinishAt:runCount:skipCount:failureCount:expectedFailureCount:uncaughtExceptionCount:testDuration:totalDuration:":
			if d.ideInterfaceListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				finishAt := extractStringArg(m, 1)
				runCount := extractUint64Arg(m, 2)
				skipCount := extractUint64Arg(m, 3)
				failureCount := extractUint64Arg(m, 4)
				expectedFailureCount := extractUint64Arg(m, 5)
				uncaughtExceptionCount := extractUint64Arg(m, 6)
				testDuration := extractFloat64Arg(m, 7)
				totalDuration := extractFloat64Arg(m, 8)

				(*d.ideInterfaceListener).TestSuiteWithIdentifierDidFinishAtRunCountSkipCountFailureCountExpectedFailureCountUncaughtExceptionCountTestDurationTotalDuration(
					testIdentifier,
					finishAt,
					runCount,
					skipCount,
					failureCount,
					expectedFailureCount,
					uncaughtExceptionCount,
					testDuration,
					totalDuration,
				)
			}
		case "_XCT_testSuite:didStartAt:":
			if d.ideInterfaceListener != nil {
				(*d.ideInterfaceListener).TestSuiteDidStartAt()
			}
		case "_XCT_testSuiteWithIdentifier:didStartAt:":
			if d.ideInterfaceListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				date := extractStringArg(m, 1)

				(*d.ideInterfaceListener).TestSuiteWithIdentifierDidStartAt(testIdentifier, date)
			}
		default:
			log.WithFields(log.Fields{"sel": method}).Infof("device called local method")
		}
	}

	if shouldAck {
		dtx.SendAckIfNeeded(d.dtxConnection, m)
	}
	log.Tracef("dispatcher received: %s", m.String())

	return shouldClose
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
