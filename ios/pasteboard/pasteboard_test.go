package pasteboard

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// fakeConn is a scripted xpcConn for exercising sendReceive without a device.
// ReceiveOnServerClientStream replays frames in order; a frame may be empty (to
// simulate a heartbeat/ack), carry a body, or carry an error. When the script
// is exhausted it blocks until Close is called (simulating a wedged daemon),
// then returns errClosed.
type fakeConn struct {
	frames []frame

	mu        sync.Mutex
	idx       int
	sent      []map[string]interface{}
	closed    chan struct{}
	closeOnce sync.Once
}

type frame struct {
	body map[string]interface{}
	err  error
}

var errClosed = errors.New("connection closed")

func newFakeConn(frames ...frame) *fakeConn {
	return &fakeConn{frames: frames, closed: make(chan struct{})}
}

func (f *fakeConn) Send(data map[string]interface{}, _ ...uint32) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, data)
	return nil
}

func (f *fakeConn) ReceiveOnServerClientStream() (map[string]interface{}, error) {
	f.mu.Lock()
	if f.idx < len(f.frames) {
		fr := f.frames[f.idx]
		f.idx++
		f.mu.Unlock()
		return fr.body, fr.err
	}
	f.mu.Unlock()
	// Script exhausted: block like a wedged daemon until Close unblocks us.
	<-f.closed
	return nil, errClosed
}

func (f *fakeConn) Close() error {
	f.closeOnce.Do(func() { close(f.closed) })
	return nil
}

func newTestConn(timeout time.Duration, frames ...frame) *Connection {
	return &Connection{conn: newFakeConn(frames...), timeout: timeout}
}

func TestSendReceiveSkipsEmptyFrames(t *testing.T) {
	tests := []struct {
		name   string
		frames []frame
		want   string
	}{
		{
			name:   "reply on first frame",
			frames: []frame{{body: map[string]interface{}{"command": "PULL_REPLY", "marker": "first"}}},
			want:   "first",
		},
		{
			name: "skips leading nil and empty frames",
			frames: []frame{
				{body: nil},
				{body: map[string]interface{}{}},
				{body: map[string]interface{}{"command": "PULL_REPLY", "marker": "real"}},
			},
			want: "real",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestConn(time.Second, tt.frames...)
			reply, err := c.sendReceive(map[string]interface{}{"command": "PULL"})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got, _ := reply["marker"].(string); got != tt.want {
				t.Fatalf("expected marker %q, got %q", tt.want, got)
			}
		})
	}
}

func TestSendReceivePropagatesReceiveError(t *testing.T) {
	wantErr := errors.New("stream broke")
	c := newTestConn(time.Second, frame{err: wantErr})
	_, err := c.sendReceive(map[string]interface{}{"command": "PULL"})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped %v, got %v", wantErr, err)
	}
}

func TestSendReceiveTimesOutAndClosesConn(t *testing.T) {
	// No frames => ReceiveOnServerClientStream blocks => the timeout fires and
	// must close the connection to release the blocked reader.
	fake := newFakeConn() // empty script: blocks until Close
	c := &Connection{conn: fake, timeout: 20 * time.Millisecond}

	_, err := c.sendReceive(map[string]interface{}{"command": "PULL"})
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}

	select {
	case <-fake.closed:
		// Connection was closed on timeout, as required to unblock the reader.
	case <-time.After(time.Second):
		t.Fatal("timeout did not close the connection; reader would leak")
	}
}

// Error frames as captured from a real device (iPhone14,2 / iOS 18.4.1, DDI
// 27A5194q): a malformed SET and an unknown command verb.
func TestSendReceiveSurfacesErrorFrames(t *testing.T) {
	tests := []struct {
		name    string
		frame   map[string]interface{}
		wantSub string
	}{
		{
			name: "NSCocoaErrorDomain decode error with NSDebugDescription",
			frame: map[string]interface{}{
				"error": map[string]interface{}{
					"code":   int64(4864),
					"domain": "NSCocoaErrorDomain",
					"userInfo": map[string]interface{}{
						"NSCodingPath":       `[CodingKeys(stringValue: "items", intValue: nil)]`,
						"NSDebugDescription": "array required here",
					},
				},
			},
			wantSub: "array required here",
		},
		{
			name: "CoreDeviceError with NSLocalizedDescription",
			frame: map[string]interface{}{
				"error": map[string]interface{}{
					"code":   int64(-1),
					"domain": "com.apple.dt.CoreDeviceError",
					"userInfo": map[string]interface{}{
						"NSLocalizedDescription": "Unknown pasteboard command: FROBNICATE",
					},
				},
			},
			wantSub: "Unknown pasteboard command",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestConn(time.Second, frame{body: tt.frame})
			_, err := c.sendReceive(map[string]interface{}{"command": "SET"})
			if err == nil {
				t.Fatal("expected an error for an error reply frame, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantSub) {
				t.Fatalf("error %q does not contain %q", err, tt.wantSub)
			}
		})
	}
}

func TestSetSurfacesServiceError(t *testing.T) {
	c := newTestConn(time.Second, frame{body: map[string]interface{}{
		"error": map[string]interface{}{
			"code":     int64(4864),
			"domain":   "NSCocoaErrorDomain",
			"userInfo": map[string]interface{}{"NSDebugDescription": "array required here"},
		},
	}})
	if err := c.SetText("x"); err == nil {
		t.Fatal("expected SetText to surface the service error, got nil")
	}
}

func TestGetTextSurfacesServiceError(t *testing.T) {
	c := newTestConn(time.Second, frame{body: map[string]interface{}{
		"error": map[string]interface{}{
			"code":     int64(-1),
			"domain":   "com.apple.dt.CoreDeviceError",
			"userInfo": map[string]interface{}{"NSLocalizedDescription": "boom"},
		},
	}})
	_, _, err := c.GetText()
	if err == nil {
		t.Fatal("expected GetText to surface the service error, got nil")
	}
}

func TestTextItem(t *testing.T) {
	item := TextItem("héllo 🌍")
	if len(item.Types) != 3 {
		t.Fatalf("expected 3 UTIs, got %d", len(item.Types))
	}
	for _, uti := range textUTIs {
		got, ok := item.Data[uti]
		if !ok {
			t.Fatalf("missing data for UTI %q", uti)
		}
		if string(got) != "héllo 🌍" {
			t.Fatalf("UTI %q: expected UTF-8 bytes of input, got %q", uti, got)
		}
	}
}

func TestDataItem(t *testing.T) {
	raw := []byte{0x01, 0x02, 0x03}
	item := DataItem(UTIURL, raw)
	if len(item.Types) != 1 || item.Types[0] != UTIURL {
		t.Fatalf("unexpected types: %v", item.Types)
	}
	if string(item.Data[UTIURL]) != string(raw) {
		t.Fatalf("unexpected data: %v", item.Data[UTIURL])
	}
}

func TestSnapshotTextFromPullReply(t *testing.T) {
	reply := map[string]interface{}{
		"command": "PULL_REPLY",
		"pasteboard": map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{
					"types": []interface{}{UTIUTF8PlainText},
					"data": map[string]interface{}{
						UTIUTF8PlainText: map[string]interface{}{"data": []byte("clipboard text")},
					},
				},
			},
		},
	}
	text, ok := snapshotText(reply)
	if !ok {
		t.Fatal("expected text to be extracted")
	}
	if text != "clipboard text" {
		t.Fatalf("unexpected text: %q", text)
	}
}

func TestSnapshotTextPrefersUTF8Order(t *testing.T) {
	reply := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"data": map[string]interface{}{
					UTIText:          map[string]interface{}{"data": []byte("plain")},
					UTIUTF8PlainText: map[string]interface{}{"data": []byte("utf8")},
				},
			},
		},
	}
	text, ok := snapshotText(reply)
	if !ok {
		t.Fatal("expected text to be extracted")
	}
	if text != "utf8" {
		t.Fatalf("expected UTF-8 UTI to win, got %q", text)
	}
}

func TestSnapshotTextPromisedItemHasNoInlineData(t *testing.T) {
	reply := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"data": map[string]interface{}{
					UTIUTF8PlainText: map[string]interface{}{
						"isPromised":  true,
						"isAvailable": false,
						"size":        int64(42),
					},
				},
			},
		},
	}
	if _, ok := snapshotText(reply); ok {
		t.Fatal("expected no text for a promised-only item")
	}
}

func TestSnapshotTextEmpty(t *testing.T) {
	if _, ok := snapshotText(map[string]interface{}{}); ok {
		t.Fatal("expected no text from empty reply")
	}
}
