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
// Transport failures abort without the fallback: they are not a denial by the
// device, and retrying would mask them or silently degrade a working container
// connection to documents-only access.
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
	var denied vendError
	if !errors.As(containerErr, &denied) {
		return nil, containerErr
	}
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
	return checkResponse(response, command)
}

// vendError is an error the device reported in a house_arrest response (e.g.
// InstallationLookupFailed), as opposed to a transport failure.
type vendError struct {
	message string
}

func (e vendError) Error() string {
	return e.message
}

func checkResponse(vendResponseBytes []byte, command string) error {
	response, err := plistFromBytes(vendResponseBytes)
	if err != nil {
		return err
	}
	if "Complete" == response.Status {
		return nil
	}
	if response.Error != "" {
		return vendError{message: response.Error}
	}
	return vendError{message: fmt.Sprintf("unknown error during %s", command)}
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
