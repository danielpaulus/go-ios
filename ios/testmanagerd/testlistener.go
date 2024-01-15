package testmanagerd

import "github.com/danielpaulus/go-ios/ios/nskeyedarchiver"

// Realtime test callbacks
type TestListener struct {
	testFinishedChannel chan struct{}
	err                 nskeyedarchiver.NSError
}

func NewTestListener() *TestListener {
	return &TestListener{
		testFinishedChannel: make(chan struct{}),
	}
}

func (t *TestListener) didFinishExecutingTestPlan() {
	close(t.testFinishedChannel)
}

func (t *TestListener) initializationForUITestingDidFailWithError(err nskeyedarchiver.NSError) {
	t.err = err
	close(t.testFinishedChannel)
}

func (t *TestListener) didFailToBootstrapWithError(err nskeyedarchiver.NSError) {
	t.err = err
	close(t.testFinishedChannel)
}

func (t *TestListener) Done() <-chan struct{} {
	return t.testFinishedChannel
}
