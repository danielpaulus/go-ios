package nskeyedarchiver

import "howett.net/plist"

var classes map[string]func(map[string]interface{}, []interface{}) interface{}

func SetupClasses() {
	classes = map[string]func(map[string]interface{}, []interface{}) interface{}{
		"DTActivityTraceTapMessage": NewDTActivityTraceTapMessage,
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
