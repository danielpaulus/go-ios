package instruments

import (
	"github.com/danielpaulus/go-ios/ios"
	dtx "github.com/danielpaulus/go-ios/ios/dtx_codec"
	log "github.com/sirupsen/logrus"
)

const serviceName string = "com.apple.instruments.remoteserver"
const serviceNameiOS14 string = "com.apple.instruments.remoteserver.DVTSecureSocketProxy"

type loggingDispatcher struct {
	conn *dtx.Connection
}

func (p loggingDispatcher) Dispatch(m dtx.Message) {
	dtx.SendAckIfNeeded(p.conn, m)
	log.Debug(m)
}

func connectInstruments(device ios.DeviceEntry) (*dtx.Connection, error) {
	dtxConn, err := dtx.NewConnection(device, serviceName)
	if err != nil {
		log.Debugf("Failed connecting to %s, trying %s", serviceName, serviceNameiOS14)
		dtxConn, err = dtx.NewConnection(device, serviceNameiOS14)
		if err != nil {
			return nil, err
		}
	}
	return dtxConn, nil
}
