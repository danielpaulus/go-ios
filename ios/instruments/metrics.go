package instruments

import (
	"encoding/json"
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

type yo struct {
}

func (y yo) Dispatch(msg dtx.Message) {

	log.Infof("%+v", msg.Payload[0])
	if "applicationStateNotification:" == msg.Payload[0].(string) {
		data, err := nskeyedarchiver.Unarchive(msg.Auxiliary.GetArguments()[0].([]byte))
		if err != nil {
			log.Warn(err)
		}
		resp := data[0]
		jsonString, err := json.Marshal(resp)
		println(string(jsonString))
		//log.Infof("%+v", data)
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
