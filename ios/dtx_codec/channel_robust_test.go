package dtx

import (
	"context"
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

// A reply arriving after the caller timed out must be dropped, not block the
// reader goroutine on the abandoned waiter channel.
func TestLateReplyAfterTimeoutDoesNotBlock(t *testing.T) {
	channel := newTestChannel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // caller "times out" immediately

	waiter := make(chan Message, 1)
	channel.AddResponseWaiter(42, waiter)
	// Simulate the sendAndAwaitReply timeout path.
	select {
	case <-ctx.Done():
		channel.removeResponseWaiter(42)
	}

	done := make(chan struct{})
	go func() {
		channel.Dispatch(Message{Identifier: 42, ConversationIndex: 1})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Dispatch blocked on a late reply with no waiter")
	}
}

// A reply for an identifier that never had a waiter must not block either
// (the old code sent on a nil channel, which blocks forever).
func TestUnknownReplyDoesNotBlock(t *testing.T) {
	channel := newTestChannel()

	done := make(chan struct{})
	go func() {
		channel.Dispatch(Message{Identifier: 999, ConversationIndex: 1})
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Dispatch blocked on a reply with an unknown identifier")
	}
}
