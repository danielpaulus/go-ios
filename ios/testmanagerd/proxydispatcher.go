package testmanagerd

import (
	"fmt"
	"time"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

type ProxyDispatcher struct {
	testBundleReadyChannel          chan dtx.Message
	testRunnerReadyWithCapabilities dtx.MethodWithResponse
	dtxConnection                   *dtx.Connection
	id                              string
	closeChannel                    chan interface{}
	closedChannel                   chan interface{}
	testListener                    *TestListener
}

func (p ProxyDispatcher) Dispatch(m dtx.Message) {
	shouldAck := true
	if len(m.Payload) == 1 {
		method := m.Payload[0].(string)
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
		case "_XCT_logMessage:":
			mbytes := m.Auxiliary.GetArguments()[0].([]byte)
			data, _ := nskeyedarchiver.Unarchive(mbytes)
			if p.testListener != nil {
				(*p.testListener).LogDebugMessage(data[0].(string))
			} else {
				log.Debug(data)
			}
		case "_XCT_logDebugMessage:":
			mbytes := m.Auxiliary.GetArguments()[0].([]byte)
			data, _ := nskeyedarchiver.Unarchive(mbytes)
			if p.testListener != nil {
				(*p.testListener).LogMessage(data[0].(string))
			} else {
				log.Debug(data)
			}
		case "_XCT_didBeginExecutingTestPlan":
			if p.testListener != nil {
				(*p.testListener).DidBeginExecutingTestPlan()
			}
		case "_XCT_didFinishExecutingTestPlan":
			if p.testListener != nil {
				(*p.testListener).DidFinishExecutingTestPlan()
			}
			log.Info("_XCT_didFinishExecutingTestPlan received. Closing test.")
			p.DispatchClose()
		case "_XCT_didFailToBootstrapWithError:":
			if p.testListener != nil {
				(*p.testListener).DidFailToBootstrapWithError(extractNSErrorArg(m, 0))
			}
			log.Info("_XCT_didFailToBootstrapWithError received. Closing test.")
			p.DispatchClose()
		case "_XCT_didBeginInitializingForUITesting":
			if p.testListener != nil {
				(*p.testListener).DidBeginInitializingForUITesting()
			}
		case "_XCT_getProgressForLaunch:": // TODO
			if p.testListener != nil {
				(*p.testListener).GetProgressForLaunch()
			}
		case "_XCT_initializationForUITestingDidFailWithError:":
			if p.testListener != nil {
				(*p.testListener).InitializationForUITestingDidFailWithError(extractNSErrorArg(m, 0))
			}
		case "_XCT_testCase:method:didFinishActivity:": // TODO
			if p.testListener != nil {
				(*p.testListener).TestCaseMethodDidFinishActivity()
			}
		case "_XCT_testCaseWithIdentifier:didFinishActivity:":
			if p.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				activityRecord := extractActivityRecordArg(m, 1)
				(*p.testListener).TestCaseWithIdentifierDidFinishActivity(testIdentifier, activityRecord)
			}
		case "_XCT_testCase:method:didStallOnMainThreadInFile:line:": // TODO
			if p.testListener != nil {
				(*p.testListener).TestCaseMethodDidStallOnMainThreadInFileLine()
			}
		case "_XCT_testCase:method:willStartActivity:": // TODO
			if p.testListener != nil {
				(*p.testListener).TestCaseMethodWillStartActivity()
			}
		case "_XCT_testCaseWithIdentifier:willStartActivity:":
			if p.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				activityRecord := extractActivityRecordArg(m, 1)
				(*p.testListener).TestCaseWithIdentifierWillStartActivity(testIdentifier, activityRecord)
			}
		case "_XCT_testCaseDidFailForTestClass:method:withMessage:file:line:": // TODO
			if p.testListener != nil {
				(*p.testListener).TestCaseDidFailForTestClassMethodWithMessageFileLine()
			}
		case "_XCT_testCaseWithIdentifier:didRecordIssue:":
			if p.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				issue := extractIssueArg(m, 1)
				(*p.testListener).TestCaseWithIdentifierDidRecordIssue(testIdentifier, issue)
			}
		case "_XCT_testCaseDidFinishForTestClass:method:withStatus:duration:": // TODO
			if p.testListener != nil {
				(*p.testListener).TestCaseDidFinishForTestClassMethodWithStatusDuration()
			}
		case "_XCT_testCaseWithIdentifier:didFinishWithStatus:duration:": // TODO
			if p.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				status := extractStringArg(m, 1)
				duration := extractFloat64Arg(m, 2)

				(*p.testListener).TestCaseWithIdentifierDidFinishWithStatusDuration(testIdentifier, status, duration)
			}
		case "_XCT_testCaseDidStartForTestClass:method:": // TODO
			if p.testListener != nil {
				(*p.testListener).TestCaseDidStartForTestClassMethod()
			}
		case "_XCT_testCaseDidStartWithIdentifier:testCaseRunConfiguration:": // TODO
			if p.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				(*p.testListener).TestCaseDidStartWithIdentifierTestCaseRunConfiguration(testIdentifier)
			}
		case "_XCT_testMethod:ofClass:didMeasureMetric:file:line:": // TODO
			if p.testListener != nil {
				(*p.testListener).TestMethodOfClassDidMeasureMetricFileLine()
			}
		case "_XCT_testSuite:didFinishAt:runCount:withFailures:unexpected:testDuration:totalDuration:": // TODO
			if p.testListener != nil {
				(*p.testListener).TestSuiteDidFinishAtRunCountWithFailuresUnexpectedTestDurationTotalDuration()
			}
		case "_XCT_testSuiteWithIdentifier:didFinishAt:runCount:skipCount:failureCount:expectedFailureCount:uncaughtExceptionCount:testDuration:totalDuration:": // TODO
			if p.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				finishAt := extractStringArg(m, 1)
				runCount := extractUint64Arg(m, 2)
				skipCount := extractUint64Arg(m, 3)
				failureCount := extractUint64Arg(m, 4)
				expectedFailureCount := extractUint64Arg(m, 5)
				uncaughtExceptionCount := extractUint64Arg(m, 6)
				testDuration := extractFloat64Arg(m, 7)
				totalDuration := extractFloat64Arg(m, 8)

				(*p.testListener).TestSuiteWithIdentifierDidFinishAtRunCountSkipCountFailureCountExpectedFailureCountUncaughtExceptionCountTestDurationTotalDuration(
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
			if p.testListener != nil {
				(*p.testListener).TestSuiteDidStartAt()
			}
		case "_XCT_testSuiteWithIdentifier:didStartAt:":
			if p.testListener != nil {
				testIdentifier := extractTestIdentifierArg(m, 0)
				date := extractStringArg(m, 1)

				(*p.testListener).TestSuiteWithIdentifierDidStartAt(testIdentifier, date)
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

func (p *ProxyDispatcher) DispatchClose() error {
	var signal interface{}
	go func() { p.closeChannel <- signal }()
	select {
	case <-p.closedChannel:
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("Failed closing, exiting due to timeout")
	}
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
