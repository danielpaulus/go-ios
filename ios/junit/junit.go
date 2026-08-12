// Package junit converts the test results collected by testmanagerd
// (TestSuite/TestCase structs) into standard JUnit XML, so `ios runtest`
// results can be consumed by CI systems and device farms.
// It is a pure formatter: it only reads the testmanagerd result structs.
package junit

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"time"

	"github.com/danielpaulus/go-ios/ios/testmanagerd"
)

// timestampFormat is the ISO8601 local-time format conventionally used in
// JUnit XML timestamp attributes.
const timestampFormat = "2006-01-02T15:04:05"

type xmlTestSuites struct {
	XMLName  xml.Name       `xml:"testsuites"`
	Tests    int            `xml:"tests,attr"`
	Failures int            `xml:"failures,attr"`
	Errors   int            `xml:"errors,attr"`
	Skipped  int            `xml:"skipped,attr"`
	Time     string         `xml:"time,attr"`
	Suites   []xmlTestSuite `xml:"testsuite"`
}

type xmlTestSuite struct {
	Name      string        `xml:"name,attr"`
	Tests     int           `xml:"tests,attr"`
	Failures  int           `xml:"failures,attr"`
	Errors    int           `xml:"errors,attr"`
	Skipped   int           `xml:"skipped,attr"`
	Time      string        `xml:"time,attr"`
	Timestamp string        `xml:"timestamp,attr,omitempty"`
	Cases     []xmlTestCase `xml:"testcase"`
}

type xmlTestCase struct {
	ClassName string     `xml:"classname,attr"`
	Name      string     `xml:"name,attr"`
	Time      string     `xml:"time,attr"`
	Failure   *xmlResult `xml:"failure,omitempty"`
	Error     *xmlResult `xml:"error,omitempty"`
	Skipped   *xmlResult `xml:"skipped,omitempty"`
}

type xmlResult struct {
	Message string `xml:"message,attr,omitempty"`
	Content string `xml:",chardata"`
}

// Write serializes the given test suites as JUnit XML to w,
// including the XML header.
func Write(w io.Writer, suites []testmanagerd.TestSuite) error {
	report := convert(suites)
	if _, err := io.WriteString(w, xml.Header); err != nil {
		return err
	}
	encoder := xml.NewEncoder(w)
	encoder.Indent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return err
	}
	_, err := io.WriteString(w, "\n")
	return err
}

func convert(suites []testmanagerd.TestSuite) xmlTestSuites {
	report := xmlTestSuites{Suites: make([]xmlTestSuite, 0, len(suites))}
	var totalTime time.Duration
	for _, suite := range suites {
		converted, duration := convertSuite(suite)
		report.Tests += converted.Tests
		report.Failures += converted.Failures
		report.Errors += converted.Errors
		report.Skipped += converted.Skipped
		totalTime += duration
		report.Suites = append(report.Suites, converted)
	}
	report.Time = formatSeconds(totalTime)
	return report
}

func convertSuite(suite testmanagerd.TestSuite) (xmlTestSuite, time.Duration) {
	converted := xmlTestSuite{
		Name:  suite.Name,
		Tests: len(suite.TestCases),
		Cases: make([]xmlTestCase, 0, len(suite.TestCases)),
	}
	if !suite.StartDate.IsZero() {
		converted.Timestamp = suite.StartDate.Format(timestampFormat)
	}
	var caseTime time.Duration
	for _, testCase := range suite.TestCases {
		caseTime += testCase.Duration
		convertedCase := convertCase(testCase)
		switch {
		case convertedCase.Failure != nil:
			converted.Failures++
		case convertedCase.Error != nil:
			converted.Errors++
		case convertedCase.Skipped != nil:
			converted.Skipped++
		}
		converted.Cases = append(converted.Cases, convertedCase)
	}
	// TestDuration is only set when the suite finished regularly, fall back
	// to the sum of the test case durations otherwise.
	duration := suite.TestDuration
	if duration == 0 {
		duration = caseTime
	}
	converted.Time = formatSeconds(duration)
	return converted, duration
}

func convertCase(testCase testmanagerd.TestCase) xmlTestCase {
	converted := xmlTestCase{
		ClassName: testCase.ClassName,
		Name:      testCase.MethodName,
		Time:      formatSeconds(testCase.Duration),
	}
	switch testCase.Status {
	case testmanagerd.StatusPassed:
		// nothing to add, a plain testcase element means success
	case testmanagerd.StatusFailed:
		converted.Failure = convertError(testCase.Err)
	case testmanagerd.StatusStalled:
		converted.Error = convertError(testCase.Err)
	case testmanagerd.StatusExpectedFailure:
		message := "expected failure"
		if testCase.Err.Message != "" {
			message += ": " + testCase.Err.Message
		}
		converted.Skipped = &xmlResult{Message: message}
	case "":
		// the test case started but never reported a final status, e.g.
		// because the app crashed or the connection to the device was lost
		message := testCase.Err.Message
		if message == "" {
			message = "no test result received"
		}
		converted.Error = &xmlResult{Message: message}
	default:
		// covers XCTSkip-style skipped tests
		message := testCase.Err.Message
		if message == "" {
			message = string(testCase.Status)
		}
		converted.Skipped = &xmlResult{Message: message}
	}
	return converted
}

func convertError(testError testmanagerd.TestError) *xmlResult {
	result := &xmlResult{Message: testError.Message}
	if testError.File != "" {
		result.Content = fmt.Sprintf("%s:%d", testError.File, testError.Line)
	}
	return result
}

func formatSeconds(d time.Duration) string {
	return strconv.FormatFloat(d.Seconds(), 'f', 3, 64)
}
