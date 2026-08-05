package instruments

import (
	"fmt"
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
	if len(resp.Payload) == 0 {
		return nil, fmt.Errorf("sysmonProcessAttributes: empty payload")
	}
	attrs, ok := resp.Payload[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("sysmonProcessAttributes: expected []interface{} payload, got %T", resp.Payload[0])
	}
	return attrs, nil
}

// systemAttributes returns the attributes list which can be used for monitoring
func (d DeviceInfoService) systemAttributes() ([]interface{}, error) {
	resp, err := d.channel.MethodCall("sysmonSystemAttributes")
	if err != nil {
		return nil, err
	}
	if len(resp.Payload) == 0 {
		return nil, fmt.Errorf("sysmonSystemAttributes: empty payload")
	}
	attrs, ok := resp.Payload[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("sysmonSystemAttributes: expected []interface{} payload, got %T", resp.Payload[0])
	}
	return attrs, nil
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

	procList, ok := resp.Payload[0].([]interface{})
	if !ok {
		return nil, fmt.Errorf("runningProcesses: expected []interface{} payload, got %T", resp.Payload[0])
	}
	return mapToProcInfo(procList)
}

// ProcessByName looks up a running process by name or real app name and
// returns its ProcessInfo. Returns an error if no match is found.
func (d DeviceInfoService) ProcessByName(name string) (ProcessInfo, error) {
	procs, err := d.ProcessList()
	if err != nil {
		return ProcessInfo{}, err
	}
	for _, p := range procs {
		if p.Name == name || p.RealAppName == name {
			return p, nil
		}
	}
	return ProcessInfo{}, fmt.Errorf("no running process found with name %q", name)
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

func mapToProcInfo(procList []interface{}) ([]ProcessInfo, error) {
	result := make([]ProcessInfo, len(procList))
	for i, procMapInt := range procList {
		procMap, ok := procMapInt.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("process entry %d: expected map[string]interface{}, got %T", i, procMapInt)
		}
		procInf := ProcessInfo{}
		isApp, ok := procMap["isApplication"].(bool)
		if !ok {
			return nil, fmt.Errorf("process entry %d: expected bool isApplication, got %T: %+v", i, procMap["isApplication"], procMap["isApplication"])
		}
		procInf.IsApplication = isApp
		name, ok := procMap["name"].(string)
		if !ok {
			return nil, fmt.Errorf("process entry %d: expected string name, got %T: %+v", i, procMap["name"], procMap["name"])
		}
		procInf.Name = name
		pid, ok := toUint64(procMap["pid"])
		if !ok {
			return nil, fmt.Errorf("process entry %d: expected numeric pid, got %T: %+v", i, procMap["pid"], procMap["pid"])
		}
		procInf.Pid = pid
		realAppName, ok := procMap["realAppName"].(string)
		if !ok {
			return nil, fmt.Errorf("process entry %d: expected string realAppName, got %T: %+v", i, procMap["realAppName"], procMap["realAppName"])
		}
		procInf.RealAppName = realAppName
		if date, ok := procMap["startDate"]; ok {
			nsDate, ok := date.(nskeyedarchiver.NSDate)
			if !ok {
				return nil, fmt.Errorf("process entry %d: expected NSDate startDate, got %T: %+v", i, date, date)
			}
			procInf.StartDate = nsDate.Timestamp
		}
		result[i] = procInf

	}
	return result, nil
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
