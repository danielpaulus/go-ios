package instruments

import (
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	log "github.com/sirupsen/logrus"
)

type yo struct {
}

func (y yo) Dispatch(msg dtx.Message) {

	log.Infof("%+v", msg.Payload[0])
	if "applicationStateNotification:" == msg.Payload[0].(string) {
		log.Infof("%+v", msg.Auxiliary)
	}
}

func GetMetrics(device ios.DeviceEntry) error {
	conn, err := connectInstruments(device)
	if err != nil {
		return err
	}
	conn.AddMinusOneReceiver(yo{})
	channel := conn.RequestChannelIdentifier(mobileNotificationsChannel, yo{})
	resp, err := channel.MethodCall("setApplicationStateNotificationsEnabled:", true)
	if err != nil {
		log.Errorf("resp:%+v, %+v", resp, resp.Payload[0])
		return err
	}
	log.Infof("ok: %+v", resp)
	resp, err = channel.MethodCall("setMemoryNotificationsEnabled:", true)
	if err != nil {
		log.Errorf("resp:%+v, %+v", resp, resp.Payload[0])
		return err
	}
	log.Infof("ok: %+v", resp)
	/*
		INFO[0006] 9.0 c5 setApplicationStateNotificationsEnabled: [{t:binary, v:true},]  d=out id="#43"
		INFO[0006] 10.0 c5 setMemoryNotificationsEnabled: [{t:binary, v:true},]  d=out id="#43"*/
	return nil
}
