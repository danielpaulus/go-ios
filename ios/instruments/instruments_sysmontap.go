package instruments

import (
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	log "github.com/sirupsen/logrus"
)

type sysmontapMsgDispatcher struct {
	messages chan dtx.Message
}

func newSysmontapMsgDispatcher() *sysmontapMsgDispatcher {
	return &sysmontapMsgDispatcher{make(chan dtx.Message)}
}

func (p *sysmontapMsgDispatcher) Dispatch(m dtx.Message) {
	p.messages <- m
}

const sysmontapName = "com.apple.instruments.server.services.sysmontap"

type sysmontapService struct {
	channel *dtx.Channel
	conn    *dtx.Connection

	deviceInfoService *DeviceInfoService
	msgDispatcher     *sysmontapMsgDispatcher
}

// NewSysmontapService creates a new sysmontapService
// - samplingInterval is the rate how often to get samples, i.e Xcode's default is 10, which results in sampling output
// each 1 second, with 500 the samples are retrieved every 15 seconds. It doesn't make any correlation between
// the expected rate and the actual rate of samples delivery. We can only conclude, that the lower the rate in digits,
// the faster the samples are delivered
func NewSysmontapService(device ios.DeviceEntry, samplingInterval int) (*sysmontapService, error) {
	deviceInfoService, err := NewDeviceInfoService(device)
	if err != nil {
		return nil, err
	}

	msgDispatcher := newSysmontapMsgDispatcher()
	dtxConn, err := connectInstrumentsWithMsgDispatcher(device, msgDispatcher)
	if err != nil {
		return nil, err
	}

	processControlChannel := dtxConn.RequestChannelIdentifier(sysmontapName, loggingDispatcher{dtxConn})

	sysAttrs, err := deviceInfoService.systemAttributes()
	if err != nil {
		return nil, err
	}

	procAttrs, err := deviceInfoService.processAttributes()
	if err != nil {
		return nil, err
	}

	config := map[string]interface{}{
		"ur":             samplingInterval,
		"bm":             0,
		"procAttrs":      procAttrs,
		"sysAttrs":       sysAttrs,
		"cpuUsage":       true,
		"physFootprint":  true,
		"sampleInterval": 500000000,
	}
	_, err = processControlChannel.MethodCall("setConfig:", config)
	if err != nil {
		return nil, err
	}

	err = processControlChannel.MethodCallAsync("start")
	if err != nil {
		return nil, err
	}

	return &sysmontapService{processControlChannel, dtxConn, deviceInfoService, msgDispatcher}, nil
}

// Close closes up the DTX connection, message dispatcher and dtx.Message channel
func (s *sysmontapService) Close() error {
	close(s.msgDispatcher.messages)

	s.deviceInfoService.Close()
	return s.conn.Close()
}

// ReceiveCPUUsage returns a chan of SysmontapMessage with CPU Usage info
// The method will close the result channel automatically as soon as sysmontapMsgDispatcher's
// dtx.Message channel is closed.
func (s *sysmontapService) ReceiveCPUUsage() chan SysmontapMessage {
	messages := make(chan SysmontapMessage)
	go func() {
		defer close(messages)

		for msg := range s.msgDispatcher.messages {
			sysmontapMessage, err := mapToCPUUsage(msg)
			if err != nil {
				log.Debugf("expected `sysmontapMessage` from global channel, but received %v", msg)
				continue
			}

			messages <- sysmontapMessage
		}

		log.Infof("sysmontap message dispatcher channel closed")
	}()

	return messages
}

// SysmontapMessage is a wrapper struct for incoming CPU samples
type SysmontapMessage struct {
	CPUCount       uint64
	EnabledCPUs    uint64
	EndMachAbsTime uint64
	Type           uint64
	SystemCPUUsage CPUUsage
}

type CPUUsage struct {
	CPU_TotalLoad float64
}

func mapToCPUUsage(msg dtx.Message) (SysmontapMessage, error) {
	payload := msg.Payload
	if len(payload) != 1 {
		return SysmontapMessage{}, fmt.Errorf("payload of message should have only one element: %+v", msg)
	}

	resultArray, ok := payload[0].([]interface{})
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected resultArray of type []interface{}: %+v", payload[0])
	}
	resultMap, ok := resultArray[0].(map[string]interface{})
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected resultMap of type map[string]interface{} as a single element of resultArray: %+v", resultArray[0])
	}
	cpuCount, ok := resultMap["CPUCount"].(uint64)
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected CPUCount of type uint64 of resultMap: %+v", resultMap)
	}
	enabledCPUs, ok := resultMap["EnabledCPUs"].(uint64)
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected EnabledCPUs of type uint64 of resultMap: %+v", resultMap)
	}
	endMachAbsTime, ok := resultMap["EndMachAbsTime"].(uint64)
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected EndMachAbsTime of type uint64 of resultMap: %+v", resultMap)
	}
	typ, ok := resultMap["Type"].(uint64)
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected Type of type uint64 of resultMap: %+v", resultMap)
	}
	sysmontapMessageMap, ok := resultMap["SystemCPUUsage"].(map[string]interface{})
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected SystemCPUUsage of type map[string]interface{} of resultMap: %+v", resultMap)
	}
	cpuTotalLoad, ok := sysmontapMessageMap["CPU_TotalLoad"].(float64)
	if !ok {
		return SysmontapMessage{}, fmt.Errorf("expected CPU_TotalLoad of type uint64 of sysmontapMessageMap: %+v", sysmontapMessageMap)
	}
	cpuUsage := CPUUsage{CPU_TotalLoad: cpuTotalLoad}

	sysmontapMessage := SysmontapMessage{
		cpuCount,
		enabledCPUs,
		endMachAbsTime,
		typ,
		cpuUsage,
	}
	return sysmontapMessage, nil
}
