package instruments

import (
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
)

const conditionInducerChannelName = "com.apple.instruments.server.services.ConditionInducer"

type DeviceStateControl struct {
	controlChannel *dtx.Channel
	conn           *dtx.Connection
}

func NewDeviceStateControl(device ios.DeviceEntry) (*DeviceStateControl, error) {
	dtxConn, err := connectInstruments(device)
	if err != nil {
		return nil, err
	}
	conditionInducerChannel := dtxConn.RequestChannelIdentifier(conditionInducerChannelName, loggingDispatcher{dtxConn})
	return &DeviceStateControl{controlChannel: conditionInducerChannel, conn: dtxConn}, nil
}

type ProfileType struct {
	ActiveProfile  string
	Identifier     string
	ProfilesSorted bool
	IsActive       bool
	Name           string
	IsDestructive  bool
	IsInternal     bool
	Profiles       []Profile
}

type Profile struct {
	Description string
	Identifier  string
	Name        string
}

func (d DeviceStateControl) List() ([]ProfileType, error) {
	const methodName = "availableConditionInducers"
	response, err := d.controlChannel.MethodCall(methodName)
	if err != nil {
		return []ProfileType{}, err
	}
	profiles, err := decodeProfileTypes(response.Payload[0])
	if err != nil {
		return []ProfileType{}, err
	}
	return profiles, nil
}

func decodeProfileTypes(response interface{}) ([]ProfileType, error) {
	profileTypeList, ok := response.([]interface{})
	if !ok {
		return []ProfileType{}, fmt.Errorf("invalid response: %+v", response)
	}
	result := make([]ProfileType, len(profileTypeList))
	for i, rawProfile := range profileTypeList {
		profileMap := rawProfile.(map[string]interface{})
		profiles, err := decodeProfiles(profileMap)
		if err != nil {
			return []ProfileType{}, err
		}
		p := ProfileType{
			ActiveProfile:  profileMap["activeProfile"].(string),
			Identifier:     profileMap["identifier"].(string),
			IsActive:       profileMap["isActive"].(bool),
			IsDestructive:  profileMap["isDestructive"].(bool),
			IsInternal:     profileMap["isInternal"].(bool),
			Name:           profileMap["name"].(string),
			ProfilesSorted: profileMap["profilesSorted"].(bool),
			Profiles:       profiles,
		}
		result[i] = p
	}
	return result, nil
}

func decodeProfiles(profileMap map[string]interface{}) ([]Profile, error) {
	profilesListRaw, ok := profileMap["profiles"]
	if !ok {
		return []Profile{}, fmt.Errorf("failed finding 'profiles' key in map: %+v", profileMap)
	}
	profilesList, ok := profilesListRaw.([]interface{})
	if !ok {
		return []Profile{}, fmt.Errorf("failed converting 'profiles' to list in map: %+v", profileMap)
	}
	result := make([]Profile, len(profilesList))
	for i, profileRaw := range profilesList {
		profile, ok := profileRaw.(map[string]interface{})
		if !ok {
			return []Profile{}, fmt.Errorf("invalid map: %+v", profileMap)
		}
		p := Profile{
			Description: profile["description"].(string),
			Identifier:  profile["identifier"].(string),
			Name:        profile["name"].(string),
		}
		result[i] = p
	}
	return result, nil
}
