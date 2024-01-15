package testmanagerd

// Realtime test callbacks
type TestListener struct {
	testFinishedChannel chan struct{}
}

func NewTestListener() *TestListener {
	return &TestListener{
		testFinishedChannel: make(chan struct{}),
	}
}

func (t *TestListener) didFinishExecutingTestPlan() {
	close(t.testFinishedChannel)
}

func (t *TestListener) Done() <-chan struct{} {
	return t.testFinishedChannel
}
