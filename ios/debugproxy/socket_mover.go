package debugproxy

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

var realSocketSuffix = fmt.Sprintf(".%s.real_socket", uuid.New().String())

func MoveSock(socket string) (string, error) {
	newLocation := socket + realSocketSuffix
	if fileExists(newLocation) {
		return "", fmt.Errorf("there is already a file named: %s please remove it or restore original usbmuxd before starting the proxy", newLocation)
	}
	log.Infof("Moving socket %s to %s", socket, newLocation)
	err := os.Rename(socket, newLocation)
	return newLocation, err
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func MoveBack(socket string) error {
	newLocation := socket + realSocketSuffix
	log.Infof("checking if '%s' exists", newLocation)
	if !fileExists(newLocation) {
		log.Infof("'%s' does not exist, doing nothing", newLocation)
		return nil
	}
	log.Infof("found '%s', deleting '%s'", newLocation, socket)

	log.Infof("Deleting fake socket %s", socket)
	err := os.Remove(socket)
	if err != nil {
		log.Warnf("Failed deleting %s with error %e", socket, err)
	}
	log.Infof("Moving back socket %s to %s", newLocation, socket)
	err = os.Rename(newLocation, socket)
	return err
}
