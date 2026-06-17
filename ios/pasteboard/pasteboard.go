// Package pasteboard reads and writes the device pasteboard (clipboard) on
// iOS 17+ devices through the com.apple.coredevice.pasteboardservice RemoteXPC
// service.
//
// The service speaks plain XPC dicts directly: there is no CoreDevice.* invoke
// envelope and no featureIdentifier/messageType wrapper. The top-level
// "command" field drives dispatch. Setting the pasteboard needs no media
// stream, which makes it a reliable path for putting arbitrary Unicode text
// (any language, emoji, long text) onto a device.
package pasteboard

import (
	"fmt"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/xpc"
)

const logModule = "go-ios/pasteboard"

// requestTimeout bounds how long we wait for a reply before assuming the
// pasteboard daemon is wedged. A wedged daemon leaves the RemoteXPC channel
// half-open and the read blocks forever; the only device-side recovery is a
// reboot.
const requestTimeout = 10 * time.Second

// ServiceName is the RemoteXPC service name for the device pasteboard.
const ServiceName = "com.apple.coredevice.pasteboardservice"

// GeneralPasteboard is the name of the general (default) pasteboard.
const GeneralPasteboard = "general"

// Command verbs understood by the service. Only PULL (read) and SET (write with
// immediate data) are needed for copy/paste.
const (
	commandPull = "PULL"
	commandSet  = "SET"
)

// Standard Uniform Type Identifiers for plain text. The same UTF-8 bytes are
// published under all three.
const (
	UTIUTF8PlainText = "public.utf8-plain-text"
	UTIPlainText     = "public.plain-text"
	UTIText          = "public.text"
	UTIURL           = "public.url"
)

// textUTIs is the ordered set of UTIs used both when writing text and when
// extracting it from a snapshot.
var textUTIs = []string{UTIUTF8PlainText, UTIPlainText, UTIText}

// Connection represents a connection to the pasteboard service on an iOS 17+
// device. It is not safe for concurrent use.
type Connection struct {
	conn   *xpc.Connection
	device ios.DeviceEntry
}

// New connects to the pasteboard service on the device. Requires a running
// tunnel (iOS 17+).
func New(device ios.DeviceEntry) (*Connection, error) {
	xpcConn, err := ios.ConnectToXpcServiceTunnelIface(device, ServiceName)
	if err != nil {
		return nil, fmt.Errorf("New: failed to connect to pasteboard service: %w", err)
	}
	return &Connection{conn: xpcConn, device: device}, nil
}

// Close closes the connection to the pasteboard service.
func (c *Connection) Close() error {
	return c.conn.Close()
}

// Item is a single pasteboard item: a set of type identifiers and the raw bytes
// published under each of them.
type Item struct {
	Types []string
	Data  map[string][]byte
}

// TextItem builds a pasteboard item that publishes text as UTF-8 bytes under
// the standard plain-text UTIs.
func TextItem(text string) Item {
	payload := []byte(text)
	data := make(map[string][]byte, len(textUTIs))
	for _, uti := range textUTIs {
		data[uti] = payload
	}
	return Item{Types: textUTIs, Data: data}
}

// DataItem builds a pasteboard item carrying raw bytes under a single UTI.
func DataItem(uti string, data []byte) Item {
	return Item{Types: []string{uti}, Data: map[string][]byte{uti: data}}
}

// SetText replaces the general pasteboard contents with a single UTF-8 text
// value.
func (c *Connection) SetText(text string) error {
	return c.Set([]Item{TextItem(text)})
}

// Set replaces the general pasteboard contents with the given items.
func (c *Connection) Set(items []Item) error {
	wireItems := make([]interface{}, len(items))
	for i, item := range items {
		data := make(map[string]interface{}, len(item.Data))
		for uti, bytes := range item.Data {
			data[uti] = map[string]interface{}{"data": bytes}
		}
		types := make([]interface{}, len(item.Types))
		for j, t := range item.Types {
			types[j] = t
		}
		wireItems[i] = map[string]interface{}{
			"types": types,
			"data":  data,
		}
	}

	request := map[string]interface{}{
		"command":        commandSet,
		"pasteboardName": GeneralPasteboard,
		"items":          wireItems,
		"sourceMetadata": nil,
	}

	golog.Debug("setting pasteboard", "module", logModule, "udid", c.device.Properties.SerialNumber, "items", len(items))

	// The reply is a SET_REPLY snapshot which we don't need, but we asked for a
	// reply via the wanting-reply flag, so read it to keep the channel in sync.
	if _, err := c.sendReceive(request); err != nil {
		return fmt.Errorf("Set: %w", err)
	}
	return nil
}

// GetText pulls the general pasteboard and returns its UTF-8 text. The bool is
// false when the pasteboard holds no decodable text.
func (c *Connection) GetText() (string, bool, error) {
	reply, err := c.Get()
	if err != nil {
		return "", false, err
	}
	text, ok := snapshotText(reply)
	return text, ok, nil
}

// Get pulls the current general pasteboard contents and returns the raw reply
// dict. Use GetText when only text is needed.
func (c *Connection) Get() (map[string]interface{}, error) {
	request := map[string]interface{}{
		"command":        commandPull,
		"pasteboardName": GeneralPasteboard,
		// allResolved => reply carries item data inline, no promise chasing.
		"dataPolicy": map[string]interface{}{"allResolved": map[string]interface{}{}},
	}

	golog.Debug("pulling pasteboard", "module", logModule, "udid", c.device.Properties.SerialNumber)

	reply, err := c.sendReceive(request)
	if err != nil {
		return nil, fmt.Errorf("Get: %w", err)
	}
	return reply, nil
}

// sendReceive sends a request expecting a reply and waits for it, bounded by
// requestTimeout. A wedged pasteboard daemon leaves the channel half-open and
// the read never returns; the timeout turns that into a clear error instead of
// an indefinite hang. The pending read is unblocked when the connection is
// closed.
func (c *Connection) sendReceive(request map[string]interface{}) (map[string]interface{}, error) {
	if err := c.conn.Send(request, xpc.HeartbeatRequestFlag); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	type result struct {
		reply map[string]interface{}
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		// The device interleaves empty control/heartbeat frames (e.g. the
		// wanting-reply ack) with the real reply. Reading a single frame can
		// return one of those empties and leave the real reply queued on the
		// stream, where the next request would pick it up as a stale read. Skip
		// empty-bodied frames and return the first one carrying a populated dict.
		for {
			reply, err := c.conn.ReceiveOnServerClientStream()
			if err != nil {
				ch <- result{nil, err}
				return
			}
			if len(reply) == 0 {
				continue
			}
			ch <- result{reply, nil}
			return
		}
	}()

	select {
	case r := <-ch:
		if r.err != nil {
			return nil, fmt.Errorf("failed to receive reply: %w", r.err)
		}
		return r.reply, nil
	case <-time.After(requestTimeout):
		return nil, fmt.Errorf("timed out after %s waiting for the pasteboard service; it may be wedged — reboot the device and try again", requestTimeout)
	}
}

// snapshotText extracts UTF-8 text from a pull reply. In the reply the items
// live under the "pasteboard" key (in a set request they are top-level). Walks
// items in order and returns the first one carrying a text UTI with inline
// data.
func snapshotText(reply map[string]interface{}) (string, bool) {
	snapshot := reply
	if pb, ok := reply["pasteboard"].(map[string]interface{}); ok {
		snapshot = pb
	}
	items, ok := snapshot["items"].([]interface{})
	if !ok {
		return "", false
	}
	for _, raw := range items {
		item, ok := raw.(map[string]interface{})
		if !ok {
			continue
		}
		dataMap, ok := item["data"].(map[string]interface{})
		if !ok {
			continue
		}
		for _, uti := range textUTIs {
			datum, ok := dataMap[uti].(map[string]interface{})
			if !ok {
				continue
			}
			// Promised items carry {isPromised:true,...} instead of {data:...}.
			if bytes, ok := datum["data"].([]byte); ok && len(bytes) > 0 {
				return string(bytes), true
			}
		}
	}
	return "", false
}
