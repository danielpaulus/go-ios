package nskeyedarchiver

import (
	"howett.net/plist"
)

var classes map[string]func(map[string]interface{}, []interface{}) interface{}

func SetupClasses() {
	classes = map[string]func(map[string]interface{}, []interface{}) interface{}{
		"DTActivityTraceTapMessage": NewDTActivityTraceTapMessage,
		"NSError":                   NewNSError,
	}
}

type DTActivityTraceTapMessage struct {
	DTTapMessagePlist map[string]interface{}
}

func NewDTActivityTraceTapMessage(object map[string]interface{}, objects []interface{}) interface{} {
	ref := object["DTTapMessagePlist"].(plist.UID)
	plist, _ := extractDictionary(objects[ref].(map[string]interface{}), objects)
	return DTActivityTraceTapMessage{DTTapMessagePlist: plist}
}

type NSError struct {
	ErrorCode uint64
	Domain    string
	UserInfo  map[string]interface{}
}

func NewNSError(object map[string]interface{}, objects []interface{}) interface{} {
	errorCode := object["NSCode"].(uint64)
	userInfo_ref := object["NSUserInfo"].(plist.UID)
	domain_ref := object["NSDomain"].(plist.UID)
	domain := objects[domain_ref].(string)
	userinfo, _ := extractDictionary(objects[userInfo_ref].(map[string]interface{}), objects)

	return NSError{ErrorCode: errorCode, Domain: domain, UserInfo: userinfo}
}
