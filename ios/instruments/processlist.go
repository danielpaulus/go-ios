package instruments

import (
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	log "github.com/sirupsen/logrus"
)

const deviceInfoServiceName = "com.apple.instruments.server.services.deviceinfo"

func (p DeviceInfoService) ProcessList() error {
	resp, err := p.channel.MethodCall("runningProcesses")
	log.Info("%+v", resp)
	return err
}

func (p DeviceInfoService) NameForPid(pid uint64) error {
	_, err := p.channel.MethodCall("execnameForPid:", pid)
	return err
}

type deviceInfoDispatcher struct {
	conn *dtx.Connection
}

func (p deviceInfoDispatcher) Dispatch(m dtx.Message) {
	dtx.SendAckIfNeeded(p.conn, m)
	log.Debug(m)
}

type DeviceInfoService struct {
	channel *dtx.Channel
	conn    *dtx.Connection
}

func NewDeviceInfoService(device ios.DeviceEntry) (*DeviceInfoService, error) {
	dtxConn, err := dtx.NewConnection(device, serviceName)
	if err != nil {
		log.Debugf("Failed connecting to %s, trying %s", serviceName, serviceNameiOS14)
		dtxConn, err = dtx.NewConnection(device, serviceNameiOS14)
		if err != nil {
			return nil, err
		}
	}
	processControlChannel := dtxConn.RequestChannelIdentifier(deviceInfoServiceName, processControlDispatcher{dtxConn})
	return &DeviceInfoService{channel: processControlChannel, conn: dtxConn}, nil
}
