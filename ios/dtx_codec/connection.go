package dtx

import (
	"io"
	"sync"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

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
func NewGlobalDispatcher(requestChannelMessages chan Message) Dispatcher {
	dispatcher := GlobalDispatcher{dispatchFunctions: map[string]func(Message){},
		requestChannelMessages: requestChannelMessages}
	const notifyPublishedCaps = "_notifyOfPublishedCapabilities:"
	dispatcher.dispatchFunctions[notifyPublishedCaps] = notifyOfPublishedCapabilities
	return dispatcher
}

//Dispatch prints log messages and errors when they are received and also creates local Channels when requested by the device.
func (g GlobalDispatcher) Dispatch(msg Message) {
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
	log.Debugf("Global Dispatcher Received: %s %s %s", msg.Payload[0], msg, msg.Auxiliary)
	if msg.HasError() {
		log.Error(msg.Payload[0])
	}
}

func notifyOfPublishedCapabilities(msg Message) {
	log.Info("capabs")
}

//NewConnection connects and starts reading from a Dtx based service on the device
func NewConnection(device ios.DeviceEntry, serviceName string) (*Connection, error) {
	conn, err := ios.ConnectToService(device.DeviceID, device.Properties.SerialNumber, serviceName)
	if err != nil {
		return nil, err
	}
	requestChannelMessages := make(chan Message, 5)

	//The global channel has channelCode 0, so we need to start with channelCodeCounter==1
	dtxConnection := &Connection{deviceConnection: conn, channelCodeCounter: 1, activeChannels: map[int]*Channel{}, requestChannelMessages: requestChannelMessages}

	//The global channel is automatically present and used for requesting other channels and some other methods like notifyPublishedCapabilities
	globalChannel := Channel{channelCode: 0,
		messageIdentifier: 5, channelName: "global_channel", connection: dtxConnection,
		messageDispatcher: NewGlobalDispatcher(requestChannelMessages), responseWaiters: map[int]chan Message{}, registeredMethods: map[string]chan Message{}}
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
			if err == io.EOF {
				log.Debug("Closing DTX Connection")
				return
			}
			log.Fatal(err)
		}
		//TODO: move this to the channel level, the connection probably should not auto ack messages
		sendAckIfNeeded(dtxConn, msg)
		if channel, ok := dtxConn.activeChannels[msg.ChannelCode]; ok {
			channel.Dispatch(msg)
		} else {
			dtxConn.globalChannel.Dispatch(msg)
		}
	}
}

func sendAckIfNeeded(dtxConn *Connection, msg Message) {
	if msg.ExpectsReply {
		err := dtxConn.Send(BuildAckMessage(msg))
		if err != nil {
			log.Fatalf("Error sending ack:%s", err)
		}
	}
}

func (dtxConn *Connection) ForChannelRequest(messageDispatcher Dispatcher) *Channel {
	msg := <-dtxConn.requestChannelMessages
	dtxConn.mutex.Lock()
	defer dtxConn.mutex.Unlock()
	code := msg.Auxiliary.GetArguments()[0].(uint32)
	identifier, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[1].([]byte))
	//TODO: Setting the channel code here manually to -1 for making testmanagerd work. For some reason it requests the TestDriver proxy channel with code 1 but sends messages on -1. Should probably be fixed somehow
	channel := &Channel{channelCode: -1, channelName: identifier[0].(string), messageIdentifier: 1, connection: dtxConn, messageDispatcher: messageDispatcher, responseWaiters: map[int]chan Message{}}
	dtxConn.activeChannels[int(code)] = channel
	return channel
}

//RequestChannelIdentifier requests a channel to be opened on the Connection with the given identifier,
//an automatically assigned channelCode and a Dispatcher for receiving messages.
func (dtxConn *Connection) RequestChannelIdentifier(identifier string, messageDispatcher Dispatcher) *Channel {
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
	channel := &Channel{channelCode: code, channelName: identifier, messageIdentifier: 1, connection: dtxConn, messageDispatcher: messageDispatcher, responseWaiters: map[int]chan Message{}}
	dtxConn.activeChannels[code] = channel
	return channel
}
