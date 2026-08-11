package dtx

// Internal (white-box) regression tests for the Channel.Dispatch hardening.
// These need package-internal access because a *Channel cannot be constructed
// device-free through the exported API, and the guards being tested live on the
// reader's dispatch path.

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// noopDispatcher is a Dispatcher that records nothing; it just must not panic.
type noopDispatcher struct{}

func (noopDispatcher) Dispatch(msg Message) {}

func newTestChannel() *Channel {
	return &Channel{
		channelName:       "test",
		messageDispatcher: noopDispatcher{},
		responseWaiters:   map[int]chan Message{},
		defragmenters:     map[int]*FragmentDecoder{},
		registeredMethods: map[string]chan Message{},
		timeout:           time.Second,
	}
}

// TestChannelDispatch_EmptyMethodInvocation_NoPanic feeds a Methodinvocation
// message with an empty payload. Before the fix, Dispatch asserted
// msg.Payload[0].(string) and panicked with index out of range.
func TestChannelDispatch_EmptyMethodInvocation_NoPanic(t *testing.T) {
	ch := newTestChannel()
	msg := Message{
		PayloadHeader: PayloadHeader{MessageType: Methodinvocation},
		Payload:       []interface{}{}, // non-nil, empty
	}
	assert.NotPanics(t, func() {
		ch.Dispatch(msg)
	})
}

// TestChannelDispatch_NonStringMethodInvocation_NoPanic feeds a non-string
// payload[0] to the method-invocation selector path.
func TestChannelDispatch_NonStringMethodInvocation_NoPanic(t *testing.T) {
	ch := newTestChannel()
	msg := Message{
		PayloadHeader: PayloadHeader{MessageType: Methodinvocation},
		Payload:       []interface{}{uint32(1)},
	}
	assert.NotPanics(t, func() {
		ch.Dispatch(msg)
	})
}

// TestChannelDispatch_ResponseNoWaiter_NoHang delivers a response
// (ConversationIndex > 0) whose Identifier has no registered waiter. The map
// lookup yields a nil channel; before the fix, `d.responseWaiters[id] <- msg`
// blocked forever. The guard must make Dispatch return promptly.
func TestChannelDispatch_ResponseNoWaiter_NoHang(t *testing.T) {
	ch := newTestChannel()
	msg := Message{
		ConversationIndex: 1, // response
		Identifier:        9999,
		PayloadHeader:     PayloadHeader{MessageType: ResponseWithReturnValueInPayload},
	}
	done := make(chan struct{})
	go func() {
		defer close(done)
		ch.Dispatch(msg)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Dispatch blocked on a nil response waiter channel (hang)")
	}
}

// TestSendResponse_NilChannel_NoBlock directly pins the helper: a nil channel
// must be a no-op, and a full/unbuffered channel with no receiver must not
// block (non-blocking send).
func TestSendResponse_NilChannel_NoBlock(t *testing.T) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		sendResponse(nil, Message{})                // nil channel: no-op
		sendResponse(make(chan Message), Message{}) // no receiver: dropped, not blocked
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("sendResponse blocked")
	}
}
