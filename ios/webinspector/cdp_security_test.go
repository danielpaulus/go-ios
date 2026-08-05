package webinspector

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

// dialTestWS spins up an httptest server that upgrades a single websocket
// connection and hands both ends (server-side and client-side) back so a test
// can drive waitForTargetID with a real *websocket.Conn.
func dialTestWS(t *testing.T) (server, client *websocket.Conn, cleanup func()) {
	t.Helper()
	upgrader := websocket.Upgrader{}
	connCh := make(chan *websocket.Conn, 1)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("upgrade failed: %v", err)
			return
		}
		connCh <- c
	}))
	url := "ws" + strings.TrimPrefix(srv.URL, "http")
	client, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		srv.Close()
		t.Fatalf("dial failed: %v", err)
	}
	server = <-connCh
	cleanup = func() {
		_ = client.Close()
		_ = server.Close()
		srv.Close()
	}
	return server, client, cleanup
}

// newTestClientWithEvents builds a Client whose NextEvent yields the supplied
// events in order via the internal events channel.
func newTestClientWithEvents(events []map[string]any) *Client {
	c := &Client{
		events:   make(chan map[string]any, len(events)+1),
		errs:     make(chan error, 1),
		disabled: make(chan error, 1),
	}
	for _, e := range events {
		c.events <- e
	}
	return c
}

// TestWaitForTargetID_MalformedParamsNoPanic feeds events whose "params" is a
// string / number / array / nil (i.e. not a JSON object) ahead of a valid
// targetInfo event. Before the fix, the chained non-comma-ok assertion
// event["params"].(map[string]any)["targetInfo"] panics on the very first such
// event. With the fix these are skipped via continue and the valid event's
// targetId is returned.
func TestWaitForTargetID_MalformedParamsNoPanic(t *testing.T) {
	malformed := []any{
		"not-an-object",
		float64(42),
		[]any{"a", "b"},
		nil,
	}
	for _, bad := range malformed {
		bad := bad
		t.Run("params_is_malformed", func(t *testing.T) {
			events := []map[string]any{
				{"method": "Target.targetCreated", "params": bad},
				{
					"method": "Target.targetCreated",
					"params": map[string]any{
						"targetInfo": map[string]any{"targetId": "T-42"},
					},
				},
			}
			client := newTestClientWithEvents(events)
			serverConn, clientConn, cleanup := dialTestWS(t)
			defer cleanup()

			// Drain whatever waitForTargetID writes on the happy path so the
			// write does not block on the socket buffer.
			go func() { _, _, _ = serverConn.ReadMessage() }()

			got := waitForTargetID(context.Background(), client, clientConn)
			if got != "T-42" {
				t.Fatalf("expected targetId %q, got %q", "T-42", got)
			}
		})
	}
}
