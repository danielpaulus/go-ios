package testmanagerd

import (
	"fmt"
	"time"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
)

// A proxy object to intercept incoming DTX messages
// Intercepted messages are converted to channel signals to notify end of connection or forwarded to the dispatch handler for further processing.

type proxyDispatcher struct {
	id              string
	closeChannel    chan interface{}
	closedChannel   chan interface{}
	dispatchHandler *ideInterfaceDtxMessageHandler
}

func (p proxyDispatcher) Dispatch(m dtx.Message) {
	if p.dispatchHandler != nil {
		shouldClose := p.dispatchHandler.handleDtxMessage(m)
		if shouldClose {
			p.Close()
		}
	}
}

func (p *proxyDispatcher) Close() error {
	var signal interface{}
	go func() { p.closeChannel <- signal }()
	select {
	case <-p.closedChannel:
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("Failed closing, exiting due to timeout")
	}
}
