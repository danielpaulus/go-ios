package screenshotr

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"

	"howett.net/plist"
)

type (
	plistArray  []interface{}
	versionInfo struct {
		major uint64
		minor uint64
	}
)

const (
	dLMessageVersionExchange = "DLMessageVersionExchange"
	dlMessageProcessMessage  = "DLMessageProcessMessage"
)

func newVersionExchangeRequest(versionMajor uint64) plistArray {
	return []interface{}{dLMessageVersionExchange, "DLVersionsOk", versionMajor}
}

func getArrayFromBytes(plistBytes []byte) (plistArray, error) {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data plistArray
	err := decoder.Decode(&data)
	return data, err
}

func getVersionfromBytes(plistBytes []byte) (versionInfo, error) {
	data, err := getArrayFromBytes(plistBytes)
	if err != nil {
		return versionInfo{}, fmt.Errorf("failed decoding bytes: %x to array with error %w", plistBytes, err)
	}
	if len(data) > 3 {
		log.Warnf("expected 2 items in version exchange response, received %+v", data)
	}
	if len(data) < 3 {
		return versionInfo{}, fmt.Errorf("expected 3 items in array, got: %+v", data)
	}
	typeName, ok := data[0].(string)
	if !ok || typeName != dLMessageVersionExchange {
		return versionInfo{},
			fmt.Errorf("version response array should contain DLMessageVersionExchange but '%s'", typeName)
	}

	major, ok := data[1].(uint64)
	if !ok {
		return versionInfo{}, fmt.Errorf("could not extract major version")
	}
	minor, ok := data[2].(uint64)
	if !ok {
		return versionInfo{}, fmt.Errorf("could not extract minor version")
	}

	return versionInfo{major, minor}, nil
}

type screenShotRequest struct {
	MessageType string
}

func newScreenShotRequest() plistArray {
	request := make([]interface{}, 2)
	request[0] = "DLMessageProcessMessage"
	request[1] = screenShotRequest{"ScreenShotRequest"}
	return request
}
