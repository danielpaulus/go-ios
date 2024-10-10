package api

import (
	"context"
	"net/http"
	"os"
	"sync"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/testmanagerd"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

type WdaConfig struct {
	BundleID     string                 `json:"bundleId" binding:"required"`
	TestbundleID string                 `json:"testBundleId" binding:"required"`
	XCTestConfig string                 `json:"xcTestConfig" binding:"required"`
	Args         []string               `json:"args"`
	Env          map[string]interface{} `json:"env"`
}

type WdaSessionKey struct {
	udid      string
	sessionID string
}

type WdaSession struct {
	Config    WdaConfig `json:"config" binding:"required"`
	SessionId string    `json:"sessionId" binding:"required"`
	Udid      string    `json:"udid" binding:"required"`
	stopWda   context.CancelFunc
}

func (session *WdaSession) Write(p []byte) (n int, err error) {
	log.
		WithField("udid", session.Udid).
		WithField("sessionId", session.SessionId).
		Debugf("WDA_LOG %s", p)

	return len(p), nil
}

var globalSessions = sync.Map{}

func CreateWdaSession(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	log.
		WithField("udid", device.Properties.SerialNumber).
		Debugf("Creating WDA session")

	var config WdaConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionKey := WdaSessionKey{
		udid:      device.Properties.SerialNumber,
		sessionID: uuid.New().String(),
	}

	wdaCtx, stopWda := context.WithCancel(context.Background())

	session := WdaSession{
		Udid:      sessionKey.udid,
		SessionId: sessionKey.sessionID,
		Config:    config,
		stopWda:   stopWda,
	}

	go func() {
		_, err := testmanagerd.RunXCUIWithBundleIdsCtx(wdaCtx, config.BundleID, config.TestbundleID, config.XCTestConfig, device, config.Args, config.Env, nil, nil, testmanagerd.NewTestListener(&session, &session, os.TempDir()), false)
		if err != nil {
			log.
				WithField("udid", sessionKey.udid).
				WithField("sessionId", sessionKey.sessionID).
				WithError(err).
				Error("Failed running WDA")
		}

		stopWda()
		globalSessions.Delete(sessionKey)

		log.
			WithField("udid", sessionKey.udid).
			WithField("sessionId", sessionKey.sessionID).
			Debug("Deleted WDA session")
	}()

	globalSessions.Store(sessionKey, session)

	log.
		WithField("udid", sessionKey.udid).
		WithField("sessionId", sessionKey.sessionID).
		Debugf("Requested to start WDA session")

	c.JSON(http.StatusOK, session)
}

func ReadWdaSession(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	sessionKey := WdaSessionKey{
		udid:      device.Properties.SerialNumber,
		sessionID: sessionID,
	}

	session, loaded := globalSessions.Load(sessionKey)
	if !loaded {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, session)
}

func DeleteWdaSession(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	sessionKey := WdaSessionKey{
		udid:      device.Properties.SerialNumber,
		sessionID: sessionID,
	}

	session, loaded := globalSessions.Load(sessionKey)
	if !loaded {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	wdaSession, ok := session.(WdaSession)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cast session"})
		return
	}
	wdaSession.stopWda()

	log.
		WithField("udid", sessionKey.udid).
		WithField("sessionId", sessionKey.sessionID).
		Debug("Requested to stop WDA")

	c.JSON(http.StatusOK, session)
}
