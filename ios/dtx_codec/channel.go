package dtx

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
)

const logModule = "go-ios/dtx_codec"

type Channel struct {
	channelCode       int
	channelName       string
	messageIdentifier int
	connection        *Connection
	messageDispatcher Dispatcher
	responseWaiters   map[int]chan Message
	defragmenters     map[int]*FragmentDecoder
	registeredMethods map[string]chan Message
	mutex             sync.Mutex
	timeout           time.Duration
}

// ChannelOption for configuring settings on dtx.Channels
type ChannelOption func(*Channel)

// WithTimeout adds a custom timeout in seconds to the channel.
// Some longer running synchronous operations need that.
func WithTimeout(seconds uint32) ChannelOption {
	return func(h *Channel) {
		h.timeout = time.Duration(seconds) * time.Second
	}
}

// registeredMethodBuffer bounds how many not-yet-consumed remote method calls
// are kept per selector. Dispatch runs on the connection's single reader
// goroutine and must never block: without a buffer, one event arriving while
// no ReceiveMethodCall is pending would stall delivery for every channel on
// the connection (head-of-line blocking with a silent hang).
const registeredMethodBuffer = 16

func (d *Channel) RegisterMethodForRemote(selector string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	if d.registeredMethods == nil {
		d.registeredMethods = map[string]chan Message{}
	}
	d.registeredMethods[selector] = make(chan Message, registeredMethodBuffer)
}

func (d *Channel) ReceiveMethodCall(selector string) Message {
	d.mutex.Lock()
	channel := d.registeredMethods[selector]
	d.mutex.Unlock()
	return <-channel
}

func (d *Channel) ReceiveMethodCallWithTimeout(ctx context.Context, selector string) (Message, error) {
	d.mutex.Lock()
	channel := d.registeredMethods[selector]
	d.mutex.Unlock()
	select {
	case msg := <-channel:
		return msg, nil
	// context is cancelled because the timeout is exceeded
	case <-ctx.Done():
		return Message{}, ctx.Err()
	}
}

// MethodCall is the standard DTX style remote method invocation pattern. The ObjectiveC Selector goes as a NSKeyedArchiver.archived NSString into the
// DTXMessage payload, and the arguments are separately NSKeyArchiver.archived and put into the Auxiliary DTXPrimitiveDictionary. It returns the response message and an error.
// Always uses the channel's default timeout.
func (d *Channel) MethodCall(selector string, args ...interface{}) (Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	return d.MethodCallWithContext(ctx, selector, args...)
}

// MethodCallWithContext is like MethodCall but respects the provided context for cancellation/timeout.
// If the context has no deadline, the channel's default timeout is applied.
func (d *Channel) MethodCallWithContext(ctx context.Context, selector string, args ...interface{}) (Message, error) {
	if _, hasDeadline := ctx.Deadline(); !hasDeadline {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, d.timeout)
		defer cancel()
	}

	auxiliary := NewPrimitiveDictionary()
	for _, arg := range args {
		auxiliary.AddNsKeyedArchivedObject(arg)
	}

	return d.methodCallWithReply(ctx, selector, auxiliary)
}

// MethodCallWithAuxiliary is a DTX style remote method invocation pattern. The ObjectiveC Selector goes as a NSKeyedArchiver.archived NSString into the
// DTXMessage payload, and the primitive arguments put into the Auxiliary DTXPrimitiveDictionary. It returns the response message and an error.
func (d *Channel) MethodCallWithAuxiliary(selector string, aux PrimitiveDictionary) (Message, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	return d.methodCallWithReply(ctx, selector, aux)
}

func (d *Channel) methodCallWithReply(ctx context.Context, selector string, auxiliary PrimitiveDictionary) (Message, error) {
	payload, _ := nskeyedarchiver.ArchiveBin(selector)
	msg, err := d.sendAndAwaitReply(ctx, true, Methodinvocation, payload, auxiliary)
	if err != nil {
		golog.Info("failed starting invoking method", "module", logModule, "channel_id", d.channelName, "error", err, "methodselector", selector)
		return msg, err
	}
	if msg.HasError() {
		var errPayload interface{}
		if len(msg.Payload) > 0 {
			errPayload = msg.Payload[0]
		}
		return msg, fmt.Errorf("failed invoking method '%s' with error: %v", selector, errPayload)
	}
	return msg, nil
}

func (d *Channel) MethodCallAsync(selector string, args ...interface{}) error {
	payload, _ := nskeyedarchiver.ArchiveBin(selector)
	auxiliary := NewPrimitiveDictionary()
	for _, arg := range args {
		auxiliary.AddNsKeyedArchivedObject(arg)
	}
	err := d.Send(false, Methodinvocation, payload, auxiliary)
	if err != nil {
		golog.Info("failed starting invoking method", "module", logModule, "channel_id", d.channelName, "error", err, "methodselector", selector)
		return err
	}
	return nil
}

func (d *Channel) Send(expectsReply bool, messageType MessageType, payloadBytes []byte, auxiliary PrimitiveDictionary) error {
	d.mutex.Lock()

	identifier := d.messageIdentifier
	d.messageIdentifier++
	d.mutex.Unlock()

	bytes, err := Encode(identifier, 0, d.channelCode, expectsReply, messageType, payloadBytes, auxiliary)
	if err != nil {
		return err
	}
	return d.connection.Send(bytes)
}

func (d *Channel) AddResponseWaiter(identifier int, channel chan Message) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.responseWaiters[identifier] = channel
}

func (d *Channel) sendAndAwaitReply(ctx context.Context, expectsReply bool, messageType MessageType, payloadBytes []byte, auxiliary PrimitiveDictionary) (Message, error) {
	d.mutex.Lock()
	identifier := d.messageIdentifier
	d.messageIdentifier++
	d.mutex.Unlock()
	bytes, err := Encode(identifier, 0, d.channelCode, expectsReply, messageType, payloadBytes, auxiliary)
	if err != nil {
		return Message{}, err
	}
	// Buffered so a reply racing with our timeout can always be deposited by
	// sendResponse instead of being dropped on the non-blocking send.
	responseChannel := make(chan Message, 1)
	d.AddResponseWaiter(identifier, responseChannel)

	err = d.connection.Send(bytes)
	if err != nil {
		d.removeResponseWaiter(identifier)
		return Message{}, err
	}
	select {
	case response := <-responseChannel:
		return response, nil
	case <-ctx.Done():
		// Deregister so the abandoned waiter does not leak in responseWaiters
		// (and a stale defragmenter for this identifier is cleaned up too).
		d.removeResponseWaiter(identifier)
		return Message{}, fmt.Errorf("Timed out waiting for response for message:%d channel:%d", identifier, d.channelCode)
	}
}

func (d *Channel) removeResponseWaiter(identifier int) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	delete(d.responseWaiters, identifier)
	delete(d.defragmenters, identifier)
}

// deliverEvent hands a device-initiated method call to the selector's queue
// without ever blocking the reader goroutine. The queue is LOSSY BY DESIGN:
// when it is full the OLDEST event is dropped (with a warning) so the most
// recent state survives. This is the right trade-off for the current
// latest-state consumers (e.g. the AX inspector's cursor position); do not
// route a selector through here if every single callback matters.
func (d *Channel) deliverEvent(queue chan Message, msg Message, selector string) {
	for {
		select {
		case queue <- msg:
			return
		default:
		}
		select {
		case dropped := <-queue:
			golog.Warn("event queue full, dropping oldest event", "module", logModule, "channel_id", d.channelName, "selector", selector, "dropped_identifier", dropped.Identifier)
		default:
		}
	}
}

func (d *Channel) Dispatch(msg Message) {
	d.mutex.Lock()
	if msg.Identifier >= d.messageIdentifier {
		d.messageIdentifier = msg.Identifier + 1
	}
	if msg.PayloadHeader.MessageType == Methodinvocation {
		var selector string
		if len(msg.Payload) > 0 {
			selector, _ = msg.Payload[0].(string)
		}
		golog.Trace("dispatching", "module", logModule, "channel_id", d.channelName, "selector", selector)
		if v, ok := d.registeredMethods[selector]; ok {
			d.mutex.Unlock()
			// Never block the reader goroutine: an event arriving while no
			// ReceiveMethodCall is pending would otherwise stall delivery for
			// every channel on this connection.
			d.deliverEvent(v, msg, selector)
			return
		}
	}
	d.mutex.Unlock()
	if msg.ConversationIndex > 0 || msg.IsFragment() {
		d.mutex.Lock()
		defer d.mutex.Unlock()
		if msg.IsFirstFragment() {
			d.defragmenters[msg.Identifier] = NewFragmentDecoder(msg)
			SendAckIfNeeded(d.connection, msg)
			return
		}
		if msg.IsFragment() {
			if defragmenter, ok := d.defragmenters[msg.Identifier]; ok {
				defragmenter.AddFragment(msg)
				if msg.IsLastFragment() {
					messagesBytes := defragmenter.Extract()
					msg, leftover, err := DecodeNonBlocking(messagesBytes)
					if len(leftover) != 0 {
						golog.Error("Decoding fragmented message failed", "module", logModule, "channel_id", d.channelName)
					}
					if err != nil {
						golog.Error("Decoding fragment", "module", logModule, "channel_id", d.channelName)
					}

					if msg.ConversationIndex > 0 {
						sendResponse(d.responseWaiters[msg.Identifier], msg)
					} else {
						d.messageDispatcher.Dispatch(msg)
					}
					delete(d.responseWaiters, msg.Identifier)
					delete(d.defragmenters, msg.Identifier)
				}
				return
			}
			golog.Warn("Received message fragment without first message, dropping it", "module", logModule, "channel_id", d.channelName)
			delete(d.responseWaiters, msg.Identifier)
			delete(d.defragmenters, msg.Identifier)
			return
		}

		sendResponse(d.responseWaiters[msg.Identifier], msg)
		delete(d.responseWaiters, msg.Identifier)
		delete(d.defragmenters, msg.Identifier)
		return
	}
	d.messageDispatcher.Dispatch(msg)
}

// sendResponse delivers msg to a response waiter without ever blocking the
// reader goroutine. A malformed/unexpected Identifier may map to no waiter (a
// nil channel, which would block forever) or to a waiter that has already given
// up, so we guard for nil and use a non-blocking send.
func sendResponse(waiter chan Message, msg Message) {
	if waiter == nil {
		return
	}
	select {
	case waiter <- msg:
	default:
	}
}
