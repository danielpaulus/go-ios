package imagemounter

import (
	"errors"
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strings"
)

const serviceName string = "com.apple.mobile.mobile_image_mounter"

//Connection to mobile image mounter
type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
	version    *semver.Version
}

//New returns a new mobile image mounter Connection for the given DeviceID and Udid
func New(device ios.DeviceEntry) (*Connection, error) {
	version, err := ios.GetProductVersion(device)
	if err != nil {
		return nil, err
	}
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}
	return &Connection{
		deviceConn: deviceConn,
		plistCodec: ios.NewPlistCodec(),
		version:    version,
	}, nil
}

//ListImages returns a list with signatures of installed developer images
func (conn *Connection) ListImages() ([][]byte, error) {
	req := map[string]interface{}{
		"Command":   "LookupImage",
		"ImageType": "Developer",
	}
	bytes, err := conn.plistCodec.Encode(req)
	if err != nil {
		return nil, err
	}

	err = conn.deviceConn.Send(bytes)
	if err != nil {
		return nil, err
	}

	bytes, err = conn.plistCodec.Decode(conn.deviceConn.Reader())

	resp, err := ios.ParsePlist(bytes)
	if err != nil {
		return nil, err
	}
	deviceError, ok := resp["Error"]
	if ok {
		return nil, fmt.Errorf("device error: %v", deviceError)
	}

	signatures, ok := resp["ImageSignature"]
	if !ok {
		if conn.version.LessThan(semver.MustParse("14.0")) {
			return [][]byte{}, nil
		}
		return nil, fmt.Errorf("invalid response: %+v", resp)
	}

	array, ok := signatures.([]interface{})
	result := make([][]byte, len(array))
	for i, intf := range array {
		bytes, ok := intf.([]byte)
		if !ok {
			return nil, fmt.Errorf("could not convert %+v to byte slice", intf)
		}
		result[i] = bytes
	}
	return result, nil
}

func (conn *Connection) MountImage(imagePath string) error {
	signatureBytes, imageSize, err := validatePathAndLoadSignature(imagePath)
	if err != nil {
		return err
	}
	err = conn.sendUploadRequest(signatureBytes, uint64(imageSize))
	if err != nil {
		return err
	}
	err = conn.checkUploadResponse()
	if err != nil {
		return err
	}
	imageFile, err := os.Open(imagePath)
	if err != nil {
		return err
	}
	defer imageFile.Close()
	n, err := io.Copy(conn.deviceConn.Writer(), imageFile)
	log.Debugf("%d bytes written", n)
	if err != nil {
		return err
	}
	err = conn.waitForUploadComplete()
	if err != nil {
		return err
	}
	err = conn.mountImage(signatureBytes)
	if err != nil {
		return err
	}

	return conn.hangUp()
}

func (conn *Connection) mountImage(signatureBytes []byte) error {
	req := map[string]interface{}{
		"Command":        "MountImage",
		"ImageSignature": signatureBytes,
		"ImageType":      "Developer",
	}
	log.Debugf("sending: %+v", req)
	bytes, err := conn.plistCodec.Encode(req)
	if err != nil {
		return err
	}

	err = conn.deviceConn.Send(bytes)
	if err != nil {
		return err
	}
	return nil
}

func validatePathAndLoadSignature(imagePath string) ([]byte, int64, error) {
	imageFile, err := os.Open(imagePath)
	if err != nil {
		return []byte{}, 0, err
	}
	defer imageFile.Close()

	// Get the file information
	info, err := imageFile.Stat()
	if err != nil {
		return []byte{}, 0, err
	}
	if info.IsDir() {
		return []byte{}, 0, errors.New("provided path is a directory")
	}

	if !strings.HasSuffix(imagePath, ".dmg") {
		return []byte{}, 0, errors.New("provided path is not a dmg file")
	}

	signatureFile, err := os.Open(imagePath + ".signature")
	if err != nil {
		return []byte{}, 0, err
	}
	defer imageFile.Close()
	signatureBytes, err := io.ReadAll(signatureFile)
	if err != nil {
		return []byte{}, 0, err
	}
	return signatureBytes, info.Size(), nil
}

//Close closes the underlying UsbMuxConnection
func (conn *Connection) Close() {
	conn.deviceConn.Close()
}

func (conn *Connection) sendUploadRequest(signatureBytes []byte, fileSize uint64) error {
	req := map[string]interface{}{
		"Command":        "ReceiveBytes",
		"ImageSignature": signatureBytes,
		"ImageSize":      fileSize,
		"ImageType":      "Developer",
	}
	log.Debugf("sending: %+v", req)
	bytes, err := conn.plistCodec.Encode(req)
	if err != nil {
		return err
	}

	err = conn.deviceConn.Send(bytes)
	if err != nil {
		return err
	}
	return nil
}

func (conn *Connection) checkUploadResponse() error {
	msg, _ := conn.plistCodec.Decode(conn.deviceConn.Reader())
	plist, _ := ios.ParsePlist(msg)
	log.Debugf("upload response: %+v", plist)
	status, ok := plist["Status"]
	if !ok {
		return fmt.Errorf("unexpected response: %+v", plist)
	}
	if "ReceiveBytesAck" != status {
		return fmt.Errorf("unexpected response: %+v", plist)
	}
	return nil
}

func (conn *Connection) waitForUploadComplete() error {
	msg, _ := conn.plistCodec.Decode(conn.deviceConn.Reader())
	plist, _ := ios.ParsePlist(msg)
	log.Debugf("received complete: %+v", plist)
	status, ok := plist["Status"]
	if !ok {
		return fmt.Errorf("unexpected response: %+v", plist)
	}
	if "Complete" != status {
		return fmt.Errorf("unexpected response: %+v", plist)
	}
	return nil
}

func (conn *Connection) hangUp() error {
	req := map[string]interface{}{
		"Command": "Hangup",
	}
	log.Debugf("sending: %+v", req)
	bytes, err := conn.plistCodec.Encode(req)
	if err != nil {
		return err
	}

	err = conn.deviceConn.Send(bytes)
	if err != nil {
		return err
	}
	return nil
}
