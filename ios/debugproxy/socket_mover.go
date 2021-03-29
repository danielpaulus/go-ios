package debugproxy

import (
	"fmt"
	"log"
	"os"
)

const realSocketSuffix = ".real_socket"

func MoveSock(socket string) (string, error) {
	newLocation := socket + realSocketSuffix
	if fileExists(newLocation) {
		return "", fmt.Errorf("there is already a file named: %s please remove it or restore original usbmuxd before starting the proxy", newLocation)
	}
	log.Printf("Moving socket %s to %s", socket, newLocation)
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
	log.Printf("Deleting fake socket %s", socket)
	err := os.Remove(socket)
	if err != nil {
		log.Printf("Warn: failed deleting %s with error %e", socket, err)
	}
	log.Printf("Moving back socket %s to %s", newLocation, socket)
	err = os.Rename(newLocation, socket)
	return err
}
