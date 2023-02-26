//go:build !fast
// +build !fast

package syslog_test

import (
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/syslog"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSyslog(t *testing.T) {
	device, err := ios.GetDevice("")
	if err != nil {
		t.Error(err)
		return
	}
	syslogConnection, err := syslog.New(device)
	if err != nil {
		t.Error(err)
		return
	}
	defer syslogConnection.Close()
	for i := 0; i < 5; i++ {
		msg, err := syslogConnection.ReadLogMessage()
		log.Debug(msg)
		if assert.NoError(t, err) {
			assert.Greater(t, len(msg), 0)
		}
	}
}
