package instruments

import (
	"fmt"
	"time"

	"github.com/danielpaulus/go-ios/ios"
)

type PerfOptions struct {

	// system
	SysCPU     bool `json:"sys_cpu,omitempty" yaml:"sys_cpu,omitempty"`
	SysMem     bool `json:"sys_mem,omitempty" yaml:"sys_mem,omitempty"`
	SysDisk    bool `json:"sys_disk,omitempty" yaml:"sys_disk,omitempty"`
	SysNetwork bool `json:"sys_network,omitempty" yaml:"sys_network,omitempty"`
	gpu        bool
	FPS        bool `json:"fps,omitempty" yaml:"fps,omitempty"`
	Network    bool `json:"network,omitempty" yaml:"network,omitempty"`
	// process
	BundleID string `json:"bundle_id,omitempty" yaml:"bundle_id,omitempty"`
	Pid      int    `json:"pid,omitempty" yaml:"pid,omitempty"`
	// config
	OutputInterval    int      `json:"output_interval,omitempty" yaml:"output_interval,omitempty"` // ms
	SystemAttributes  []string `json:"system_attributes,omitempty" yaml:"system_attributes,omitempty"`
	ProcessAttributes []string `json:"process_attributes,omitempty" yaml:"process_attributes,omitempty"`
}

func defaulPerfOption() *PerfOptions {
	return &PerfOptions{
		SysCPU:         false,
		SysMem:         false,
		SysDisk:        false,
		SysNetwork:     false,
		gpu:            false,
		FPS:            false,
		Network:        false,
		OutputInterval: 1000, // default 1000ms
		// SystemAttributes:  []string{"vmExtPageCount", "vmFreeCount", "vmPurgeableCount", "vmSpeculativeCount", "physMemSize"},
		// ProcessAttributes: []string{"memVirtualSize", "cpuUsage", "ctxSwitch", "intWakeups", "physFootprint", "memResidentSize", "memAnon", "pid"},
		SystemAttributes: []string{
			// disk
			"diskBytesRead",
			"diskBytesWritten",
			"diskReadOps",
			"diskWriteOps",
			// memory
			"vmCompressorPageCount",
			"vmExtPageCount",
			"vmFreeCount",
			"vmIntPageCount",
			"vmPurgeableCount",
			"vmWireCount",
			"vmUsedCount",
			"__vmSwapUsage",
			// network
			"netBytesIn",
			"netBytesOut",
			"netPacketsIn",
			"netPacketsOut",
		},
		ProcessAttributes: []string{
			"pid",
			"cpuUsage",
		},
	}
}

func ListenSysmontap(device ios.DeviceEntry) (func() (map[string]interface{}, error), func() error, error) {
	conn, err := connectInstruments(device)
	if err != nil {
		return nil, nil, err
	}

	channel := conn.RequestChannelIdentifier(Sysmontap, channelDispatcher{})
	options := defaulPerfOption()
	// interval := time.Millisecond * time.Duration(options.OutputInterval)
	// config := map[string]interface{}{
	// 	"bm":             0,
	// 	"cpuUsage":       true,
	// 	"procAttrs":      []string{"memVirtualSize", "cpuUsage", "ctxSwitch", "intWakeups", "physFootprint", "memResidentSize", "memAnon", "pid"},
	// 	"sampleInterval": interval,
	// 	"sysAttrs":       []string{"vmExtPageCount", "vmFreeCount", "vmPurgeableCount", "vmSpeculativeCount", "physMemSize"},
	// 	"ur":             1000,
	// }
	config := map[string]interface{}{
		"bm":             0,
		"cpuUsage":       true,
		"sampleInterval": options.OutputInterval,    // time.Duration
		"ur":             options.OutputInterval,    // 输出频率
		"procAttrs":      options.ProcessAttributes, // process performance
		"sysAttrs":       options.SystemAttributes,  // system performance
	}
	channel.MethodCall("setConfig:", config)
	channel.MethodCall("start")

	time.Sleep(time.Duration(3) * time.Second)
	channel.MethodCall("stop")
	conn.Close()
	fmt.Println("conn.Close")
	return nil, nil, nil

}
