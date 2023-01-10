package diagnostics

import ios "github.com/danielpaulus/go-ios/ios"

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

func (diagnosticsConn *Connection) MobileGestaltQuery(keys []string) (interface{}, error) {
	err := diagnosticsConn.deviceConn.Send(gestaltRequest(keys))
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
