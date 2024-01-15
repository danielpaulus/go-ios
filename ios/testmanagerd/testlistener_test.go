package testmanagerd

import (
	"sync"
	"testing"
)

func TestFinishExecutingTestPlan(t *testing.T) {
	t.Parallel()

	t.Run("Wait for test finish with single waiter", func(t *testing.T) {
		testListener := TestListener{
			testFinishedChannel: make(chan struct{}),
		}

		go func() {
			testListener.didFinishExecutingTestPlan()
		}()

		<-testListener.testFinishedChannel
	})

	t.Run("Wait for test finish with multiple waiters", func(t *testing.T) {
		testListener := TestListener{
			testFinishedChannel: make(chan struct{}),
		}

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
}
