package instruments

import (
	"time"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	log "github.com/sirupsen/logrus"
)

const deviceInfoServiceName = "com.apple.instruments.server.services.deviceinfo"

//ProcessInfo contains all the properties for a process
//running on an iOS devices that we get back from instruments
type ProcessInfo struct {
	IsApplication bool
	Name          string
	Pid           uint64
	RealAppName   string
	StartDate     time.Time
}

//ProcessList returns a []ProcessInfo, one for each process running on the iOS device
func (p DeviceInfoService) ProcessList() ([]ProcessInfo, error) {
	resp, err := p.channel.MethodCall("runningProcesses")
	result := mapToProcInfo(resp.Payload[0].([]interface{}))
	return result, err
}

//NameForPid resolves a process name for a given pid
func (p DeviceInfoService) NameForPid(pid uint64) error {
	_, err := p.channel.MethodCall("execnameForPid:", pid)
	return err
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
		if date, ok := procMap["startDate"]; ok {
			procInf.StartDate = date.(nskeyedarchiver.NSDate).Timestamp
		}
		result[i] = procInf

	}
	return result
}

type deviceInfoDispatcher struct {
	conn *dtx.Connection
}

func (p deviceInfoDispatcher) Dispatch(m dtx.Message) {
	dtx.SendAckIfNeeded(p.conn, m)
	log.Debug(m)
}

//DeviceInfoService gives us access to retrieving process lists and resolving names for PIDs
type DeviceInfoService struct {
	channel *dtx.Channel
	conn    *dtx.Connection
}

//NewDeviceInfoService creates a new DeviceInfoService for a given device
func NewDeviceInfoService(device ios.DeviceEntry) (*DeviceInfoService, error) {
	dtxConn, err := dtx.NewConnection(device, serviceName)
	if err != nil {
		log.Debugf("Failed connecting to %s, trying %s", serviceName, serviceNameiOS14)
		dtxConn, err = dtx.NewConnection(device, serviceNameiOS14)
		if err != nil {
			return nil, err
		}
	}
	processControlChannel := dtxConn.RequestChannelIdentifier(deviceInfoServiceName, loggingDispatcher{dtxConn})
	return &DeviceInfoService{channel: processControlChannel, conn: dtxConn}, nil
}

//Close closes up the DTX connection
func (d *DeviceInfoService) Close() {
	d.conn.Close()
}
