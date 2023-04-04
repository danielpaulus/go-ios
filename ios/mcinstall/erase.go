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
	log.Info("send flush request")
	_, err = check(conn.sendAndReceive(request("Flush")))
	if err != nil {
		return err
	}
	log.Info("get cloud config")
	config, err := check(conn.sendAndReceive(request("GetCloudConfiguration")))
	if err != nil {
		return err
	}
	log.Infof("config: %v", config)
	/*
		log.Info("get host identifier")
		hostId, err := check(conn.sendAndReceive(request("HelloHostIdentifier")))
		if err != nil {
			return err
		}
		log.Infof("identifier: %v", hostId)

		err = conn.EscalateUnsupervised()
		if err != nil {
			return err
		}

		log.Info("get host identifier")
		hostId, err = check(conn.sendAndReceive(request("HelloHostIdentifier")))
		if err != nil {
			return err
		}
		log.Infof("identifier: %v", hostId)
	*/
	log.Info("erase")
	eraseRequest := map[string]interface{}{
		"RequestType":      "EraseDevice",
		"PreserveDataPlan": 1,
	}
	eraseResp, err := check(conn.sendAndReceive(eraseRequest))
	if err != nil {
		return err
	}
	log.Infof("erase: %v", eraseResp)

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
