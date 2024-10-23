package diagnostics

import (
	"bytes"

	plist "howett.net/plist"
)

type diagnosticsRequest struct {
	Request string
}

func diagnosticsfromBytes(plistBytes []byte) allDiagnosticsResponse {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data allDiagnosticsResponse
	_ = decoder.Decode(&data)
	return data
}

type rebootRequest struct {
	Request           string
	WaitForDisconnect bool
	DisplayPass       bool
	DisplayFail       bool
}

type allDiagnosticsResponse struct {
	Diagnostics Diagnostics
	Status      string
}

type Diagnostics struct {
	GasGauge   GasGauge
	HDMI       HDMI
	NAND       NAND
	WiFi       WiFi
	IORegistry IORegistry
}

// IORegistry relates to the battery stats
type IORegistry struct {
	InstantAmperage int
	Temperature     int
	Voltage         int
	IsCharging      bool
	CurrentCapacity int
}

type WiFi struct {
	Active string
	Status string
}

type NAND struct {
	Status string
}

type HDMI struct {
	Connection string
	Status     string
}

type GasGauge struct {
	CycleCount         uint64
	DesignCapacity     uint64
	FullChargeCapacity uint64
	Status             string
}
