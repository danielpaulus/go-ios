package testmanagerd

import (
	"fmt"
	"time"

	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
)

type ProxyDispatcher struct {
	dtxConnection   *dtx.Connection
	id              string
	closeChannel    chan interface{}
	closedChannel   chan interface{}
	dispatchHandler *DispatchHandler
}

func (p ProxyDispatcher) Dispatch(m dtx.Message) {
	if p.dispatchHandler != nil {
		p.dispatchHandler.HandleDispatch(m, &p)
	}
}

func (p *ProxyDispatcher) Close() error {
	var signal interface{}
	go func() { p.closeChannel <- signal }()
	select {
	case <-p.closedChannel:
		return nil
	case <-time.After(10 * time.Second):
		return fmt.Errorf("Failed closing, exiting due to timeout")
	}
}
