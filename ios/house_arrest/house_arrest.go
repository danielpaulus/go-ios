package house_arrest

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/danielpaulus/go-ios/ios/afc"
	"github.com/danielpaulus/go-ios/ios/golog"
	"howett.net/plist"

	"github.com/danielpaulus/go-ios/ios"
)

const serviceName = "com.apple.mobile.house_arrest"

const logModule = "go-ios/house_arrest"

const (
	commandVendContainer = "VendContainer"
	commandVendDocuments = "VendDocuments"
)

func New(device ios.DeviceEntry, bundleID string) (*afc.Client, error) {
	return connect(func() (ios.DeviceConnectionInterface, error) {
		return ios.ConnectToService(device, serviceName)
	}, bundleID, device.Properties.SerialNumber)
}

// connect vends the full app container. If the device denies that (it does for
// apps not installed with a developer profile, e.g. App Store apps, reporting
// InstallationLookupFailed), it falls back to documents-only access on a fresh
// connection, since house_arrest accepts only one command per connection.
func connect(dial func() (ios.DeviceConnectionInterface, error), bundleID string, udid string) (*afc.Client, error) {
	deviceConn, err := dial()
	if err != nil {
		return nil, err
	}
	containerErr := vend(deviceConn, commandVendContainer, bundleID)
	if containerErr == nil {
		return afc.NewFromConn(deviceConn), nil
	}
	deviceConn.Close()
	golog.Info("VendContainer was denied, falling back to VendDocuments", "module", logModule, "udid", udid, "bundleID", bundleID, "err", containerErr)
	deviceConn, err = dial()
	if err != nil {
		return nil, err
	}
	documentsErr := vend(deviceConn, commandVendDocuments, bundleID)
	if documentsErr == nil {
		return afc.NewFromConn(deviceConn), nil
	}
	deviceConn.Close()
	return nil, fmt.Errorf("house_arrest: VendContainer failed (%v) and the VendDocuments fallback failed (%v). Full container access requires an app installed with a developer profile; documents access requires the app to set UIFileSharingEnabled in its Info.plist", containerErr, documentsErr)
}

func vend(deviceConn ios.DeviceConnectionInterface, command string, bundleID string) error {
	plistCodec := ios.NewPlistCodec()
	request := map[string]interface{}{"Command": command, "Identifier": bundleID}
	msg, err := plistCodec.Encode(request)
	if err != nil {
		return fmt.Errorf("%s Encoding cannot fail unless the encoder is broken: %v", command, err)
	}
	err = deviceConn.Send(msg)
	if err != nil {
		return err
	}
	reader := deviceConn.Reader()
	response, err := plistCodec.Decode(reader)
	if err != nil {
		return err
	}
	return checkResponse(response)
}

func checkResponse(vendContainerResponseBytes []byte) error {
	response, err := plistFromBytes(vendContainerResponseBytes)
	if err != nil {
		return err
	}
	if "Complete" == response.Status {
		return nil
	}
	if response.Error != "" {
		return errors.New(response.Error)
	}
	return errors.New("unknown error during vendcontainer")
}

func plistFromBytes(plistBytes []byte) (vendContainerResponse, error) {
	var vendResponse vendContainerResponse
	decoder := plist.NewDecoder(bytes.NewReader(plistBytes))

	err := decoder.Decode(&vendResponse)
	if err != nil {
		return vendResponse, err
	}
	return vendResponse, nil
}

type vendContainerResponse struct {
	Status string
	Error  string
}
