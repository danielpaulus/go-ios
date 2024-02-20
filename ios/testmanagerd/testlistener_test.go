package testmanagerd

import (
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

		<-testListener.executionFinished
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

		assert.Equal(t, 2024, testListener.TestSuite.StartDate.Year())
		assert.Equal(t, time.Month(1), testListener.TestSuite.StartDate.Month())
		assert.Equal(t, 16, testListener.TestSuite.StartDate.Day())
		assert.Equal(t, "mysuite", testListener.TestSuite.Name)
	})

	t.Run("Check test case creation", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("myclass", "mymethod")

		assert.Equal(t, 1, len(testListener.TestSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, TestCase{
			ClassName:  "myclass",
			MethodName: "mymethod",
		}, testListener.TestSuite.TestCases[0])
	})

	t.Run("Check test start invalid date", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		testListener.testSuiteDidStart("mysuite", "INVALIDDATE")
		testListener.testSuiteFinished("mysuite", "INVALIDDATE", 0, 0, 0, 0, 0, 0, 1.0, 1.0)

		assert.Equal(t, time.Now().Year(), testListener.TestSuite.StartDate.Year())
		assert.Equal(t, time.Now().Year(), testListener.TestSuite.EndDate.Year())
	})

	t.Run("Check test case failure", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("myclass", "mymethod")

		testListener.testCaseFailedForClass("myclass", "mymethod", "error", "file://app.swift", 123)

		assert.Equal(t, 1, len(testListener.TestSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, TestCaseStatus("failed"), testListener.TestSuite.TestCases[0].Status)
		assert.Equal(t, "error", testListener.TestSuite.TestCases[0].Err.Message)
		assert.Equal(t, "file://app.swift", testListener.TestSuite.TestCases[0].Err.File)
		assert.Equal(t, uint64(123), testListener.TestSuite.TestCases[0].Err.Line)
	})

	t.Run("Check test case finish", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("myclass", "mymethod")
		testListener.testCaseDidFinishForTest("myclass", "mymethod", "passed", 1.0)

		assert.Equal(t, 1, len(testListener.TestSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, TestCaseStatus("passed"), testListener.TestSuite.TestCases[0].Status)
		assert.Equal(t, 1.0, testListener.TestSuite.TestCases[0].Duration.Seconds())
	})

	t.Run("Check test suite finish", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("myclass", "mymethod1")
		testListener.testCaseDidFinishForTest("myclass", "mymethod1", "passed", 1.0)
		testListener.testCaseDidStartForClass("myclass", "mymethod2")
		testListener.testCaseDidFinishForTest("myclass", "mymethod2", "passed", 1.0)
		testListener.testSuiteFinished("myclass", "2024-01-16 15:36:44 +0000", 2, 0, 0, 0, 0, 0, 1.0, 2.0)

		assert.Equal(t, 2, len(testListener.TestSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, TestCaseStatus("passed"), testListener.TestSuite.TestCases[0].Status)
		assert.Equal(t, TestCaseStatus("passed"), testListener.TestSuite.TestCases[1].Status)
		assert.Equal(t, 2.0, testListener.TestSuite.TotalDuration.Seconds())
		assert.Equal(t, 1, testListener.TestSuite.EndDate.Second()-testListener.TestSuite.StartDate.Second())
	})

	t.Run("Check test case stall", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("myclass", "mymethod1")
		testListener.testCaseStalled("myclass", "mymethod1", "file://app.swift", 123)
		testListener.testCaseDidFinishForTest("myclass", "mymethod1", "failed", 1.0)

		assert.Equal(t, 1, len(testListener.TestSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, TestCaseStatus("stalled"), testListener.TestSuite.TestCases[0].Status)
	})

	t.Run("Check test case with attachments", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard, os.TempDir())

		payload := []byte("test")
		attachments := make([]nskeyedarchiver.XCTAttachment, 1)
		attachments[0] = nskeyedarchiver.XCTAttachment{
			Payload: payload,
		}
		testListener.testCaseFinished("none", "none", nskeyedarchiver.XCActivityRecord{
			Finish:       nskeyedarchiver.NSDate{},
			Start:        nskeyedarchiver.NSDate{},
			Title:        "test",
			UUID:         nskeyedarchiver.NSUUID{},
			ActivityType: "userDefined",
			Attachments:  attachments,
		})

		assert.Equal(t, 1, len(testListener.TestSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, 1, len(testListener.TestSuite.TestCases[0].Attachments), "Test must have 1 attachment")

		path := testListener.TestSuite.TestCases[0].Attachments[0].Path
		attachment, err := os.ReadFile(path)
		assert.NoError(t, err)
		defer os.RemoveAll(path)

		assert.Equal(t, "test", string(attachment), "Attachment content should be put in a file")
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
