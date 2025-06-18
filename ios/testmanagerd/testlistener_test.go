package testmanagerd

import (
	"errors"
	"io"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	"github.com/stretchr/testify/assert"
)

func TestFinishExecutingTestPlan(t *testing.T) {
	t.Parallel()

	t.Run("Wait for test finish with single waiter", func(t *testing.T) {
		testListener := NewTestListener(os.Stdout, os.Stdout, os.TempDir())

		go func() {
			testListener.didFinishExecutingTestPlan()
		}()

		<-testListener.finished
	})

	t.Run("Wait for test finish with multiple waiters", func(t *testing.T) {
		testListener := NewTestListener(os.Stdout, os.Stdout, os.TempDir())

		var wg sync.WaitGroup
		wg.Add(2)

		testListener.didFinishExecutingTestPlan()

		go func() {
			<-testListener.Done()
			wg.Done()
		}()

		go func() {
			<-testListener.Done()
			wg.Done()
		}()

		wg.Wait()
	})

	t.Run("Check error on a failed test run", func(t *testing.T) {
		testListener := NewTestListener(os.Stdout, os.Stdout, os.TempDir())

		testListener.initializationForUITestingDidFailWithError(nskeyedarchiver.NSError{
			ErrorCode: 1, Domain: "testdomain", UserInfo: map[string]interface{}{}})

		<-testListener.Done()
		assert.Error(t, testListener.err)
	})

	t.Run("Check error on a failed test run with bootstrap error", func(t *testing.T) {
		testListener := NewTestListener(os.Stdout, os.Stdout, os.TempDir())

		testListener.didFailToBootstrapWithError(nskeyedarchiver.NSError{
			ErrorCode: 1, Domain: "testdomain", UserInfo: map[string]interface{}{}})

		<-testListener.Done()
		assert.Error(t, testListener.err)
	})

	t.Run("Check test log callbacks", func(t *testing.T) {
		logWriter := assertionWriter{}
		debugLogWriter := assertionWriter{}

		testListener := NewTestListener(&logWriter, &debugLogWriter, os.TempDir())

		testListener.LogMessage("log")
		testListener.LogDebugMessage("debug")

		assert.True(t, logWriter.hasBytes, "Test listener must receive test logs")
		assert.True(t, debugLogWriter.hasBytes, "Test listener must receive test debug logs")
	})

	t.Run("Check test suite creation", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testSuiteFinished("mysuite", "2024-01-16 15:36:44 +0000", 0, 0, 0, 0, 0, 0, 1.0, 1.0)
		testListener.testSuiteDidStart("mysuite2", "2024-01-17 16:36:45 +0000")
		testListener.testSuiteFinished("mysuite2", "2024-01-17 16:36:46 +0000", 0, 0, 0, 0, 0, 0, 1.0, 1.0)

		firstSuite := testListener.TestSuites[0]
		secondSuite := testListener.TestSuites[1]

		assert.Equal(t, 2, len(testListener.TestSuites))
		assert.Equal(t, 2024, firstSuite.StartDate.Year())
		assert.Equal(t, time.Month(1), firstSuite.StartDate.Month())
		assert.Equal(t, 16, firstSuite.StartDate.Day())
		assert.Equal(t, "mysuite", firstSuite.Name)

		assert.Equal(t, 2, len(testListener.TestSuites))
		assert.Equal(t, 2024, secondSuite.StartDate.Year())
		assert.Equal(t, time.Month(1), secondSuite.StartDate.Month())
		assert.Equal(t, 17, secondSuite.StartDate.Day())
		assert.Equal(t, "mysuite2", secondSuite.Name)
	})

	t.Run("Check test case creation", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("mysuite", "mymethod")

		assert.Equal(t, 1, len(testListener.runningTestSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, TestCase{
			ClassName:  "mysuite",
			MethodName: "mymethod",
		}, testListener.runningTestSuite.TestCases[0])
	})

	t.Run("Check test start invalid date", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		testListener.testSuiteDidStart("mysuite", "INVALIDDATE")
		testListener.testSuiteFinished("mysuite", "INVALIDDATE", 0, 0, 0, 0, 0, 0, 1.0, 1.0)

		assert.Equal(t, time.Now().Year(), testListener.TestSuites[0].StartDate.Year())
		assert.Equal(t, time.Now().Year(), testListener.TestSuites[0].EndDate.Year())
	})

	t.Run("Check test case failure", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("mysuite", "mymethod")

		testListener.testCaseFailedForClass("mysuite", "mymethod", "error", "file://app.swift", 123)
		testListener.testSuiteFinished("mysuite", "2024-01-16 15:37:43 +0000", 0, 0, 0, 0, 0, 0, 1.0, 1.0)

		assert.Equal(t, 1, len(testListener.TestSuites[0].TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, TestCaseStatus("failed"), testListener.TestSuites[0].TestCases[0].Status)
		assert.Equal(t, "error", testListener.TestSuites[0].TestCases[0].Err.Message)
		assert.Equal(t, "file://app.swift", testListener.TestSuites[0].TestCases[0].Err.File)
		assert.Equal(t, uint64(123), testListener.TestSuites[0].TestCases[0].Err.Line)
	})

	t.Run("Check test case finish", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("mysuite", "mymethod")
		testListener.testCaseDidFinishForTest("mysuite", "mymethod", "passed", 1.0)

		t.Run("Check running test suite is saved on FinishWithError", func(t *testing.T) {
			testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

			testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
			testListener.testCaseDidStartForClass("mysuite", "mymethod")
			testListener.FinishWithError(errors.New("test error"))

			assert.Equal(t, 1, len(testListener.TestSuites))
			assert.Equal(t, "mysuite", testListener.TestSuites[0].Name)
			assert.Equal(t, 1, len(testListener.TestSuites[0].TestCases))
		})

		assert.Equal(t, 1, len(testListener.runningTestSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, TestCaseStatus("passed"), testListener.runningTestSuite.TestCases[0].Status)
		assert.Equal(t, 1.0, testListener.runningTestSuite.TestCases[0].Duration.Seconds())
	})

	t.Run("Check test suite finish", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("mysuite", "mymethod1")
		testListener.testCaseDidFinishForTest("mysuite", "mymethod1", "passed", 1.0)
		testListener.testCaseDidStartForClass("mysuite", "mymethod2")
		testListener.testCaseDidFinishForTest("mysuite", "mymethod2", "passed", 1.0)
		testListener.testSuiteFinished("mysuite", "2024-01-16 15:36:44 +0000", 2, 0, 0, 0, 0, 0, 1.0, 2.0)

		assert.Equal(t, 2, len(testListener.TestSuites[0].TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, TestCaseStatus("passed"), testListener.TestSuites[0].TestCases[0].Status)
		assert.Equal(t, TestCaseStatus("passed"), testListener.TestSuites[0].TestCases[1].Status)
		assert.Equal(t, 2.0, testListener.TestSuites[0].TotalDuration.Seconds())
		assert.Equal(t, 1, testListener.TestSuites[0].EndDate.Second()-testListener.TestSuites[0].StartDate.Second())
	})

	t.Run("Check test case stall", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("mysuite", "mymethod1")
		testListener.testCaseStalled("mysuite", "mymethod1", "file://app.swift", 123)
		testListener.testCaseDidFinishForTest("mysuite", "mymethod1", "failed", 1.0)

		assert.Equal(t, 1, len(testListener.runningTestSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, TestCaseStatus("stalled"), testListener.runningTestSuite.TestCases[0].Status)
	})

	t.Run("Check test case with attachments", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		payload := []byte("test")
		attachments := make([]nskeyedarchiver.XCTAttachment, 1)
		attachments[0] = nskeyedarchiver.XCTAttachment{
			Payload: payload,
		}

		// Suite start
		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")

		// Test with 0 attachments
		testListener.testCaseDidStartForClass("mysuite", "mymethod1")
		testListener.testCaseDidFinishForTest("mysuite", "mymethod1", "failed", 1.0)

		// Test with 1 attachment
		testListener.testCaseDidStartForClass("mysuite", "mymethod2")
		// Attachments of activity records are reported under a special test class named "none". This is the default behavior defined by Apple.
		// We have a safe guard to auto correct the test case information by keeping track of the active test case
		testListener.testCaseFinished("none", "none", nskeyedarchiver.XCActivityRecord{
			Finish:       nskeyedarchiver.NSDate{},
			Start:        nskeyedarchiver.NSDate{},
			Title:        "test",
			UUID:         nskeyedarchiver.NSUUID{},
			ActivityType: "userDefined",
			Attachments:  attachments,
		})
		testListener.testCaseDidFinishForTest("mysuite", "mymethod2", "failed", 1.0)

		// Suite end
		testListener.testSuiteFinished("mysuite", "2024-01-16 15:36:44 +0000", 0, 0, 0, 0, 0, 0, 1.0, 1.0)

		assert.Equal(t, 2, len(testListener.TestSuites[0].TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, 0, len(testListener.TestSuites[0].TestCases[0].Attachments), "First test must have 0 attachments")
		assert.Equal(t, 1, len(testListener.TestSuites[0].TestCases[1].Attachments), "Second test must have 1 attachment")

		path := testListener.TestSuites[0].TestCases[1].Attachments[0].Path
		attachment, err := os.ReadFile(path)
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		assert.Equal(t, "test", string(attachment), "Attachment content should be put in a file")
	})

	t.Run("Check test case without suite initialization", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		// This should trigger the nil pointer dereference error if not handled properly
		// Call testCaseDidStartForClass without first calling testSuiteDidStart
		assert.NotPanics(t, func() {
			testListener.testCaseDidStartForClass("mysuite", "mymethod")
		})

		// Verify that a new suite was created automatically
		assert.Equal(t, 1, len(testListener.TestSuites), "A test suite should be created automatically")

		// Verify the test case was added to the newly created suite
		assert.Equal(t, 1, len(testListener.TestSuites[0].TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, TestCase{
			ClassName:  "mysuite",
			MethodName: "mymethod",
		}, testListener.TestSuites[0].TestCases[0])
	})
}

type assertionWriter struct {
	hasBytes bool
}

func (w *assertionWriter) Write(p []byte) (n int, err error) {
	if len(p) > 0 {
		w.hasBytes = true
	}

	return len(p), nil
}
