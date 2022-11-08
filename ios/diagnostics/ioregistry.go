package diagnostics

import (
	ios "github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

func ioregentryRequest(key string) []byte {
	requestMap := map[string]interface{}{
		"Request":   "IORegistry",
		"EntryName": key,
	}
	bt, err := ios.PlistCodec{}.Encode(requestMap)
	if err != nil {
		log.WithError(err).Error("query request encoding should never fail")
	}
	return bt
}

func (diagnosticsConn *Connection) IORegEntryQuery(key string) (interface{}, error) {
	err := diagnosticsConn.deviceConn.Send(ioregentryRequest(key))
	if err != nil {
		return "", err
	}
	respBytes, err := diagnosticsConn.plistCodec.Decode(diagnosticsConn.deviceConn.Reader())
	if err != nil {
		return "", err
	}
	plist, err := ios.ParsePlist(respBytes)
	return plist, err
}
