package proxy_utils

import (
	"log"
	"os"
)

func MoveSock(socket string) (string, error) {
	newLocation := socket + ".real"
	log.Printf("Moving socket %s to %s", socket, newLocation)
	err := os.Rename(socket, newLocation)
	return newLocation, err
}

func MoveBack(socket string) error {
	newLocation := socket + ".real"
	log.Printf("Deleting fake socket %s", socket)
	err := os.Remove(socket)
	if err != nil {
		log.Printf("Warn: failed deleting %s with error %e", socket, err)
	}
	log.Printf("Moving back socket %s to %s", newLocation, socket)
	err = os.Rename(newLocation, socket)
	return err
}
