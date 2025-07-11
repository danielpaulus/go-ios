package notificationproxy

import (
	"bytes"
	"errors"
	"sync"
	"time"

	ios "github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
	"howett.net/plist"
)

const serviceName = "com.apple.mobile.notification_proxy"

type Connection struct {
	deviceConn          ios.DeviceConnectionInterface
	plistCodec          ios.PlistCodec
	alreadyObserving    map[string]interface{}
	notificationChannel chan string
	proxyDeathChannel   chan interface{}
	mux                 sync.Mutex
}

// Close sends a Shutdown command to notification proxy and closes the DeviceConnectionInterface
func (c *Connection) Close() {
	log.Debugf("shutting down %s", serviceName)
	request := notificationProxyRequest{Command: "Shutdown"}
	bytes, err := c.plistCodec.Encode(request)
	if err != nil {
		log.Debug(err)
	}
	err = c.deviceConn.Send(bytes)
	if err != nil {
		log.Debug(err)
	}
	c.deviceConn.Close()
}

func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}
	c := &Connection{
		deviceConn: deviceConn, plistCodec: ios.NewPlistCodec(), alreadyObserving: make(map[string]interface{}),
		notificationChannel: make(chan string), proxyDeathChannel: make(chan interface{}),
	}
	go read(c)
	return c, nil
}

// WaitUntilSpringboardStarted waits up to 5 minutes for springboard to restart
func WaitUntilSpringboardStarted(device ios.DeviceEntry) error {
	c, err := New(device)
	if err != nil {
		return err
	}
	defer c.Close()
	return c.Observe("com.apple.springboard.finishedstartup", time.Minute*5)
}

func read(c *Connection) error {
	log.Debug("notificationproxy start reading")
	reader := c.deviceConn.Reader()
	for {
		messageBytes, err := c.plistCodec.Decode(reader)
		if err != nil {
			return err
		}
		message, err := plistFromBytes(messageBytes)
		if err != nil {
			return err
		}
		log.Debugf("NotificationProxy: %+v", message)
		if command, ok := message["Command"].(string); ok {
			switch command {
			case "RelayNotification":
				c.notificationChannel <- message["Name"].(string)
			case "ProxyDeath":
				var signal interface{}
				c.proxyDeathChannel <- signal
			default:
				log.Debugf("Unknown message: %x", messageBytes)
			}
		} else {
			log.Debugf("Unknown message: %x", messageBytes)
		}
	}
}

func plistFromBytes(plistBytes []byte) (map[string]interface{}, error) {
	var message map[string]interface{}
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))

	err := decoder.Decode(&message)
	return message, err
}

// Observe waits for a notification up to a timeout. Currently supports only one listener per notification
func (c *Connection) Observe(notification string, timeout time.Duration) error {
	if yes := c.newNotification(notification); yes {
		err := c.startObserving(notification)
		if err != nil {
			return err
		}
	}
	for {
		select {
		case remoteNotification := <-c.notificationChannel:
			if notification == remoteNotification {
				return nil
			}
		case <-c.proxyDeathChannel:
			return errors.New("ProxyDeath")
		case <-time.After(timeout):
			return errors.New("Timeout")
		}
	}
}

func (c *Connection) startObserving(notification string) error {
	request := notificationProxyRequest{Command: "ObserveNotification", Name: notification}
	bytes, err := c.plistCodec.Encode(request)
	if err != nil {
		return err
	}
	return c.deviceConn.Send(bytes)
}

func (c *Connection) newNotification(notification string) bool {
	c.mux.Lock()
	defer c.mux.Unlock()
	if _, alreadyObserving := c.alreadyObserving[notification]; alreadyObserving {
		return false
	}
	var empty interface{}
	c.alreadyObserving[notification] = empty
	return true
}

type notificationProxyRequest struct {
	Command string
	Name    string `plist:"Name,omitempty"`
}
