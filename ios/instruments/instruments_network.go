package instruments

import (
	"time"

	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

// type channelDispatcher2 struct {
// 	messageChannel chan dtx.Message
// 	closeChannel   chan struct{}
// }

// func (dispatcher channelDispatcher2) Receive() (map[string]interface{}, error) {
// 	log.Println("--->channelDispatcher2.Receive")
// 	for {
// 		fmt.Println("---1")
// 		select {
// 		case msg := <-dispatcher.messageChannel:
// 			fmt.Println("---2")
// 			selector, result, err := toMap(msg)
// 			if "applicationStateNotification:" == selector && err == nil {
// 				return result, nil
// 			}
// 			if err != nil {
// 				log.Debugf("error extracting message %+v, %v", msg, err)
// 			}
// 		case <-dispatcher.closeChannel:
// 			fmt.Println("---4")
// 			return map[string]interface{}{}, io.EOF
// 		}
// 		fmt.Println("---3")
// 	}
// }

// func (dispatcher *channelDispatcher2) Close() error {
// 	select {
// 	case dispatcher.closeChannel <- struct{}{}:
// 		return nil
// 	case <-time.After(time.Second * 5):
// 		return fmt.Errorf("timeout")
// 	}
// }

// func (dispatcher channelDispatcher2) Dispatch(msg dtx.Message) {
// 	fmt.Println("---2---Dispatch")
// 	dispatcher.messageChannel <- msg
// }

func ListenNetwork(device ios.DeviceEntry) (func() (map[string]interface{}, error), func() error, error) {
	conn, err := connectInstruments(device)
	if err != nil {
		return nil, nil, err
	}
	// dispatcher := channelDispatcher2{messageChannel: make(chan dtx.Message), closeChannel: make(chan struct{})}
	// conn.AddDefaultChannelReceiver(dispatcher)

	// capabs := map[string]interface{}{
	// 	"com.apple.private.DTXBlockCompression": uint64(2),
	// 	"com.apple.private.DTXConnection":       uint64(1),
	// }

	// conn.GlobalChannel().MethodCall("_notifyOfPublishedCapabilities:", capabs)

	channel := conn.RequestChannelIdentifier(mobileNetworkingChannel, channelDispatcher{})

	resp, err := channel.MethodCall("startMonitoring")
	if err != nil {
		log.Errorf("resp:%+v", resp)
		return nil, nil, err
	}
	time.Sleep(time.Duration(5) * time.Second)
	conn.Close()
	// return dispatcher.Receive, dispatcher.Close, nil
	return nil, nil, nil

}
