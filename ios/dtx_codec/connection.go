package dtx

import (
	"io"
	"strings"
	"sync"
	"time"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

type MethodWithResponse func(msg Message) (interface{}, error)

//Connection manages channels, including the GlobalChannel, for a DtxConnection and dispatches received messages
//to the right channel.
type Connection struct {
	deviceConnection       ios.DeviceConnectionInterface
	channelCodeCounter     int
	activeChannels         map[int]*Channel
	globalChannel          *Channel
	capabilities           map[string]interface{}
	mutex                  sync.Mutex
	requestChannelMessages chan Message
}

//Dispatcher is a simple interface containing a Dispatch func to receive dtx.Messages
type Dispatcher interface {
	Dispatch(msg Message)
}

//GlobalDispatcher the message dispatcher for the automatically created global Channel
type GlobalDispatcher struct {
	dispatchFunctions      map[string]func(Message)
	requestChannelMessages chan Message
	dtxConnection          *Connection
}

const requestChannel = "_requestChannelWithCode:identifier:"

//Close closes the underlying deviceConnection
func (dtxConn *Connection) Close() {
	dtxConn.deviceConnection.Close()
}

//GlobalChannel returns the connections automatically created global channel.
func (dtxConn *Connection) GlobalChannel() *Channel {
	return dtxConn.globalChannel
}

//NewGlobalDispatcher create a Dispatcher for the GlobalChannel
func NewGlobalDispatcher(requestChannelMessages chan Message, dtxConnection *Connection) Dispatcher {
	dispatcher := GlobalDispatcher{dispatchFunctions: map[string]func(Message){},
		requestChannelMessages: requestChannelMessages,
		dtxConnection:          dtxConnection,
	}
	const notifyPublishedCaps = "_notifyOfPublishedCapabilities:"
	dispatcher.dispatchFunctions[notifyPublishedCaps] = notifyOfPublishedCapabilities
	return dispatcher
}

//Dispatch prints log messages and errors when they are received and also creates local Channels when requested by the device.
func (g GlobalDispatcher) Dispatch(msg Message) {
	SendAckIfNeeded(g.dtxConnection, msg)
	if msg.Payload != nil {
		if requestChannel == msg.Payload[0] {
			g.requestChannelMessages <- msg
		}
		//TODO: use the dispatchFunctions map
		if "outputReceived:fromProcess:atTime:" == msg.Payload[0] {
			msg, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
			log.Info(msg[0])
			return
		}
	}
	log.Debugf("Global Dispatcher Received: %s %s", msg.Payload, msg.Auxiliary)
	if msg.HasError() {
		log.Error(msg.Payload[0])
	}
}

func notifyOfPublishedCapabilities(msg Message) {
	log.Debug("capabs received")
}

//NewConnection connects and starts reading from a Dtx based service on the device
func NewConnection(device ios.DeviceEntry, serviceName string) (*Connection, error) {
	conn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return nil, err
	}
	requestChannelMessages := make(chan Message, 5)

	//The global channel has channelCode 0, so we need to start with channelCodeCounter==1
	dtxConnection := &Connection{deviceConnection: conn, channelCodeCounter: 1, activeChannels: map[int]*Channel{}, requestChannelMessages: requestChannelMessages}

	//The global channel is automatically present and used for requesting other channels and some other methods like notifyPublishedCapabilities
	globalChannel := Channel{channelCode: 0,
		messageIdentifier: 5, channelName: "global_channel", connection: dtxConnection,
		messageDispatcher: NewGlobalDispatcher(requestChannelMessages, dtxConnection),
		responseWaiters:   map[int]chan Message{},
		registeredMethods: map[string]chan Message{},
		defragmenters:     map[int]*FragmentDecoder{},
		timeout:           5 * time.Second}
	dtxConnection.globalChannel = &globalChannel
	go reader(dtxConnection)

	return dtxConnection, nil
}

//Send sends the byte slice directly to the device using the underlying DeviceConnectionInterface
func (dtxConn *Connection) Send(message []byte) error {
	return dtxConn.deviceConnection.Send(message)
}

//reader reads messages from the byte stream and dispatches them to the right channel when they are decoded.
func reader(dtxConn *Connection) {
	for {
		reader := dtxConn.deviceConnection.Reader()
		msg, err := ReadMessage(reader)
		if err != nil {
			errText := err.Error()
			if err == io.EOF || strings.Contains(errText, "use of closed network") {
				log.Debug("DTX Connection with EOF")
				return
			}
			log.Errorf("error reading dtx connection %+v", err)
			return
		}

		if channel, ok := dtxConn.activeChannels[msg.ChannelCode]; ok {
			channel.Dispatch(msg)
		} else {
			dtxConn.globalChannel.Dispatch(msg)
		}
	}
}

func SendAckIfNeeded(dtxConn *Connection, msg Message) {
	if msg.ExpectsReply {
		log.Debug("sending ack")
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
	//code := msg.Auxiliary.GetArguments()[0].(uint32)
	identifier, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[1].([]byte))
	//TODO: Setting the channel code here manually to -1 for making testmanagerd work. For some reason it requests the TestDriver proxy channel with code 1 but sends messages on -1. Should probably be fixed somehow
	channel := &Channel{channelCode: -1, channelName: identifier[0].(string), messageIdentifier: 1, connection: dtxConn, messageDispatcher: messageDispatcher, responseWaiters: map[int]chan Message{}, defragmenters: map[int]*FragmentDecoder{}, timeout: 5 * time.Second}
	dtxConn.activeChannels[-1] = channel
	return channel
}

//RequestChannelIdentifier requests a channel to be opened on the Connection with the given identifier,
//an automatically assigned channelCode and a Dispatcher for receiving messages.
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
	dtxConn.activeChannels[code] = channel
	for _, opt := range opts {
		opt(channel)
	}

	return channel
}
