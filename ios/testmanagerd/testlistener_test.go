package testmanagerd

import (
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
