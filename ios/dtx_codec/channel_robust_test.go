package dtx

import (
	"testing"
	"time"
)

// newTestChannel is shared with hardening_channel_internal_test.go.

func methodInvocationMessage(identifier int, selector string) Message {
	return Message{
		Identifier:    identifier,
		PayloadHeader: PayloadHeader{MessageType: Methodinvocation},
		Payload:       []interface{}{selector},
	}
}

// Dispatch must never block the connection's reader goroutine, even when a
// registered selector has no active ReceiveMethodCall consumer.
func TestDispatchDoesNotBlockWithoutConsumer(t *testing.T) {
	channel := newTestChannel()
	channel.RegisterMethodForRemote("someEvent:")

	done := make(chan struct{})
	go func() {
		// Far more events than the per-selector buffer holds.
		for i := 0; i < registeredMethodBuffer*4; i++ {
			channel.Dispatch(methodInvocationMessage(i, "someEvent:"))
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Dispatch blocked with no consumer for a registered selector")
	}
}

// When the queue overflows, the OLDEST events are dropped so a consumer that
// comes back always sees the most recent state.
func TestDispatchDropsOldestOnOverflow(t *testing.T) {
	channel := newTestChannel()
	channel.RegisterMethodForRemote("someEvent:")

	total := registeredMethodBuffer + 5
	for i := 0; i < total; i++ {
		channel.Dispatch(methodInvocationMessage(i, "someEvent:"))
	}

	first := channel.ReceiveMethodCall("someEvent:")
	if first.Identifier != total-registeredMethodBuffer {
		t.Fatalf("expected oldest surviving event %d, got %d", total-registeredMethodBuffer, first.Identifier)
	}
	// Drain the rest; the newest event must have survived.
	last := first
	for i := 1; i < registeredMethodBuffer; i++ {
		last = channel.ReceiveMethodCall("someEvent:")
	}
	if last.Identifier != total-1 {
		t.Fatalf("expected newest event %d to survive, got %d", total-1, last.Identifier)
	}
}

// A timed-out caller must deregister its abandoned waiter AND any partial
// defragmenter for that identifier, so neither leaks in responseWaiters /
// defragmenters — and a late reply for it must still not block the reader
// goroutine. (Non-blocking on a nil/unknown waiter is already covered by
// TestChannelDispatch_ResponseNoWaiter_NoHang / TestSendResponse_NilChannel_NoBlock.)
func TestTimeoutCleansUpWaiterAndDefragmenter(t *testing.T) {
	channel := newTestChannel()

	// A pending request left a waiter and a partial (fragmented) reply behind.
	channel.AddResponseWaiter(42, make(chan Message, 1))
	channel.defragmenters[42] = &FragmentDecoder{}

	if _, ok := channel.responseWaiters[42]; !ok {
		t.Fatal("precondition: waiter should be registered")
	}
	if _, ok := channel.defragmenters[42]; !ok {
		t.Fatal("precondition: defragmenter should be registered")
	}

	// The sendAndAwaitReply timeout/send-error path deregisters both.
	channel.removeResponseWaiter(42)

	if _, ok := channel.responseWaiters[42]; ok {
		t.Fatal("responseWaiters leaked after timeout cleanup")
	}
	if _, ok := channel.defragmenters[42]; ok {
		t.Fatal("defragmenters leaked after timeout cleanup")
	}

	// A late reply for the now-abandoned identifier must not block the reader.
	done := make(chan struct{})
	go func() {
		channel.Dispatch(Message{Identifier: 42, ConversationIndex: 1})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Dispatch blocked on a late reply after timeout cleanup")
	}
}
