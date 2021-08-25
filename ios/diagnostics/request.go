package diagnostics

import (
	"bytes"
	"github.com/danielpaulus/go-ios/ios"

	plist "howett.net/plist"
)

type diagnosticsRequest struct {
	Request string
}

func gestaltRequest(keys []string) []byte {
	goodbyeMap := map[string]interface{}{
		"Request":           "MobileGestalt",
		"MobileGestaltKeys": keys,
	}
	bt, err := ios.PlistCodec{}.Encode(goodbyeMap)
	if err != nil {
		panic("gestalt request encoding should never fail")
	}
	return bt
}

func goodBye() []byte {
	goodbyeMap := map[string]interface{}{
		"Request": "Goodbye",
	}
	bt, err := ios.PlistCodec{}.Encode(goodbyeMap)
	if err != nil {
		panic("goodbye request encoding should never fail")
	}
	return bt
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
	GasGauge GasGauge
	HDMI     HDMI
	NAND     NAND
	WiFi     WiFi
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
