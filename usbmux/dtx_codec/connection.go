package dtx

import (
	"sync"

	"github.com/danielpaulus/go-ios/usbmux"
	"github.com/danielpaulus/go-ios/usbmux/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

type DtxConnection struct {
	dtxConnection          usbmux.DeviceConnectionInterface
	channelCodeCounter     int
	activeChannels         map[int]*DtxChannel
	globalChannel          *DtxChannel
	capabilities           map[string]interface{}
	mutex                  sync.Mutex
	requestChannelMessages chan Message
}

type GlobalDispatcher struct {
	dispatchFunctions      map[string]func(Message)
	requestChannelMessages chan Message
}
type DtxDispatcher interface {
	Dispatch(msg Message)
}

const requestChannel = "_requestChannelWithCode:identifier:"

func (d *DtxConnection) Close() {
	d.dtxConnection.Close()
}

func (d *DtxConnection) GlobalChannel() *DtxChannel {
	return d.globalChannel
}

func NewGlobalDispatcher(requestChannelMessages chan Message) DtxDispatcher {
	dispatcher := GlobalDispatcher{dispatchFunctions: map[string]func(Message){},
		requestChannelMessages: requestChannelMessages}
	const notifyPublishedCaps = "_notifyOfPublishedCapabilities:"
	dispatcher.dispatchFunctions[notifyPublishedCaps] = notifyOfPublishedCapabilities
	return dispatcher
}
func (g GlobalDispatcher) Dispatch(msg Message) {
	if msg.Payload != nil {
		if requestChannel == msg.Payload[0] {
			g.requestChannelMessages <- msg
		}
		if "outputReceived:fromProcess:atTime:" == msg.Payload[0] {
			msg, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
			log.Info(msg[0])
			return
		}
	}
	log.Infof("Global Dispatcher Received: %s %s %s", msg.Payload[0], msg, msg.Auxiliary)
	if msg.HasError() {
		log.Error(msg.Payload[0])
	}
}

func notifyOfPublishedCapabilities(msg Message) {
	log.Info("capabs")
}

func NewDtxConnection(deviceId int, udid string, serviceName string) (*DtxConnection, error) {
	conn, err := usbmux.ConnectToService(deviceId, udid, serviceName)
	if err != nil {
		return nil, err
	}
	requestChannelMessages := make(chan Message, 5)
	dtxConnection := &DtxConnection{dtxConnection: conn, channelCodeCounter: 1, activeChannels: map[int]*DtxChannel{}, requestChannelMessages: requestChannelMessages}
	globalChannel := DtxChannel{channelCode: 0,
		messageIdentifier: 5, channelName: "global_channel", connection: dtxConnection,
		messageDispatcher: NewGlobalDispatcher(requestChannelMessages), responseWaiters: map[int]chan Message{}, registeredMethods: map[string]chan Message{}}
	dtxConnection.globalChannel = &globalChannel
	go reader(dtxConnection)

	return dtxConnection, nil
}

func (d *DtxConnection) Send(message []byte) error {
	return d.dtxConnection.Send(message)
}

func reader(dtxConn *DtxConnection) {
	for {
		reader := dtxConn.dtxConnection.Reader()
		msg, err := ReadMessage(reader)
		if err != nil {
			log.Fatal(err)
		}
		sendAckIfNeeded(dtxConn, msg)
		if channel, ok := dtxConn.activeChannels[msg.ChannelCode]; ok {
			channel.Dispatch(msg)
		} else {
			dtxConn.globalChannel.Dispatch(msg)
		}
	}
}

func sendAckIfNeeded(dtxConn *DtxConnection, msg Message) {
	if msg.ExpectsReply {
		err := dtxConn.Send(BuildAckMessage(msg))
		if err != nil {
			log.Fatalf("Error sending ack:%s", err)
		}
	}
}

func (d *DtxConnection) ForChannelRequest(messageDispatcher DtxDispatcher) *DtxChannel {
	msg := <-d.requestChannelMessages
	d.mutex.Lock()
	defer d.mutex.Unlock()
	code := msg.Auxiliary.GetArguments()[0].(uint32)
	identifier, _ := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[1].([]byte))
	channel := &DtxChannel{channelCode: -1, channelName: identifier[0].(string), messageIdentifier: 1, connection: d, messageDispatcher: messageDispatcher, responseWaiters: map[int]chan Message{}}
	d.activeChannels[int(code)] = channel
	return channel
}

func (d *DtxConnection) RequestChannelIdentifier(identifier string, messageDispatcher DtxDispatcher) *DtxChannel {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	code := d.channelCodeCounter
	d.channelCodeCounter++

	payload, _ := nskeyedarchiver.ArchiveBin(requestChannel)
	auxiliary := NewPrimitiveDictionary()
	auxiliary.AddInt32(code)
	arch, _ := nskeyedarchiver.ArchiveBin(identifier)
	auxiliary.AddBytes(arch)
	log.WithFields(log.Fields{"channel_id": identifier}).Debug("Requesting channel")

	rply, err := d.globalChannel.SendAndAwaitReply(true, Methodinvocation, payload, auxiliary)
	log.Debug(rply)
	if err != nil {
		log.WithFields(log.Fields{"channel_id": identifier, "error": err}).Error("failed requesting channel")
	}
	log.WithFields(log.Fields{"channel_id": identifier}).Debug("Channel open")
	channel := &DtxChannel{channelCode: code, channelName: identifier, messageIdentifier: 1, connection: d, messageDispatcher: messageDispatcher, responseWaiters: map[int]chan Message{}}
	d.activeChannels[code] = channel
	return channel
}
