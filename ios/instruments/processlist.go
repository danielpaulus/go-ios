package instruments

import (
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	log "github.com/sirupsen/logrus"
)

const deviceInfoServiceName = "com.apple.instruments.server.services.deviceinfo"

type ProcessInfo struct {
	IsApplication bool
	Name          string
	Pid           uint64
	RealAppName   string
	StartDate     string
}

func (p DeviceInfoService) ProcessList() ([]ProcessInfo, error) {
	resp, err := p.channel.MethodCall("runningProcesses")
	result := mapToProcInfo(resp.Payload[0].([]interface{}))
	return result, err
}

func mapToProcInfo(procList []interface{}) []ProcessInfo {
	result := make([]ProcessInfo, len(procList))
	for i, procMapInt := range procList {
		procMap := procMapInt.(map[string]interface{})
		procInf := ProcessInfo{}
		procInf.IsApplication = procMap["isApplication"].(bool)
		procInf.Name = procMap["name"].(string)
		procInf.Pid = procMap["pid"].(uint64)
		procInf.RealAppName = procMap["realAppName"].(string)
		//procInf.StartDate = procMap["startDate"].(nskeyedarchiver.NSDate).String()
		result[i] = procInf

	}
	return result
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
