package testmanagerd

import (
	"sync"
	"testing"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	"github.com/stretchr/testify/assert"
)

func TestFinishExecutingTestPlan(t *testing.T) {
	t.Parallel()

	t.Run("Wait for test finish with single waiter", func(t *testing.T) {
		testListener := NewTestListener()

		go func() {
			testListener.didFinishExecutingTestPlan()
		}()

		<-testListener.testFinishedChannel
	})

	t.Run("Wait for test finish with multiple waiters", func(t *testing.T) {
		testListener := NewTestListener()

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
		testListener := NewTestListener()

		testListener.initializationForUITestingDidFailWithError(nskeyedarchiver.NSError{
			ErrorCode: 1, Domain: "testdomain", UserInfo: map[string]interface{}{}})

		<-testListener.Done()
		assert.Error(t, testListener.err)
	})

	t.Run("Check error on a failed test run with bootstrap error", func(t *testing.T) {
		testListener := NewTestListener()

		testListener.didFailToBootstrapWithError(nskeyedarchiver.NSError{
			ErrorCode: 1, Domain: "testdomain", UserInfo: map[string]interface{}{}})

		<-testListener.Done()
		assert.Error(t, testListener.err)
	})
}
