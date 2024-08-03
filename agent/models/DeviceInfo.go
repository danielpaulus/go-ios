package models

import (
	"time"

	"github.com/danielpaulus/go-ios/agent/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type DevicePool struct {
	Hostname string
	Ip       string
	Port     string
	ID       string
	Devices  []DeviceInfo
}

const (
	DeviceTypeIos               string = "ios"
	DeviceTypeAndroid           string = "android"
	ConnectionStateConnected           = "connected"
	ConnectionStateDisconnected        = "disconnected"
	SessionStateFree                   = "free"
	SessionStateInUse                  = "inuse"
)

type DeviceInfo struct {
	Serial                  string
	ConfigurationState      ConfigurationState
	SessionState            SessionState
	PhysicalConnectionState PhysicalConnectionState
	Name                    string
	DeviceType              string
	MetaInfo                map[string]interface{}
}

// UNUSED
// type DeviceAuditLogEntry struct {
// 	id           uuid.UUID
// 	TimeOfChange time.Time
// 	NewState     ConfigurationState
// }

type PhysicalConnectionState struct {
	//is this device physically connected to the host, yes or no
	ConnectionState string
	//extra info
	MetaInfo map[string]interface{}
	// When was this device seen last
	LastDetected time.Time
}

func (p PhysicalConnectionState) Equals(other PhysicalConnectionState) bool {
	return p.ConnectionState == other.ConnectionState &&
		p.LastDetected == other.LastDetected &&
		utils.MapEquals(p.MetaInfo, other.MetaInfo)
}

type SessionState struct {
	// is this device allocated by some user, then set to SessionStateInUse
	// otherwise set to SessionStateFree
	SessionState string
	// Last time the session owner sent a heartbeat to us to keep the session alive
	SessionStateLastPing time.Time
	//extra info
	MetaInfo map[string]interface{}
	// Session key
	SessionKey uuid.UUID
	// AccountID
	AccountID uuid.UUID
	// SessionTimeout
	SessionTimeout time.Duration
}

func (s SessionState) Equals(state SessionState) bool {
	return s.SessionState == state.SessionState &&
		s.SessionStateLastPing == state.SessionStateLastPing &&
		s.SessionKey == state.SessionKey &&
		s.AccountID == state.AccountID &&
		utils.MapEquals(s.MetaInfo, state.MetaInfo)
}

type ConfigurationState struct {
	//extra info
	MetaInfo map[string]interface{}
	// iOS specific, has supervision enabled
	Supervised bool
	// everything related to developer mode is enabled
	// if not, put details in MetaInfo
	DeveloperModeEnabled bool
	// device is responsive to basic commands
	BasicCommandsWork bool
	// device allows to install apps
	CanInstallApps bool
	// device has everything installed and setup to allow automation
	// f.ex. iOS device get this when they can run XCTest
	DeviceAutomationAvailable bool
}

func (c ConfigurationState) Equals(other ConfigurationState) bool {
	return c.DeviceAutomationAvailable == other.DeviceAutomationAvailable &&
		c.Supervised == other.Supervised &&
		c.DeveloperModeEnabled == other.DeveloperModeEnabled &&
		c.BasicCommandsWork == other.BasicCommandsWork &&
		c.CanInstallApps == other.CanInstallApps &&
		utils.MapEquals(c.MetaInfo, other.MetaInfo)
}

func NewDeviceStateAndroid() ConfigurationState {
	return ConfigurationState{MetaInfo: map[string]interface{}{}}

}

func NewDeviceStateIos() ConfigurationState {
	return ConfigurationState{MetaInfo: map[string]interface{}{}}
}

func UpdateDeviceInfo(newState []DeviceInfo) error {
	log.Info("update device info")
	pool, err := LoadOrInitPool()
	if err != nil {
		return err
	}
	pool.Devices = newState
	err = UpdatePool(pool)
	return err
}

func GetDeviceInfo() ([]DeviceInfo, error) {
	pool, err := LoadOrInitPool()
	if err != nil {
		return []DeviceInfo{}, err
	}
	return pool.Devices, nil
}
