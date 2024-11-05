package dtx

import (
	"bufio"
	"errors"
	"io"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"

	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

type MethodWithResponse func(msg Message) (interface{}, error)

var ErrConnectionClosed = errors.New("Connection closed")

// Connection manages channels, including the GlobalChannel, for a DtxConnection and dispatches received messages
// to the right channel.
type Connection struct {
	deviceConnection       ios.DeviceConnectionInterface
	channelCodeCounter     int
	activeChannels         sync.Map
	globalChannel          *Channel
	capabilities           map[string]interface{}
	mutex                  sync.Mutex
	requestChannelMessages chan Message

	// MessageDispatcher use this prop to catch messages from GlobalDispatcher
	// and handle it accordingly in a custom dispatcher of the dedicated service
	//
	// Set this prop when creating a connection instance
	//
	// Refer to end-to-end example of `instruments/instruments_sysmontap.go`
	MessageDispatcher Dispatcher

	closed    chan struct{}
	err       error
	closeOnce sync.Once
}

// Dispatcher is a simple interface containing a Dispatch func to receive dtx.Messages
type Dispatcher interface {
	Dispatch(msg Message)
}

// GlobalDispatcher the message dispatcher for the automatically created global Channel
type GlobalDispatcher struct {
	dispatchFunctions      map[string]func(Message)
	requestChannelMessages chan Message
	dtxConnection          *Connection
}

const requestChannel = "_requestChannelWithCode:identifier:"

// Closed is closed when the underlying DTX connection was closed for any reason (either initiated by calling Close() or due to an error)
func (dtxConn *Connection) Closed() <-chan struct{} {
	return dtxConn.closed
}

// Err is non-nil when the connection was closed (when Close was called this will be ErrConnectionClosed)
func (dtxConn *Connection) Err() error {
	return dtxConn.err
}

// Close closes the underlying deviceConnection
func (dtxConn *Connection) Close() error {
	if dtxConn.deviceConnection != nil {
		err := dtxConn.deviceConnection.Close()
		dtxConn.close(err)
		return err
	}
	dtxConn.close(ErrConnectionClosed)
	return nil
}

// GlobalChannel returns the connections automatically created global channel.
func (dtxConn *Connection) GlobalChannel() *Channel {
	return dtxConn.globalChannel
}

// NewGlobalDispatcher create a Dispatcher for the GlobalChannel
func NewGlobalDispatcher(requestChannelMessages chan Message, dtxConnection *Connection) Dispatcher {
	dispatcher := GlobalDispatcher{
		dispatchFunctions:      map[string]func(Message){},
		requestChannelMessages: requestChannelMessages,
		dtxConnection:          dtxConnection,
	}
	const notifyPublishedCaps = "_notifyOfPublishedCapabilities:"
	dispatcher.dispatchFunctions[notifyPublishedCaps] = notifyOfPublishedCapabilities
	return dispatcher
}

// Dispatch to a MessageDispatcher of the Connection if set
func (dtxConn *Connection) Dispatch(msg Message) {
	msgDispatcher := dtxConn.MessageDispatcher
	if msgDispatcher != nil {
		log.Debugf("msg dispatcher found: %T", msgDispatcher)
		msgDispatcher.Dispatch(msg)
		return
	}

	log.Errorf("no connection dispatcher registered for global channel, msg: %v", msg)
}

// Dispatch prints log messages and errors when they are received and also creates local Channels when requested by the device.
func (g GlobalDispatcher) Dispatch(msg Message) {
	SendAckIfNeeded(g.dtxConnection, msg)
	if msg.Payload != nil {
		if requestChannel == msg.Payload[0] {
			g.requestChannelMessages <- msg
		}
		// TODO: use the dispatchFunctions map
		if "outputReceived:fromProcess:atTime:" == msg.Payload[0] {
			logmsg, err := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
			if err == nil {
				log.WithFields(log.Fields{
					"msg":  logmsg[0],
					"pid":  msg.Auxiliary.GetArguments()[1],
					"time": msg.Auxiliary.GetArguments()[2],
				}).Info("outputReceived:fromProcess:atTime:")
			}
			return
		}
	}
	log.Tracef("Global Dispatcher Received: %s %s", msg.Payload, msg.Auxiliary)
	if msg.HasError() {
		log.Error(msg.Payload[0])
	}
	if msg.PayloadHeader.MessageType == UnknownTypeOne {
		g.dtxConnection.Dispatch(msg)
	}
}

func notifyOfPublishedCapabilities(msg Message) {
	log.Debug("capabs received")
}

// NewUsbmuxdConnection connects and starts reading from a Dtx based service on the device
func NewUsbmuxdConnection(device ios.DeviceEntry, serviceName string) (*Connection, error) {
	conn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return nil, err
	}

	return newDtxConnection(conn)
}

// NewTunnelConnection connects and starts reading from a Dtx based service on the device, using tunnel interface instead of usbmuxd
func NewTunnelConnection(device ios.DeviceEntry, serviceName string) (*Connection, error) {
	conn, err := ios.ConnectToServiceTunnelIface(device, serviceName)
	if err != nil {
		return nil, err
	}

	return newDtxConnection(conn)
}

func newDtxConnection(conn ios.DeviceConnectionInterface) (*Connection, error) {
	requestChannelMessages := make(chan Message, 5)

	// The global channel has channelCode 0, so we need to start with channelCodeCounter==1
	dtxConnection := &Connection{deviceConnection: conn, channelCodeCounter: 1, requestChannelMessages: requestChannelMessages}
	dtxConnection.closed = make(chan struct{})

	// The global channel is automatically present and used for requesting other channels and some other methods like notifyPublishedCapabilities
	globalChannel := Channel{
		channelCode:       0,
		messageIdentifier: 5, channelName: "global_channel", connection: dtxConnection,
		messageDispatcher: NewGlobalDispatcher(requestChannelMessages, dtxConnection),
		responseWaiters:   map[int]chan Message{},
		registeredMethods: map[string]chan Message{},
		defragmenters:     map[int]*FragmentDecoder{},
		timeout:           5 * time.Second,
	}
	dtxConnection.globalChannel = &globalChannel
	go reader(dtxConnection)

	return dtxConnection, nil
}

// Send sends the byte slice directly to the device using the underlying DeviceConnectionInterface
func (dtxConn *Connection) Send(message []byte) error {
	return dtxConn.deviceConnection.Send(message)
}

// reader reads messages from the byte stream and dispatches them to the right channel when they are decoded.
func reader(dtxConn *Connection) {
	reader := bufio.NewReader(dtxConn.deviceConnection.Reader())
	for {
		msg, err := ReadMessage(reader)
		if err != nil {
			defer dtxConn.close(err)
			errText := err.Error()
			if err == io.EOF || strings.Contains(errText, "use of closed network") {
				log.Debug("DTX Connection with EOF")
				return
			}
			log.Errorf("error reading dtx connection %+v", err)
			return
		}
		if _channel, ok := dtxConn.activeChannels.Load(msg.ChannelCode); ok {
			channel := _channel.(*Channel)
			channel.Dispatch(msg)
		} else {
			dtxConn.globalChannel.Dispatch(msg)
		}
	}
}

func SendAckIfNeeded(dtxConn *Connection, msg Message) {
	if msg.ExpectsReply {
		ack := BuildAckMessage(msg)
		err := dtxConn.Send(ack)
		if err != nil {
			log.Errorf("Error sending ack:%s", err)
		}
	}
}

func (dtxConn *Connection) ForChannelRequest(messageDispatcher Dispatcher) *Channel {
	msg := <-dtxConn.requestChannelMessages
	dtxConn.mutex.Lock()
	defer dtxConn.mutex.Unlock()
	// code := msg.Auxiliary.GetArguments()[0].(uint32)
	identifier, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[1].([]byte))
	// TODO: Setting the channel code here manually to -1 for making testmanagerd work. For some reason it requests the TestDriver proxy channel with code 1 but sends messages on -1. Should probably be fixed somehow
	// TODO: try to refactor testmanagerd/xcuitest code and use AddDefaultChannelReceiver instead of this function. The only code calling this is in testmanagerd right now.
	channel := &Channel{channelCode: -1, channelName: identifier[0].(string), messageIdentifier: 1, connection: dtxConn, messageDispatcher: messageDispatcher, responseWaiters: map[int]chan Message{}, defragmenters: map[int]*FragmentDecoder{}, timeout: 5 * time.Second}
	dtxConn.activeChannels.Store(-1, channel)
	return channel
}

// AddDefaultChannelReceiver let's you set the Dispatcher for the Channel with code -1 ( or 4294967295 for uint32).
// I am just calling it the "default" channel now, without actually figuring out what it is for exactly from disassembled code.
// If someone wants to do that and bring some clarity, please go ahead :-)
// This channel seems to always be there without explicitly requesting it and sometimes it is used.
func (dtxConn *Connection) AddDefaultChannelReceiver(messageDispatcher Dispatcher) *Channel {
	channel := &Channel{channelCode: -1, channelName: "c -1/ 4294967295 receiver channel ", messageIdentifier: 1, connection: dtxConn, messageDispatcher: messageDispatcher, responseWaiters: map[int]chan Message{}, defragmenters: map[int]*FragmentDecoder{}, timeout: 5 * time.Second}
	dtxConn.activeChannels.Store(uint32(math.MaxUint32), channel)
	return channel
}

// RequestChannelIdentifier requests a channel to be opened on the Connection with the given identifier,
// an automatically assigned channelCode and a Dispatcher for receiving messages.
func (dtxConn *Connection) RequestChannelIdentifier(identifier string, messageDispatcher Dispatcher, opts ...ChannelOption) *Channel {
	dtxConn.mutex.Lock()
	defer dtxConn.mutex.Unlock()
	code := dtxConn.channelCodeCounter
	dtxConn.channelCodeCounter++

	payload, _ := nskeyedarchiver.ArchiveBin(requestChannel)
	auxiliary := NewPrimitiveDictionary()
	auxiliary.AddInt32(code)
	arch, _ := nskeyedarchiver.ArchiveBin(identifier)
	auxiliary.AddBytes(arch)
	log.WithFields(log.Fields{"channel_id": identifier}).Debug("Requesting channel")

	rply, err := dtxConn.globalChannel.SendAndAwaitReply(true, Methodinvocation, payload, auxiliary)
	log.Debug(rply)
	if err != nil {
		log.WithFields(log.Fields{"channel_id": identifier, "error": err}).Error("failed requesting channel")
	}
	log.WithFields(log.Fields{"channel_id": identifier}).Debug("Channel open")
	channel := &Channel{channelCode: code, channelName: identifier, messageIdentifier: 1, connection: dtxConn, messageDispatcher: messageDispatcher, responseWaiters: map[int]chan Message{}, defragmenters: map[int]*FragmentDecoder{}, timeout: 5 * time.Second}
	dtxConn.activeChannels.Store(code, channel)
	for _, opt := range opts {
		opt(channel)
	}

	return channel
}

func (dtxConn *Connection) close(err error) {
	dtxConn.closeOnce.Do(func() {
		dtxConn.err = err
		close(dtxConn.closed)
	})
}
