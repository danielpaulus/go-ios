package instruments

import (
	"time"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	"github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
)

const deviceInfoServiceName = "com.apple.instruments.server.services.deviceinfo"

// ProcessInfo contains all the properties for a process
// running on an iOS devices that we get back from instruments
type ProcessInfo struct {
	IsApplication bool
	Name          string
	Pid           uint64
	RealAppName   string
	StartDate     time.Time
}

// processAttributes returns the attributes list which can be used for monitoring
func (d DeviceInfoService) processAttributes() ([]interface{}, error) {
	resp, err := d.channel.MethodCall("sysmonProcessAttributes")
	if err != nil {
		return nil, err
	}
	return resp.Payload[0].([]interface{}), nil
}

// systemAttributes returns the attributes list which can be used for monitoring
func (d DeviceInfoService) systemAttributes() ([]interface{}, error) {
	resp, err := d.channel.MethodCall("sysmonSystemAttributes")
	if err != nil {
		return nil, err
	}
	return resp.Payload[0].([]interface{}), nil
}

// ProcessList returns a []ProcessInfo, one for each process running on the iOS device
func (d DeviceInfoService) ProcessList() ([]ProcessInfo, error) {
	resp, err := d.channel.MethodCall("runningProcesses")
	if err != nil {
		return nil, err
	}

	if len(resp.Payload) == 0 {
		return []ProcessInfo{}, nil
	}

	result := mapToProcInfo(resp.Payload[0].([]interface{}))
	return result, err
}

// NameForPid resolves a process name for a given pid
func (d DeviceInfoService) NameForPid(pid uint64) error {
	_, err := d.channel.MethodCall("execnameForPid:", pid)
	return err
}

// HardwareInformation gets some nice extra details from Instruments. Here is an example result for an old iPhone 5:
// map[hwCPU64BitCapable:1 hwCPUsubtype:1 hwCPUtype:16777228 numberOfCpus:2 numberOfPhysicalCpus:2 speedOfCpus:0]
func (d DeviceInfoService) HardwareInformation() (map[string]interface{}, error) {
	response, err := d.channel.MethodCall("hardwareInformation")
	if err != nil {
		return map[string]interface{}{}, err
	}
	return extractMapPayload(response)
}

// NetworkInformation gets a list of all network interfaces for the device. Example result:
// map[en0:Wi-Fi en1:Ethernet Adaptor (en1) en2:Ethernet Adaptor (en2) lo0:Loopback pdp_ip0:Cellular (pdp_ip0)
// pdp_ip1:Cellular (pdp_ip1) pdp_ip2:Cellular (pdp_ip2) pdp_ip3:Cellular (pdp_ip3) pdp_ip4:Cellular (pdp_ip4)]
func (d DeviceInfoService) NetworkInformation() (map[string]interface{}, error) {
	response, err := d.channel.MethodCall("networkInformation")
	if err != nil {
		return map[string]interface{}{}, err
	}
	return extractMapPayload(response)
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

// DeviceInfoService gives us access to retrieving process lists and resolving names for PIDs
type DeviceInfoService struct {
	channel *dtx.Channel
	conn    *dtx.Connection
}

// NewDeviceInfoService creates a new DeviceInfoService for a given device
func NewDeviceInfoService(device ios.DeviceEntry) (*DeviceInfoService, error) {
	dtxConn, err := connectInstruments(device)
	if err != nil {
		return nil, err
	}
	processControlChannel := dtxConn.RequestChannelIdentifier(deviceInfoServiceName, loggingDispatcher{dtxConn})
	return &DeviceInfoService{channel: processControlChannel, conn: dtxConn}, nil
}

// Close closes up the DTX connection
func (d *DeviceInfoService) Close() {
	d.conn.Close()
}
