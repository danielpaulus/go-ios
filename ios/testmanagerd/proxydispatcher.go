package testmanagerd

import (
	"fmt"
	"runtime/debug"
	"strings"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
)

type proxyDispatcher struct {
	testBundleReadyChannel          chan dtx.Message
	testRunnerReadyWithCapabilities dtx.MethodWithResponse
	dtxConnection                   *dtx.Connection
	id                              string
	testListener                    *TestListener
}

func (p proxyDispatcher) Dispatch(m dtx.Message) {
	var dispatcher = &p
	defer func() {
		if r := recover(); r != nil {
			stacktrace := string(debug.Stack())
			dispatcher.testListener.err = fmt.Errorf("Dispatch: %s\n%s", r, stacktrace)
		}
	}()

	var decoderErr error
	shouldAck := true
	if len(m.Payload) == 1 {
		method := m.Payload[0].(string)

		if !strings.Contains(method, "logDebugMessage") {
			golog.Debug("method received", "module", logModule, "id", p.id, "method", method)
		}

		switch method {
		case "_XCT_testBundleReadyWithProtocolVersion:minimumVersion:":
			p.testBundleReadyChannel <- m
			return
		case "_XCT_testRunnerReadyWithCapabilities:":
			shouldAck = false
			golog.Debug("received testRunnerReadyWithCapabilities", "module", logModule, "id", p.id)
			resp, _ := p.testRunnerReadyWithCapabilities(m)
			payload, _ := nskeyedarchiver.ArchiveBin(resp)
			var messageBytes []byte
			messageBytes, decoderErr = dtx.Encode(m.Identifier, 1, m.ChannelCode, false, dtx.ResponseWithReturnValueInPayload, payload, dtx.NewPrimitiveDictionary())
			if decoderErr != nil { // Actually an encoder error but we can utilize the same logic for decoder errors and quit early
				break
			}

			golog.Debug("sending response for capabs", "module", logModule, "id", p.id)
			p.dtxConnection.Send(messageBytes)
		case "_XCT_didBeginExecutingTestPlan":
			golog.Debug("_XCT_didBeginExecutingTestPlan received. Executing test.", "module", logModule, "id", p.id)
		case "_XCT_didFinishExecutingTestPlan":
			golog.Debug("_XCT_didFinishExecutingTestPlan received. Closing test.", "module", logModule, "id", p.id)

			p.testListener.didFinishExecutingTestPlan()
		case "_XCT_initializationForUITestingDidFailWithError:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 1)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var err nskeyedarchiver.NSError
			err, decoderErr = extractNSErrorArg(m, 0)
			if decoderErr != nil {
				break
			}

			p.testListener.initializationForUITestingDidFailWithError(err)
		case "_XCT_didFailToBootstrapWithError:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 1)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var err nskeyedarchiver.NSError
			err, decoderErr = extractNSErrorArg(m, 0)
			if decoderErr != nil {
				break
			}
			p.testListener.didFailToBootstrapWithError(err)
		case "_XCT_logMessage:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 1)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			mbytes := m.Auxiliary.GetArguments()[0].([]byte)
			data, _ := nskeyedarchiver.Unarchive(mbytes)

			p.testListener.LogDebugMessage(data[0].(string))
		case "_XCT_logDebugMessage:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 1)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			mbytes := m.Auxiliary.GetArguments()[0].([]byte)
			var data []interface{}
			data, decoderErr = nskeyedarchiver.Unarchive(mbytes)
			if decoderErr != nil {
				break
			}

			p.testListener.LogMessage(data[0].(string))
		case "_XCT_didBeginInitializingForUITesting":
			golog.Debug("_XCT_didBeginInitializingForUITesting received. ", "module", logModule, "id", p.id)
		case "_XCT_getProgressForLaunch:":
			golog.Debug("_XCT_getProgressForLaunch received. ", "module", logModule, "id", p.id)
		case "_XCT_testCase:method:didFinishActivity:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 3)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testCase, testMethod string
			var activityRecord nskeyedarchiver.XCActivityRecord
			testCase, decoderErr = extractStringArg(m, 0)
			if decoderErr != nil {
				break
			}
			testMethod, decoderErr = extractStringArg(m, 1)
			if decoderErr != nil {
				break
			}
			activityRecord, decoderErr = extractActivityRecordArg(m, 2)
			if decoderErr != nil {
				break
			}

			p.testListener.testCaseFinished(testCase, testMethod, activityRecord)
		case "_XCT_testCaseWithIdentifier:didFinishActivity:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 2)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testIdentifier nskeyedarchiver.XCTTestIdentifier
			var activityRecord nskeyedarchiver.XCActivityRecord
			testIdentifier, decoderErr = extractTestIdentifierArg(m, 0)
			if decoderErr != nil {
				break
			}
			activityRecord, decoderErr = extractActivityRecordArg(m, 1)
			if decoderErr != nil {
				break
			}

			p.testListener.testCaseFinished(testIdentifier.C[0], testIdentifier.C[1], activityRecord)
		case "_XCT_testCase:method:didStallOnMainThreadInFile:line:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 4)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testCase, testMethod, file string
			var line uint64
			testCase, decoderErr = extractStringArg(m, 0)
			if decoderErr != nil {
				break
			}
			testMethod, decoderErr = extractStringArg(m, 1)
			if decoderErr != nil {
				break
			}
			file, decoderErr = extractStringArg(m, 2)
			if decoderErr != nil {
				break
			}
			line, decoderErr = extractUint64Arg(m, 3)
			if decoderErr != nil {
				break
			}

			p.testListener.testCaseStalled(testCase, testMethod, file, line)
		case "_XCT_testCase:method:willStartActivity:":
			golog.Debug("_XCT_testCase:method:willStartActivity: received.", "module", logModule, "id", p.id)
		case "_XCT_testCaseWithIdentifier:willStartActivity:":
			golog.Debug("_XCT_testCaseWithIdentifier:willStartActivity: received.", "module", logModule, "id", p.id)
		case "_XCT_testCaseDidFailForTestClass:method:withMessage:file:line:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 5)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testCase, testMethod, message, file string
			var line uint64
			testCase, decoderErr = extractStringArg(m, 0)
			if decoderErr != nil {
				break
			}
			testMethod, decoderErr = extractStringArg(m, 1)
			if decoderErr != nil {
				break
			}
			message, decoderErr = extractStringArg(m, 2)
			if decoderErr != nil {
				break
			}
			file, decoderErr = extractStringArg(m, 3)
			if decoderErr != nil {
				break
			}
			line, decoderErr = extractUint64Arg(m, 4)
			if decoderErr != nil {
				break
			}

			p.testListener.testCaseFailedForClass(testCase, testMethod, message, file, line)
		case "_XCT_testCaseWithIdentifier:didRecordIssue:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 2)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testIdentifier nskeyedarchiver.XCTTestIdentifier
			var issue nskeyedarchiver.XCTIssue
			testIdentifier, decoderErr = extractTestIdentifierArg(m, 0)
			if decoderErr != nil {
				break
			}
			issue, decoderErr = extractIssueArg(m, 1)
			if decoderErr != nil {
				break
			}

			p.testListener.testCaseFailedForClass(testIdentifier.C[0], testIdentifier.C[1], issue.CompactDescription, issue.SourceCodeContext.Location.FileUrl.Path, issue.SourceCodeContext.Location.LineNumber)
		case "_XCT_testCaseDidFinishForTestClass:method:withStatus:duration:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 4)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testCase, testMethod, status string
			var duration float64
			testCase, decoderErr = extractStringArg(m, 0)
			if decoderErr != nil {
				break
			}
			testMethod, decoderErr = extractStringArg(m, 1)
			if decoderErr != nil {
				break
			}
			status, decoderErr = extractStringArg(m, 2)
			if decoderErr != nil {
				break
			}
			duration, decoderErr = extractFloat64Arg(m, 3)
			if decoderErr != nil {
				break
			}

			p.testListener.testCaseDidFinishForTest(testCase, testMethod, status, duration)
		case "_XCT_testCaseWithIdentifier:didFinishWithStatus:duration:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 3)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testIdentifier nskeyedarchiver.XCTTestIdentifier
			var status string
			var duration float64
			testIdentifier, decoderErr = extractTestIdentifierArg(m, 0)
			if decoderErr != nil {
				break
			}
			status, decoderErr = extractStringArg(m, 1)
			if decoderErr != nil {
				break
			}
			duration, decoderErr = extractFloat64Arg(m, 2)
			if decoderErr != nil {
				break
			}

			p.testListener.testCaseDidFinishForTest(testIdentifier.C[0], testIdentifier.C[1], status, duration)
		case "_XCT_testCaseDidStartForTestClass:method:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 2)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testClass, testMethod string
			testClass, decoderErr = extractStringArg(m, 0)
			if decoderErr != nil {
				break
			}
			testMethod, decoderErr = extractStringArg(m, 1)
			if decoderErr != nil {
				break
			}

			p.testListener.testCaseDidStartForClass(testClass, testMethod)
		case "_XCT_testCaseDidStartWithIdentifier:iteration:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 2)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testIdentifier nskeyedarchiver.XCTTestIdentifier
			testIdentifier, decoderErr = extractTestIdentifierArg(m, 0)
			if decoderErr != nil {
				break
			}

			p.testListener.testCaseDidStartForClass(testIdentifier.C[0], testIdentifier.C[1])
		case "_XCT_testCaseDidStartWithIdentifier:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 1)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testIdentifier nskeyedarchiver.XCTTestIdentifier
			testIdentifier, decoderErr = extractTestIdentifierArg(m, 0)
			if decoderErr != nil {
				break
			}

			p.testListener.testCaseDidStartForClass(testIdentifier.C[0], testIdentifier.C[1])
		case "_XCT_testCaseDidStartWithIdentifier:testCaseRunConfiguration:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 2)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testIdentifier nskeyedarchiver.XCTTestIdentifier
			testIdentifier, decoderErr = extractTestIdentifierArg(m, 0)
			if decoderErr != nil {
				break
			}

			p.testListener.testCaseDidStartForClass(testIdentifier.C[0], testIdentifier.C[1])
		case "_XCT_testMethod:ofClass:didMeasureMetric:file:line:":
			golog.Debug("_XCT_testMethod:ofClass:didMeasureMetric:file:line: received.", "module", logModule, "id", p.id)
		case "_XCT_testSuite:didFinishAt:runCount:withFailures:unexpected:testDuration:totalDuration:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 7)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testSuite, finishAt string
			var runCount, failures, unexpectedFailureCount uint64
			var testDuration, totalDuration float64
			testSuite, decoderErr = extractStringArg(m, 0)
			if decoderErr != nil {
				break
			}
			finishAt, decoderErr = extractStringArg(m, 1)
			if decoderErr != nil {
				break
			}
			runCount, decoderErr = extractUint64Arg(m, 2)
			if decoderErr != nil {
				break
			}
			failures, decoderErr = extractUint64Arg(m, 3)
			if decoderErr != nil {
				break
			}
			unexpectedFailureCount, decoderErr = extractUint64Arg(m, 4)
			if decoderErr != nil {
				break
			}
			testDuration, decoderErr = extractFloat64Arg(m, 5)
			if decoderErr != nil {
				break
			}
			totalDuration, decoderErr = extractFloat64Arg(m, 6)
			if decoderErr != nil {
				break
			}

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
			argumentLengthErr := assertArgumentsLengthEqual(m, 9)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testIdentifier nskeyedarchiver.XCTTestIdentifier
			var finishAt string
			var runCount, skipCount, failureCount, expectedFailureCount, uncaughtExceptionCount uint64
			var testDuration, totalDuration float64
			testIdentifier, decoderErr = extractTestIdentifierArg(m, 0)
			if decoderErr != nil {
				break
			}
			finishAt, decoderErr = extractStringArg(m, 1)
			if decoderErr != nil {
				break
			}
			runCount, decoderErr = extractUint64Arg(m, 2)
			if decoderErr != nil {
				break
			}
			skipCount, decoderErr = extractUint64Arg(m, 3)
			if decoderErr != nil {
				break
			}
			failureCount, decoderErr = extractUint64Arg(m, 4)
			if decoderErr != nil {
				break
			}
			expectedFailureCount, decoderErr = extractUint64Arg(m, 5)
			if decoderErr != nil {
				break
			}
			uncaughtExceptionCount, decoderErr = extractUint64Arg(m, 6)
			if decoderErr != nil {
				break
			}
			testDuration, decoderErr = extractFloat64Arg(m, 7)
			if decoderErr != nil {
				break
			}
			totalDuration, decoderErr = extractFloat64Arg(m, 8)
			if decoderErr != nil {
				break
			}

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
			argumentLengthErr := assertArgumentsLengthEqual(m, 2)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testSuite, date string
			testSuite, decoderErr = extractStringArg(m, 0)
			if decoderErr != nil {
				break
			}
			date, decoderErr = extractStringArg(m, 1)
			if decoderErr != nil {
				break
			}

			p.testListener.testSuiteDidStart(testSuite, date)
		case "_XCT_testSuiteWithIdentifier:didStartAt:":
			argumentLengthErr := assertArgumentsLengthEqual(m, 2)
			if argumentLengthErr != nil {
				decoderErr = argumentLengthErr
				break
			}

			var testIdentifier nskeyedarchiver.XCTTestIdentifier
			testIdentifier, decoderErr = extractTestIdentifierArg(m, 0)
			if decoderErr != nil {
				break
			}

			if len(testIdentifier.C) > 0 && testIdentifier.C[0] != "All tests" {
				var date string
				date, decoderErr = extractStringArg(m, 1)
				if decoderErr != nil {
					break
				}
				p.testListener.testSuiteDidStart(testIdentifier.C[0], date)
			}
		default:
			golog.Info("device called local method", "module", logModule, "id", p.id, "sel", method)
		}
	}

	if decoderErr != nil {
		dispatcher.testListener.err = decoderErr
	}

	if shouldAck {
		dtx.SendAckIfNeeded(p.dtxConnection, m)
	}

	golog.Trace("dispatcher received message", "module", logModule, "id", p.id, "message", m.String())
}

func assertArgumentsLengthEqual(m dtx.Message, expectedLength uint) error {
	if len(m.Auxiliary.GetArguments()) != int(expectedLength) {
		stacktrace := string(debug.Stack())
		return fmt.Errorf("assertArgumentsLengthEqual: %s\n%s", "Unexpected number of DTX message arguments", stacktrace)
	}

	return nil
}

func extractStringArg(m dtx.Message, index int) (string, error) {
	mbytes, ok := m.Auxiliary.GetArguments()[index].([]byte)
	if !ok {
		stacktrace := string(debug.Stack())
		return "", fmt.Errorf("extractStringArg: %s\n%s", "Unrecognized argument", stacktrace)
	}

	data, err := nskeyedarchiver.Unarchive(mbytes)
	if err != nil {
		return "", err
	}

	if len(data) == 0 {
		return "", fmt.Errorf("extractStringArg: Argument is of unknown type")
	}

	return data[0].(string), nil
}

func extractFloat64Arg(m dtx.Message, index int) (float64, error) {
	mbytes, ok := m.Auxiliary.GetArguments()[index].([]byte)
	if !ok {
		stacktrace := string(debug.Stack())
		return 0, fmt.Errorf("extractFloat64Arg: %s\n%s", "Unrecognized argument", stacktrace)
	}

	data, err := nskeyedarchiver.Unarchive(mbytes)
	if err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, fmt.Errorf("extractFloat64Arg: Argument is of unknown type")
	}

	return data[0].(float64), nil
}

func extractUint64Arg(m dtx.Message, index int) (uint64, error) {
	mbytes, ok := m.Auxiliary.GetArguments()[index].([]byte)
	if !ok {
		stacktrace := string(debug.Stack())
		return 0, fmt.Errorf("extractUint64Arg: %s\n%s", "Unrecognized argument", stacktrace)
	}

	data, err := nskeyedarchiver.Unarchive(mbytes)
	if err != nil {
		return 0, err
	}

	if len(data) == 0 {
		return 0, fmt.Errorf("extractUint64Arg: Argument is of unknown type")
	}

	return data[0].(uint64), nil
}

func extractNSErrorArg(m dtx.Message, index int) (nskeyedarchiver.NSError, error) {
	mbytes, ok := m.Auxiliary.GetArguments()[index].([]byte)
	if !ok {
		stacktrace := string(debug.Stack())
		return nskeyedarchiver.NSError{}, fmt.Errorf("extractNSErrorArg: %s\n%s", "Unrecognized argument", stacktrace)
	}

	data, err := nskeyedarchiver.Unarchive(mbytes)
	if err != nil {
		return nskeyedarchiver.NSError{}, err
	}

	if len(data) == 0 {
		return nskeyedarchiver.NSError{}, fmt.Errorf("extractNSErrorArg: Argument is of unknown type")
	}

	return data[0].(nskeyedarchiver.NSError), nil
}
func extractTestIdentifierArg(m dtx.Message, index int) (nskeyedarchiver.XCTTestIdentifier, error) {
	mbytes, ok := m.Auxiliary.GetArguments()[index].([]byte)
	if !ok {
		stacktrace := string(debug.Stack())
		return nskeyedarchiver.XCTTestIdentifier{}, fmt.Errorf("extractTestIdentifierArg: %s\n%s", "Unrecognized argument", stacktrace)
	}

	data, err := nskeyedarchiver.Unarchive(mbytes)
	if err != nil {
		return nskeyedarchiver.XCTTestIdentifier{}, err
	}

	if len(data) == 0 {
		return nskeyedarchiver.XCTTestIdentifier{}, fmt.Errorf("extractTestIdentifierArg: Argument is of unknown type")
	}

	return data[0].(nskeyedarchiver.XCTTestIdentifier), nil
}

func extractIssueArg(m dtx.Message, index int) (nskeyedarchiver.XCTIssue, error) {
	mbytes, ok := m.Auxiliary.GetArguments()[index].([]byte)
	if !ok {
		stacktrace := string(debug.Stack())
		return nskeyedarchiver.XCTIssue{}, fmt.Errorf("extractIssueArg: %s\n%s", "Unrecognized argument", stacktrace)
	}

	data, err := nskeyedarchiver.Unarchive(mbytes)
	if err != nil {
		return nskeyedarchiver.XCTIssue{}, err
	}

	if len(data) == 0 {
		return nskeyedarchiver.XCTIssue{}, fmt.Errorf("extractIssueArg: Argument is of unknown type")
	}

	return data[0].(nskeyedarchiver.XCTIssue), nil
}

func extractActivityRecordArg(m dtx.Message, index int) (nskeyedarchiver.XCActivityRecord, error) {
	mbytes, ok := m.Auxiliary.GetArguments()[index].([]byte)
	if !ok {
		stacktrace := string(debug.Stack())
		return nskeyedarchiver.XCActivityRecord{}, fmt.Errorf("extractActivityRecordArg: %s\n%s", "Unrecognized argument", stacktrace)
	}

	data, err := nskeyedarchiver.Unarchive(mbytes)
	if err != nil {
		return nskeyedarchiver.XCActivityRecord{}, err
	}

	if len(data) == 0 {
		return nskeyedarchiver.XCActivityRecord{}, fmt.Errorf("extractActivityRecordArg: Argument is of unknown type")
	}

	return data[0].(nskeyedarchiver.XCActivityRecord), nil
}
