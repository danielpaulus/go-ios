package screenshotr

import (
	"bytes"

	plist "howett.net/plist"
)

type plistArray []interface{}
type versionInfo struct {
	major uint64
	minor uint64
}

func newVersionExchangeRequest(versionMajor uint64) plistArray {
	return []interface{}{"DLMessageVersionExchange", "DLVersionsOk", versionMajor}
}

func getArrayFromBytes(plistBytes []byte) plistArray {
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))
	var data plistArray
	_ = decoder.Decode(&data)
	return data
}

func getVersionfromBytes(plistBytes []byte) versionInfo {
	data := getArrayFromBytes(plistBytes)
	version := versionInfo{data[1].(uint64), data[2].(uint64)}
	return version
}
