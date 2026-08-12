package junit_test

import (
	"bytes"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios/junit"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func render(t *testing.T, suites []testmanagerd.TestSuite) string {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, junit.Write(&buf, suites))
	return buf.String()
}

func TestEmptyReport(t *testing.T) {
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="0" failures="0" errors="0" skipped="0" time="0.000"></testsuites>
`
	assert.Equal(t, expected, render(t, nil))
}

func TestPassFailSkipAndDurations(t *testing.T) {
	suites := []testmanagerd.TestSuite{
		{
			Name:         "ExampleUITests",
			StartDate:    time.Date(2026, 8, 5, 10, 30, 0, 0, time.UTC),
			TestDuration: 3500 * time.Millisecond,
			TestCases: []testmanagerd.TestCase{
				{
					ClassName:  "LoginTests",
					MethodName: "testLoginSucceeds",
					Status:     testmanagerd.StatusPassed,
					Duration:   1500 * time.Millisecond,
				},
				{
					ClassName:  "LoginTests",
					MethodName: "testLoginFails",
					Status:     testmanagerd.StatusFailed,
					Duration:   1250 * time.Millisecond,
					Err: testmanagerd.TestError{
						Message: "XCTAssertTrue failed",
						File:    "LoginTests.swift",
						Line:    42,
					},
				},
				{
					ClassName:  "LoginTests",
					MethodName: "testSkipped",
					Status:     testmanagerd.TestCaseStatus("skipped"),
					Duration:   250 * time.Millisecond,
				},
			},
		},
	}
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="3" failures="1" errors="0" skipped="1" time="3.500">
  <testsuite name="ExampleUITests" tests="3" failures="1" errors="0" skipped="1" time="3.500" timestamp="2026-08-05T10:30:00">
    <testcase classname="LoginTests" name="testLoginSucceeds" time="1.500"></testcase>
    <testcase classname="LoginTests" name="testLoginFails" time="1.250">
      <failure message="XCTAssertTrue failed">LoginTests.swift:42</failure>
    </testcase>
    <testcase classname="LoginTests" name="testSkipped" time="0.250">
      <skipped message="skipped"></skipped>
    </testcase>
  </testsuite>
</testsuites>
`
	assert.Equal(t, expected, render(t, suites))
}

func TestStalledExpectedFailureAndMissingStatus(t *testing.T) {
	suites := []testmanagerd.TestSuite{
		{
			Name: "FlakySuite",
			TestCases: []testmanagerd.TestCase{
				{
					ClassName:  "FlakyTests",
					MethodName: "testStalls",
					Status:     testmanagerd.StatusStalled,
					Duration:   2 * time.Second,
					Err: testmanagerd.TestError{
						Message: "Test case stalled",
						File:    "FlakyTests.swift",
						Line:    7,
					},
				},
				{
					ClassName:  "FlakyTests",
					MethodName: "testExpectedFailure",
					Status:     testmanagerd.StatusExpectedFailure,
					Duration:   1 * time.Second,
				},
				{
					ClassName:  "FlakyTests",
					MethodName: "testNeverFinished",
				},
			},
		},
	}
	// no suite TestDuration set: falls back to the sum of the case durations
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="3" failures="0" errors="2" skipped="1" time="3.000">
  <testsuite name="FlakySuite" tests="3" failures="0" errors="2" skipped="1" time="3.000">
    <testcase classname="FlakyTests" name="testStalls" time="2.000">
      <error message="Test case stalled">FlakyTests.swift:7</error>
    </testcase>
    <testcase classname="FlakyTests" name="testExpectedFailure" time="1.000">
      <skipped message="expected failure"></skipped>
    </testcase>
    <testcase classname="FlakyTests" name="testNeverFinished" time="0.000">
      <error message="no test result received"></error>
    </testcase>
  </testsuite>
</testsuites>
`
	assert.Equal(t, expected, render(t, suites))
}

func TestXMLEscaping(t *testing.T) {
	suites := []testmanagerd.TestSuite{
		{
			Name: `Suite <&> "quoted"`,
			TestCases: []testmanagerd.TestCase{
				{
					ClassName:  "EscapeTests",
					MethodName: `test<Generic> & "quotes"`,
					Status:     testmanagerd.StatusFailed,
					Err: testmanagerd.TestError{
						Message: `expected "<nil>" but got <Error & Panic>`,
						File:    "Escape&Tests.swift",
						Line:    1,
					},
				},
			},
		},
	}
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="1" failures="1" errors="0" skipped="0" time="0.000">
  <testsuite name="Suite &lt;&amp;&gt; &#34;quoted&#34;" tests="1" failures="1" errors="0" skipped="0" time="0.000">
    <testcase classname="EscapeTests" name="test&lt;Generic&gt; &amp; &#34;quotes&#34;" time="0.000">
      <failure message="expected &#34;&lt;nil&gt;&#34; but got &lt;Error &amp; Panic&gt;">Escape&amp;Tests.swift:1</failure>
    </testcase>
  </testsuite>
</testsuites>
`
	assert.Equal(t, expected, render(t, suites))
}

func TestMultipleSuitesAggregateCounts(t *testing.T) {
	suites := []testmanagerd.TestSuite{
		{
			Name:         "SuiteA",
			TestDuration: 1 * time.Second,
			TestCases: []testmanagerd.TestCase{
				{ClassName: "A", MethodName: "testOne", Status: testmanagerd.StatusPassed, Duration: time.Second},
			},
		},
		{
			Name:         "SuiteB",
			TestDuration: 2 * time.Second,
			TestCases: []testmanagerd.TestCase{
				{ClassName: "B", MethodName: "testTwo", Status: testmanagerd.StatusFailed, Duration: 2 * time.Second},
			},
		},
	}
	expected := `<?xml version="1.0" encoding="UTF-8"?>
<testsuites tests="2" failures="1" errors="0" skipped="0" time="3.000">
  <testsuite name="SuiteA" tests="1" failures="0" errors="0" skipped="0" time="1.000">
    <testcase classname="A" name="testOne" time="1.000"></testcase>
  </testsuite>
  <testsuite name="SuiteB" tests="1" failures="1" errors="0" skipped="0" time="2.000">
    <testcase classname="B" name="testTwo" time="2.000">
      <failure></failure>
    </testcase>
  </testsuite>
</testsuites>
`
	assert.Equal(t, expected, render(t, suites))
}
