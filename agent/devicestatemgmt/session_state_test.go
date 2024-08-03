package devicestatemgmt_test

import (
	"os"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/agent/devicestatemgmt"
	"github.com/danielpaulus/go-ios/agent/models"
	"github.com/danielpaulus/go-ios/agent/utils"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func TestSessions(t *testing.T) {

	utils.SESSION_TIMEOUT_CHECK_INTERVAL_SEC = time.Millisecond
	closeFunc, err := models.InitDb("test.db")
	if err != nil {
		log.Fatalf("error creating local db: %v", err)
	}
	defer func() { os.Remove("test.db") }()
	defer closeFunc()
	deviceList := CreateDeviceList()
	managerCloseFunc := devicestatemgmt.StartDeviceStateManager(deviceList, false)
	defer managerCloseFunc()
	d := deviceList.GetCurrentInfo()
	accountID := uuid.New()
	sessionHandle, err := devicestatemgmt.CreateSession(d[0], accountID, time.Minute)
	if err != nil {
		t.Errorf("error creating session: %v", err)
		return
	}

	_, err = devicestatemgmt.CreateSession(d[0], accountID, time.Minute)
	if err == nil {
		t.Errorf("session creation should have failed")
		return
	}

	_, err = devicestatemgmt.CreateSession(d[0], uuid.New(), time.Minute)
	if err == nil {
		t.Errorf("session creation should have failed")
		return
	}
	d = deviceList.GetCurrentInfo()
	time1 := d[0].SessionState.SessionStateLastPing
	err = devicestatemgmt.PingSession(d[0], accountID, d[0].SessionState.SessionKey)
	if err != nil {
		t.Errorf("session ping failed with: %v", err)
		return
	}
	d = deviceList.GetCurrentInfo()
	time2 := d[0].SessionState.SessionStateLastPing
	if !time2.After(time1) {
		t.Errorf("session ping did not update last ping time")
		return
	}

	err = devicestatemgmt.PingSession(d[1], accountID, d[0].SessionState.SessionKey)
	if err == nil {
		t.Errorf("session ping should have failed")
		return
	}

	err = devicestatemgmt.PingSession(d[0], uuid.New(), d[0].SessionState.SessionKey)
	if err == nil {
		t.Errorf("session ping should have failed")
		return
	}

	err = devicestatemgmt.PingSession(d[0], accountID, uuid.New())
	if err == nil {
		t.Errorf("session ping should have failed")
		return
	}

	err = devicestatemgmt.TerminateSession(d[1], uuid.New(), d[0].SessionState.SessionKey, false)
	if err == nil {
		t.Errorf("session creation should have failed")
		return
	}
	err = devicestatemgmt.TerminateSession(d[0], uuid.New(), d[0].SessionState.SessionKey, false)
	if err == nil {
		t.Errorf("session creation should have failed")
		return
	}
	err = devicestatemgmt.TerminateSession(d[0], accountID, uuid.New(), false)
	if err == nil {
		t.Errorf("session creation should have failed")
		return
	}

	err = devicestatemgmt.TerminateSession(d[0], accountID, d[0].SessionState.SessionKey, false)
	if err != nil {
		t.Errorf("session creation should have succeeded but failed with: %v", err)
		return
	}
	select {
	case <-sessionHandle:
	case <-time.After(time.Millisecond * 10):
		t.Errorf("timed out waiting for session handle being closed")
	}

	sessionHandle, err = devicestatemgmt.CreateSession(d[1], accountID, time.Nanosecond)
	select {
	case <-sessionHandle:
	case <-time.After(time.Millisecond * 10):
		t.Errorf("session should have timed out")
	}

	sessionHandle, err = devicestatemgmt.CreateSession(d[1], accountID, time.Hour)
	d = deviceList.GetCurrentInfo()
	err = devicestatemgmt.TerminateSession(d[1], uuid.Nil, d[1].SessionState.SessionKey, true)
	if err != nil {
		t.Errorf("session should have been terminated through root but failed with: %v", err)
		return
	}
	select {
	case <-sessionHandle:
	case <-time.After(time.Millisecond * 10):
		t.Errorf("timed out waiting for session handle being closed")
	}
}

func CreateDeviceList() *devicestatemgmt.DeviceList {
	return devicestatemgmt.NewDeviceList([]models.DeviceInfo{
		{
			Serial:     "123",
			Name:       "test",
			DeviceType: models.DeviceTypeIos,
		},
		{
			Serial:     "456",
			Name:       "test",
			DeviceType: models.DeviceTypeAndroid,
		},
	})
}
