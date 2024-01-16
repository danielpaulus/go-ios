package testmanagerd

import (
	"io"
	"os"
	"sync"
	"testing"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	"github.com/stretchr/testify/assert"
)

func TestFinishExecutingTestPlan(t *testing.T) {
	t.Parallel()

	t.Run("Wait for test finish with single waiter", func(t *testing.T) {
		testListener := NewTestListener(os.Stdout, os.Stdout)

		go func() {
			testListener.didFinishExecutingTestPlan()
		}()

		<-testListener.testFinishedChannel
	})

	t.Run("Wait for test finish with multiple waiters", func(t *testing.T) {
		testListener := NewTestListener(os.Stdout, os.Stdout)

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
		testListener := NewTestListener(os.Stdout, os.Stdout)

		testListener.initializationForUITestingDidFailWithError(nskeyedarchiver.NSError{
			ErrorCode: 1, Domain: "testdomain", UserInfo: map[string]interface{}{}})

		<-testListener.Done()
		assert.Error(t, testListener.err)
	})

	t.Run("Check error on a failed test run with bootstrap error", func(t *testing.T) {
		testListener := NewTestListener(os.Stdout, os.Stdout)

		testListener.didFailToBootstrapWithError(nskeyedarchiver.NSError{
			ErrorCode: 1, Domain: "testdomain", UserInfo: map[string]interface{}{}})

		<-testListener.Done()
		assert.Error(t, testListener.err)
	})

	t.Run("Check test log callbacks", func(t *testing.T) {
		logWriter := assertionWriter{}
		debugLogWriter := assertionWriter{}

		testListener := NewTestListener(&logWriter, &debugLogWriter)

		testListener.LogMessage("log")
		testListener.LogDebugMessage("debug")

		assert.True(t, logWriter.hasBytes, "Test listener must receive test logs")
		assert.True(t, debugLogWriter.hasBytes, "Test listener must receive test debug logs")
	})

	t.Run("Check test suite creation", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard)

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")

		assert.Equal(t, "2024", testListener.testSuite.StartDate.Year())
		assert.Equal(t, "01", testListener.testSuite.StartDate.Month())
		assert.Equal(t, "16", testListener.testSuite.StartDate.Day())
		assert.Equal(t, "mysuite", testListener.testSuite.Name)
	})

	t.Run("Check test case creation", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard)

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("myclass", "mymethod")

		assert.Equal(t, 1, len(testListener.testSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, "myclass", testListener.testSuite.TestCases[0].ClassName)
		assert.Equal(t, "mymethod", testListener.testSuite.TestCases[0].MethodName)
	})

	t.Run("Check test case failure", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard)

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("myclass", "mymethod")

		testListener.testCaseFailedForClass("myclass", "mymethod", "error", "file://app.swift", 123)

		assert.Equal(t, 1, len(testListener.testSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, "failed", testListener.testSuite.TestCases[0].Status.Status)
		assert.Equal(t, "error", testListener.testSuite.TestCases[0].Status.Err.Message)
		assert.Equal(t, "file://app.swift", testListener.testSuite.TestCases[0].Status.Err.File)
		assert.Equal(t, uint64(123), testListener.testSuite.TestCases[0].Status.Err.Line)
	})

	t.Run("Check test case finish", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard)

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("myclass", "mymethod")
		testListener.testCaseDidFinishForTest("myclass", "mymethod", "passed", 1.0)

		assert.Equal(t, 1, len(testListener.testSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, "passed", testListener.testSuite.TestCases[0].Status.Status)
		assert.Equal(t, 1.0, testListener.testSuite.TestCases[0].Duration.Seconds())
	})

	t.Run("Check test suite finish", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard)

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("myclass", "mymethod1")
		testListener.testCaseDidFinishForTest("myclass", "mymethod1", "passed", 1.0)
		testListener.testCaseDidStartForClass("myclass", "mymethod2")
		testListener.testCaseDidFinishForTest("myclass", "mymethod2", "passed", 1.0)
		testListener.testSuiteFinished("myclass", "2024-01-16 15:36:44 +0000", 2, 0, 0, 0, 0, 0, 1.0, 2.0)

		assert.Equal(t, 2, len(testListener.testSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, "passed", testListener.testSuite.TestCases[0].Status.Status)
		assert.Equal(t, "passed", testListener.testSuite.TestCases[1].Status.Status)
		assert.Equal(t, 2.0, testListener.testSuite.TotalDuration.Seconds())
		assert.Equal(t, 1, testListener.testSuite.EndDate.Second()-testListener.testSuite.StartDate.Second())
	})

	t.Run("Check test case stall", func(t *testing.T) {
		testListener := NewTestListener(io.Discard, io.Discard)

		testListener.testSuiteDidStart("mysuite", "2024-01-16 15:36:43 +0000")
		testListener.testCaseDidStartForClass("myclass", "mymethod1")
		testListener.testCaseStalled("myclass", "mymethod1", "file://app.swift", 123)
		testListener.testCaseDidFinishForTest("myclass", "mymethod2", "failed", 1.0)

		assert.Equal(t, 1, len(testListener.testSuite.TestCases), "TestCase must be appended to list of test cases")
		assert.Equal(t, "stalled", testListener.testSuite.TestCases[0].Status.Status)
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
