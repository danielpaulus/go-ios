package mcinstall

import (
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

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
	eraseResp, err := check(conn.sendAndReceive(eraseRequest))
	if err != nil {
		return err
	}
	log.Infof("erase resp: %v", eraseResp)

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
