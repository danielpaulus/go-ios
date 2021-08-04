package instruments

import (
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
)

const conditionInducerChannelName = "com.apple.instruments.server.services.ConditionInducer"

//DeviceStateControl allows to access the ConditionInducer so we can set device states like
//  "SlowNetworkCondition"  and  "SlowNetwork3GGood".
//Use the List() command to get all available ProfileType and Profile combinations.
//Then use Enable() and Disable() to control them.
type DeviceStateControl struct {
	controlChannel *dtx.Channel
	conn           *dtx.Connection
}

//NewDeviceStateControl creates and connects a new DeviceStateControl that is ready to use
func NewDeviceStateControl(device ios.DeviceEntry) (*DeviceStateControl, error) {
	dtxConn, err := connectInstruments(device)
	if err != nil {
		return nil, err
	}
	conditionInducerChannel := dtxConn.RequestChannelIdentifier(
		conditionInducerChannelName,
		loggingDispatcher{dtxConn},
		//ThermalConditions tend to take a lot of time to enable, so we have to increase the timeout here.
		dtx.WithTimeout(120),
	)
	return &DeviceStateControl{controlChannel: conditionInducerChannel, conn: dtxConn}, nil
}

//ProfileType a profile type we can activate
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

//Profile belongs to a ProfileType
type Profile struct {
	Description string
	Identifier  string
	Name        string
}

//VerifyProfileAndType checks that the given string profileTypeIdentifier and profileIdentifier are contained in the given types.
func VerifyProfileAndType(types []ProfileType, profileTypeIdentifier string, profileIdentifier string) (ProfileType, Profile, error) {
	foundProfileType := false
	foundProfile := false
	var resultType ProfileType
	var resultProfile Profile
	for _, profileType := range types {
		if profileType.Identifier == profileTypeIdentifier {
			resultType = profileType
			foundProfileType = true
			for _, profile := range profileType.Profiles {
				if profile.Identifier == profileIdentifier {
					foundProfile = true
					resultProfile = profile
				}
			}
		}
	}
	if foundProfileType && foundProfile {
		return resultType, resultProfile, nil
	}
	return ProfileType{}, Profile{}, fmt.Errorf("ProfiletypeIdentifier '%s' valid: %v.  Profile identifier %s valid:%v", profileTypeIdentifier, foundProfileType, profileIdentifier, foundProfile)
}

//List returns a list of all available profile types and profiles.
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

//Enable activates a given profileType and profile received from a List command.
//Note, that the device will automatically deactivate the profile if this dtx connection closes
// f.ex. when the process is terminated. Make sure to keep it open if you use this and use the Disable command.
func (d DeviceStateControl) Enable(pType ProfileType, profile Profile) error {
	response, err := d.controlChannel.MethodCall("enableConditionWithIdentifier:profileIdentifier:", pType.Identifier, profile.Identifier)
	if err != nil {
		return err
	}
	success, ok := response.Payload[0].(bool)
	if !ok || !success {
		return fmt.Errorf("failed enabling profile %+v", response)
	}
	return err
}

//Disable deactivates the currently active profileType
func (d DeviceStateControl) Disable(pType ProfileType) error {
	response, err := d.controlChannel.MethodCall("disableConditionWithIdentifier:", pType.Identifier)
	if err != nil {
		return err
	}
	success, ok := response.Payload[0].(bool)
	if !ok || !success {
		return fmt.Errorf("failed enabling profile %+v", response)
	}
	return err
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
