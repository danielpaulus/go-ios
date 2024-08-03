package devicestatemgmt

import (
	"fmt"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/agent/models"
	"github.com/danielpaulus/go-ios/agent/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var sessionHandles = sync.Map{}

func sessionTimeoutChecker(closer chan bool) {
	for {
		select {
		case <-time.After(utils.SESSION_TIMEOUT_CHECK_INTERVAL_SEC):
			break
		case <-closer:
			return
		}
		log.Info("cleaning up timed out sessions")
		dl := deviceList.GetCopy()
		for _, d := range dl.iosDevices {
			checkSessionTimeout(&d.mux, &d.sessionState)
		}

	}
}

func checkSessionTimeout(s *sync.Mutex, m *models.SessionState) {
	s.Lock()
	defer s.Unlock()
	if m.SessionState != models.SessionStateInUse {
		return
	}
	if time.Now().After(m.SessionStateLastPing.Add(m.SessionTimeout)) {
		log.Infof("session timeout for session %s account %s", m.SessionKey, m.AccountID)
		closeSession(m, m.SessionKey)
	}
}

func TerminateSession(device models.DeviceInfo, accountID uuid.UUID, sessionId uuid.UUID, root bool) error {
	dl := deviceList.GetCopy()
	if device.DeviceType == models.DeviceTypeIos {
		for _, d := range dl.iosDevices {
			if d.udid == device.Serial {
				return terminateSession(&d.mux, &d.sessionState, accountID, sessionId, root)
			}
		}
	}

	return fmt.Errorf("device not found %s", device.Serial)
}

func terminateSession(s *sync.Mutex, m *models.SessionState, accountID uuid.UUID, sessionId uuid.UUID, root bool) error {
	s.Lock()
	defer s.Unlock()
	if m.SessionState != models.SessionStateInUse {
		return fmt.Errorf("session not in use")
	}
	if root {
		closeSession(m, sessionId)
		return nil
	}
	if m.AccountID != accountID {
		return fmt.Errorf("session not in use by this account")
	}
	if m.SessionKey != sessionId {
		return fmt.Errorf("terminate for sessionId %s but actual session key is %s", sessionId, m.SessionKey)
	}
	closeSession(m, sessionId)
	return nil

}

func closeSession(m *models.SessionState, sessionId uuid.UUID) {
	go func() {
		h, ok := sessionHandles.Load(sessionId)
		if !ok {
			return
		}
		h.(chan error) <- nil
		close(h.(chan error))
		sessionHandles.Delete(sessionId)
	}()
	m.SessionState = models.SessionStateFree
	m.AccountID = uuid.Nil
	m.SessionKey = uuid.Nil
}

func PingSession(device models.DeviceInfo, accountID uuid.UUID, sessionId uuid.UUID) error {
	dl := deviceList.GetCopy()
	if device.DeviceType == models.DeviceTypeIos {
		for _, d := range dl.iosDevices {
			if d.udid == device.Serial {
				return pingSession(&d.mux, &d.sessionState, accountID, sessionId)
			}
		}
	}

	return fmt.Errorf("device not found %s", device.Serial)
}

func pingSession(s *sync.Mutex, m *models.SessionState, accountId uuid.UUID, sessionId uuid.UUID) error {
	s.Lock()
	defer s.Unlock()
	if m.SessionState != models.SessionStateInUse {
		return fmt.Errorf("session not in use")
	}
	if m.AccountID != accountId {
		return fmt.Errorf("session not in use by this account")
	}
	if m.SessionKey != sessionId {
		return fmt.Errorf("ping for sessionId %s but actual session key is %s", sessionId, m.SessionKey)
	}
	m.SessionStateLastPing = time.Now()
	return nil

}

func CreateSession(device models.DeviceInfo, accountID uuid.UUID, sessionTimeout time.Duration) (chan error, error) {
	dl := deviceList.GetCopy()
	if device.DeviceType == models.DeviceTypeIos {
		for _, d := range dl.iosDevices {
			if d.udid == device.Serial {
				return createSession(&d.mux, &d.sessionState, accountID, sessionTimeout)
			}
		}
	}

	return nil, fmt.Errorf("device not found %s", device.Serial)
}

func createSession(mux *sync.Mutex, sessionState *models.SessionState, accountId uuid.UUID, sessionTimeout time.Duration) (chan error, error) {
	mux.Lock()
	defer mux.Unlock()
	if sessionState.SessionState == models.SessionStateFree || sessionState.SessionState == "" {
		sessionState.SessionState = models.SessionStateInUse
		sessionState.SessionKey = uuid.New()
		sessionState.AccountID = accountId
		sessionState.SessionStateLastPing = time.Now()
		sessionState.SessionTimeout = sessionTimeout
		sessionState.MetaInfo = map[string]interface{}{}
		sessionHandle := make(chan error)
		sessionHandles.Store(sessionState.SessionKey, sessionHandle)
		return sessionHandle, nil
	}
	if sessionState.SessionState == models.SessionStateInUse && sessionState.AccountID == accountId {
		return nil, fmt.Errorf("session already in use by this account")
	}
	if sessionState.SessionState == models.SessionStateInUse {
		return nil, fmt.Errorf("session already in use by another account")
	}
	return nil, fmt.Errorf("unknown session error")
}
