package dtx

import (
	"sync"

	"github.com/danielpaulus/go-ios/usbmux"
	"github.com/danielpaulus/go-ios/usbmux/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

type DtxConnection struct {
	dtxConnection      usbmux.DeviceConnectionInterface
	channelCodeCounter int
	activeChannels     map[int]DtxChannel
	globalChannel      DtxChannel
	capabilities       map[string]interface{}
	mutex              sync.Mutex
}

type GlobalDispatcher struct {
	dispatchFunctions map[string]func(DtxMessage)
}
type DtxDispatcher interface {
	Dispatch(msg DtxMessage)
}

func NewGlobalDispatcher() DtxDispatcher {
	dispatcher := GlobalDispatcher{dispatchFunctions: map[string]func(DtxMessage){}}
	const notifyPublishedCaps = "_notifyOfPublishedCapabilities:"
	dispatcher.dispatchFunctions[notifyPublishedCaps] = notifyOfPublishedCapabilities
	return dispatcher
}
func (g GlobalDispatcher) Dispatch(msg DtxMessage) {
	log.Info(msg)
	if msg.HasError() {
		log.Error(msg.Payload[0])
	}
}

func notifyOfPublishedCapabilities(msg DtxMessage) {
	log.Info("capabs")
}

func NewDtxConnection(deviceId int, udid string, serviceName string) (*DtxConnection, error) {
	conn, err := usbmux.ConnectToService(deviceId, udid, serviceName)
	if err != nil {
		return nil, err
	}
	dtxConnection := &DtxConnection{dtxConnection: conn, channelCodeCounter: 1, activeChannels: map[int]DtxChannel{}}
	globalChannel := DtxChannel{channelCode: 0, messageIdentifier: 5, channelName: "global_channel", connection: dtxConnection, messageDispatcher: NewGlobalDispatcher()}
	dtxConnection.globalChannel = globalChannel
	go reader(dtxConnection)

	return dtxConnection, nil
}

func (d DtxConnection) Send(message []byte) error {
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

func sendAckIfNeeded(dtxConn *DtxConnection, msg DtxMessage) {
	if msg.ExpectsReply {
		err := dtxConn.Send(BuildAckMessage(msg))
		if err != nil {
			log.Fatalf("Error sending ack:%s", err)
		}
	}
}

func (d *DtxConnection) RequestChannelIdentifier(identifier string, messageDispatcher DtxDispatcher) DtxChannel {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	code := d.channelCodeCounter
	d.channelCodeCounter++
	const requestChannel = "_requestChannelWithCode:identifier:"
	payload, _ := nskeyedarchiver.ArchiveBin(requestChannel)
	auxiliary := NewDtxPrimitiveDictionary()
	auxiliary.AddInt32(code)
	arch, _ := nskeyedarchiver.ArchiveBin(identifier)
	auxiliary.AddBytes(arch)
	d.globalChannel.Send(true, MethodinvocationWithoutExpectedReply, payload, auxiliary)

	channel := DtxChannel{channelCode: code, channelName: identifier, connection: d, messageDispatcher: messageDispatcher}
	d.activeChannels[code] = channel
	return channel
}

type DtxChannel struct {
	channelCode       int
	channelName       string
	messageIdentifier int
	connection        *DtxConnection
	messageDispatcher DtxDispatcher
}

func (d *DtxChannel) Send(expectsReply bool, messageType int, payloadBytes []byte, auxiliary DtxPrimitiveDictionary) error {
	identifier := d.messageIdentifier
	d.messageIdentifier++
	bytes, err := Encode(identifier, d.channelCode, expectsReply, messageType, payloadBytes, auxiliary)
	if err != nil {
		return err
	}
	return d.connection.Send(bytes)
}

func (d DtxChannel) Dispatch(msg DtxMessage) {
	d.messageDispatcher.Dispatch(msg)
}
