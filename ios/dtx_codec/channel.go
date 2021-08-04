package dtx

import (
	"fmt"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

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

//ChannelOption for configuring settings on dtx.Channels
type ChannelOption func(*Channel)

//WithTimeout adds a custom timeout in seconds to the channel.
//Some longer running synchronous operations need that.
func WithTimeout(seconds uint32) ChannelOption {
	return func(h *Channel) {
		h.timeout = time.Duration(seconds) * time.Second
	}
}

func (d *Channel) RegisterMethodForRemote(selector string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.registeredMethods[selector] = make(chan Message)
}

func (d *Channel) ReceiveMethodCall(selector string) Message {
	d.mutex.Lock()
	channel := d.registeredMethods[selector]
	d.mutex.Unlock()
	return <-channel
}

//MethodCall is the standard DTX style remote method invocation pattern. The ObjectiveC Selector goes as a NSKeyedArchiver.archived NSString into the
//DTXMessage payload, and the arguments are separately NSKeyArchiver.archived and put into the Auxiliary DTXPrimitiveDictionary. It returns the response message and an error.
func (d *Channel) MethodCall(selector string, args ...interface{}) (Message, error) {
	payload, _ := nskeyedarchiver.ArchiveBin(selector)
	auxiliary := NewPrimitiveDictionary()
	for _, arg := range args {
		auxiliary.AddNsKeyedArchivedObject(arg)
	}
	msg, err := d.SendAndAwaitReply(true, Methodinvocation, payload, auxiliary)
	if err != nil {
		log.WithFields(log.Fields{"channel_id": d.channelName, "error": err, "methodselector": selector}).Info("failed starting invoking method")
	}
	if msg.HasError() {
		return msg, fmt.Errorf("Failed invoking method '%s' with error: %s", selector, msg.Payload[0])
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
		log.WithFields(log.Fields{"channel_id": d.channelName, "error": err, "methodselector": selector}).Info("failed starting invoking method")
	}
	return nil
}

func (d *Channel) Send(expectsReply bool, messageType int, payloadBytes []byte, auxiliary PrimitiveDictionary) error {
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

func (d *Channel) SendAndAwaitReply(expectsReply bool, messageType int, payloadBytes []byte, auxiliary PrimitiveDictionary) (Message, error) {
	d.mutex.Lock()
	identifier := d.messageIdentifier
	d.messageIdentifier++
	d.mutex.Unlock()
	bytes, err := Encode(identifier, 0, d.channelCode, expectsReply, messageType, payloadBytes, auxiliary)
	if err != nil {
		return Message{}, err
	}
	responseChannel := make(chan Message)
	d.AddResponseWaiter(identifier, responseChannel)

	err = d.connection.Send(bytes)
	if err != nil {
		return Message{}, err
	}
	select {
	case response := <-responseChannel:
		return response, nil
	case <-time.After(d.timeout):
		return Message{}, fmt.Errorf("Timed out waiting for response for message:%d channel:%d", identifier, d.channelCode)
	}

}

func (d *Channel) Dispatch(msg Message) {
	d.mutex.Lock()
	if msg.Identifier >= d.messageIdentifier {
		d.messageIdentifier = msg.Identifier + 1
	}
	if msg.PayloadHeader.MessageType == Methodinvocation {
		log.Debug("Dispatching:", msg.Payload[0].(string))
		if v, ok := d.registeredMethods[msg.Payload[0].(string)]; ok {
			d.mutex.Unlock()
			v <- msg
			return
		}
	}
	d.mutex.Unlock()
	if msg.ConversationIndex > 0 {
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
						log.Error("Decoding fragmented message failed")
					}
					if err != nil {
						log.Error("decoding framente")
					}
					d.responseWaiters[msg.Identifier] <- msg
					delete(d.responseWaiters, msg.Identifier)
				}
				return
			}
			log.Warn("received message fragment without first message, dropping it")
			delete(d.responseWaiters, msg.Identifier)
			return
		}

		d.responseWaiters[msg.Identifier] <- msg
		delete(d.responseWaiters, msg.Identifier)
		return
	}
	d.messageDispatcher.Dispatch(msg)
}
