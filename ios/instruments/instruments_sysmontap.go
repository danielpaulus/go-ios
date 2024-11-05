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
	channel       *dtx.Channel
	conn          *dtx.Connection
	msgDispatcher *sysmontapMsgDispatcher
}

// Creates a new sysmontapService
func newSysmontapService(device ios.DeviceEntry) (*sysmontapService, error) {
	msgDispatcher := newSysmontapMsgDispatcher()
	dtxConn, err := connectInstrumentsWithMsgDispatcher(device, msgDispatcher)
	if err != nil {
		return nil, err
	}

	processControlChannel := dtxConn.RequestChannelIdentifier(sysmontapName, loggingDispatcher{dtxConn})

	return &sysmontapService{channel: processControlChannel, conn: dtxConn, msgDispatcher: msgDispatcher}, nil
}

// Close closes up the DTX connection
func (s *sysmontapService) Close() error {
	close(s.msgDispatcher.messages)
	return s.conn.Close()
}

// start sends a start method call async and waits until the cpu info & stats come back
// the method is a part of the @protocol DTTapAuthorizedAPI
func (s *sysmontapService) start() (SysmontapMessage, error) {
	err := s.channel.MethodCallAsync("start")
	if err != nil {
		return SysmontapMessage{}, err
	}

	for {
		select {
		case msg := <-s.msgDispatcher.messages:
			sysmontapMessage, err := mapToCPUUsage(msg)
			if err != nil {
				log.Debug(fmt.Sprintf("expected `sysmontapMessage` from global channel, but was %v", msg))
				continue
			}

			return sysmontapMessage, nil
		}
	}
}

// setConfig sets configuration to allow the sysmontap service getting desired data points
func (s *sysmontapService) setConfig(procAttrs, sysAttrs []interface{}) error {
	config := map[string]interface{}{
		"ur":             500,
		"bm":             0,
		"procAttrs":      procAttrs,
		"sysAttrs":       sysAttrs,
		"cpuUsage":       true,
		"physFootprint":  true,
		"sampleInterval": 500000000,
	}

	_, err := s.channel.MethodCall("setConfig:", config)

	if err != nil {
		return err
	}

	return nil
}

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
