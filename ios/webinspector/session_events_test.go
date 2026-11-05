package webinspector

import (
	"context"
	"errors"
	"testing"
	"time"
)

func newSessionTestClient() *Client {
	return &Client{
		events:   make(chan map[string]any, 8),
		sessions: make(map[string]chan map[string]any),
		errs:     make(chan error, 1),
		disabled: make(chan error, 1),
	}
}

func deliver(t *testing.T, c *Client, destination string, data string) {
	t.Helper()
	if err := c.handleApplicationData(map[string]any{"WIRDestinationKey": destination, "WIRMessageDataKey": data}); err != nil {
		t.Fatalf("handleApplicationData: %v", err)
	}
}

func nextSessionEvent(c *Client, sessionID string) (map[string]any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	return c.NextSessionEvent(ctx, sessionID)
}

// TestSessionEventsAreRoutedByDestination: two CDP sessions on one device must
// not read each other's events. Before, every session read the shared queue.
func TestSessionEventsAreRoutedByDestination(t *testing.T) {
	c := newSessionTestClient()
	c.OpenSessionEvents("A")
	c.OpenSessionEvents("B")

	deliver(t, c, "A", `{"method":"Target.targetCreated","n":1}`)

	if _, err := nextSessionEvent(c, "B"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("session B got an event addressed to A (err=%v)", err)
	}
	event, err := nextSessionEvent(c, "A")
	if err != nil || event["n"] != float64(1) {
		t.Fatalf("session A: event=%v err=%v", event, err)
	}
	if len(c.events) != 0 {
		t.Fatalf("routed event also landed in the shared queue")
	}
}

// TestClosedSessionEventsAreDropped: a session that ended must not leave its
// late events for the next session to pick up as a stale target.
func TestClosedSessionEventsAreDropped(t *testing.T) {
	c := newSessionTestClient()
	c.OpenSessionEvents("A")
	c.closeSessionEvents("A")

	deliver(t, c, "A", `{"method":"Target.targetCreated"}`)

	if len(c.events) != 0 {
		t.Fatalf("event of a closed session landed in the shared queue")
	}
	if _, err := nextSessionEvent(c, "A"); err == nil {
		t.Fatalf("expected an error reading a closed session")
	}
}

// TestUnroutedEventsUseSharedQueue keeps NextEvent working for callers that do
// not open a session queue (eval, js-shell).
func TestUnroutedEventsUseSharedQueue(t *testing.T) {
	c := newSessionTestClient()

	deliver(t, c, "unknown", `{"method":"Console.messageAdded"}`)

	event, err := nextSessionEvent(c, "")
	if err != nil || event["method"] != "Console.messageAdded" {
		t.Fatalf("shared queue: event=%v err=%v", event, err)
	}
}
