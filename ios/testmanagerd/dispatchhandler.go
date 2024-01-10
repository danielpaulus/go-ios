package testmanagerd

import (
	"io"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

type DispatchHandler struct {
	dtxConnection                   *dtx.Connection
	testListener                    *TestListener
	testBundleReadyChannel          chan dtx.Message
	testRunnerReadyWithCapabilities dtx.MethodWithResponse
}

func (d DispatchHandler) HandleDispatch(m dtx.Message, closer io.Closer) {
	shouldAck := true
	if len(m.Payload) == 1 {
		method := m.Payload[0].(string)
		switch method {
		case "_XCT_testBundleReadyWithProtocolVersion:minimumVersion:":
			d.testBundleReadyChannel <- m
			return
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
			if d.testListener != nil {
				(*d.testListener).LogDebugMessage(data[0].(string))
			} else {
				log.Debug(data)
			}
		case "_XCT_logDebugMessage:":
			mbytes := m.Auxiliary.GetArguments()[0].([]byte)
			data, _ := nskeyedarchiver.Unarchive(mbytes)
			if d.testListener != nil {
				(*d.testListener).LogMessage(data[0].(string))
			} else {
				log.Debug(data)
			}
		case "_XCT_didBeginExecutingTestPlan":
			if d.testListener != nil {
				(*d.testListener).DidBeginExecutingTestPlan()
			}
		case "_XCT_didFinishExecutingTestPlan":
			if d.testListener != nil {
				(*d.testListener).DidFinishExecutingTestPlan()
			}
			log.Info("_XCT_didFinishExecutingTestPlan received. Closing test.")
			closer.Close()
		case "_XCT_didFailToBootstrapWithError:":
			if d.testListener != nil {
				(*d.testListener).DidFailToBootstrapWithError(extractNSErrorArg(m, 0))
			}
			log.Info("_XCT_didFailToBootstrapWithError received. Closing test.")
			closer.Close()
		case "_XCT_didBeginInitializingForUITesting":
			if d.testListener != nil {
				(*d.testListener).DidBeginInitializingForUITesting()
			}
		case "_XCT_getProgressForLaunch:": // TODO
			if d.testListener != nil {
				(*d.testListener).GetProgressForLaunch()
			}
		case "_XCT_initializationForUITestingDidFailWithError:":
			if d.testListener != nil {
				(*d.testListener).InitializationForUITestingDidFailWithError(extractNSErrorArg(m, 0))
			}
		case "_XCT_testCase:method:didFinishActivity:": // TODO
			if d.testListener != nil {
				(*d.testListener).TestCaseMethodDidFinishActivity()
			}
		case "_XCT_testCaseWithIdentifier:didFinishActivity:":
			if d.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				activityRecord := extractActivityRecordArg(m, 1)
				(*d.testListener).TestCaseWithIdentifierDidFinishActivity(testIdentifier, activityRecord)
			}
		case "_XCT_testCase:method:didStallOnMainThreadInFile:line:": // TODO
			if d.testListener != nil {
				(*d.testListener).TestCaseMethodDidStallOnMainThreadInFileLine()
			}
		case "_XCT_testCase:method:willStartActivity:": // TODO
			if d.testListener != nil {
				(*d.testListener).TestCaseMethodWillStartActivity()
			}
		case "_XCT_testCaseWithIdentifier:willStartActivity:":
			if d.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				activityRecord := extractActivityRecordArg(m, 1)
				(*d.testListener).TestCaseWithIdentifierWillStartActivity(testIdentifier, activityRecord)
			}
		case "_XCT_testCaseDidFailForTestClass:method:withMessage:file:line:": // TODO
			if d.testListener != nil {
				(*d.testListener).TestCaseDidFailForTestClassMethodWithMessageFileLine()
			}
		case "_XCT_testCaseWithIdentifier:didRecordIssue:":
			if d.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				issue := extractIssueArg(m, 1)
				(*d.testListener).TestCaseWithIdentifierDidRecordIssue(testIdentifier, issue)
			}
		case "_XCT_testCaseDidFinishForTestClass:method:withStatus:duration:": // TODO
			if d.testListener != nil {
				(*d.testListener).TestCaseDidFinishForTestClassMethodWithStatusDuration()
			}
		case "_XCT_testCaseWithIdentifier:didFinishWithStatus:duration:": // TODO
			if d.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				status := extractStringArg(m, 1)
				duration := extractFloat64Arg(m, 2)

				(*d.testListener).TestCaseWithIdentifierDidFinishWithStatusDuration(testIdentifier, status, duration)
			}
		case "_XCT_testCaseDidStartForTestClass:method:": // TODO
			if d.testListener != nil {
				(*d.testListener).TestCaseDidStartForTestClassMethod()
			}
		case "_XCT_testCaseDidStartWithIdentifier:testCaseRunConfiguration:": // TODO
			if d.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				(*d.testListener).TestCaseDidStartWithIdentifierTestCaseRunConfiguration(testIdentifier)
			}
		case "_XCT_testMethod:ofClass:didMeasureMetric:file:line:": // TODO
			if d.testListener != nil {
				(*d.testListener).TestMethodOfClassDidMeasureMetricFileLine()
			}
		case "_XCT_testSuite:didFinishAt:runCount:withFailures:unexpected:testDuration:totalDuration:": // TODO
			if d.testListener != nil {
				(*d.testListener).TestSuiteDidFinishAtRunCountWithFailuresUnexpectedTestDurationTotalDuration()
			}
		case "_XCT_testSuiteWithIdentifier:didFinishAt:runCount:skipCount:failureCount:expectedFailureCount:uncaughtExceptionCount:testDuration:totalDuration:": // TODO
			if d.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				finishAt := extractStringArg(m, 1)
				runCount := extractUint64Arg(m, 2)
				skipCount := extractUint64Arg(m, 3)
				failureCount := extractUint64Arg(m, 4)
				expectedFailureCount := extractUint64Arg(m, 5)
				uncaughtExceptionCount := extractUint64Arg(m, 6)
				testDuration := extractFloat64Arg(m, 7)
				totalDuration := extractFloat64Arg(m, 8)

				(*d.testListener).TestSuiteWithIdentifierDidFinishAtRunCountSkipCountFailureCountExpectedFailureCountUncaughtExceptionCountTestDurationTotalDuration(
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
		case "_XCT_testSuite:didStartAt:": // TODO
			if d.testListener != nil {
				(*d.testListener).TestSuiteDidStartAt()
			}
		case "_XCT_testSuiteWithIdentifier:didStartAt:":
			if d.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				date := extractStringArg(m, 1)

				(*d.testListener).TestSuiteWithIdentifierDidStartAt(testIdentifier, date)
			}
		default:
			log.WithFields(log.Fields{"sel": method}).Infof("device called local method")
		}
	}
	if shouldAck {
		dtx.SendAckIfNeeded(d.dtxConnection, m)
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
