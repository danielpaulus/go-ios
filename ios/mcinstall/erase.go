package mcinstall

import (
	"fmt"
	"io"

	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

// Erase tells a device to remove all apps and settings. You need to activate it afterwards.
// Be careful with this if you do not have a backup!
func Erase(device ios.DeviceEntry) error {
	conn, err := New(device)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Info("start erasing")
	log.Debug("send flush request")
	_, err = check(conn.sendAndReceive(request("Flush")))
	if err != nil {
		return err
	}
	log.Debug("get cloud config")
	config, err := check(conn.sendAndReceive(request("GetCloudConfiguration")))
	if err != nil {
		return err
	}
	log.Debugf("config: %v", config)

	log.Debug("send erase request")
	eraseRequest := map[string]interface{}{
		"RequestType":      "EraseDevice",
		"PreserveDataPlan": 1,
	}
	_, err = check(conn.sendAndReceive(eraseRequest))
	if err != nil && err != io.EOF {
		return err
	}
	log.Info("device should be rebooting now")
	return nil
}

func check(request map[string]interface{}, err error) (map[string]interface{}, error) {
	if err != nil {
		return map[string]interface{}{}, err
	}
	if !checkStatus(request) {
		return map[string]interface{}{}, fmt.Errorf("failed command: %v", request)
	}
	return request, nil
}
