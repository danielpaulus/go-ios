package imagemounter

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

const serviceName string = "com.apple.mobile.mobile_image_mounter"

// developerDiskImageMounter to mobile image mounter
type developerDiskImageMounter struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
	version    *semver.Version
	plistRw    ios.PlistCodecReadWriter
}

type ImageMounter interface {
	ListImages() ([][]byte, error)
	MountImage(imagePath string) error
}

// New returns a new mobile image mounter developerDiskImageMounter for the given DeviceID and Udid
func New(device ios.DeviceEntry) (ImageMounter, error) {
	version, err := ios.GetProductVersion(device)
	if err != nil {
		return nil, err
	}
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return nil, err
	}
	return &developerDiskImageMounter{
		deviceConn: deviceConn,
		plistCodec: ios.NewPlistCodec(),
		version:    version,
		plistRw:    ios.NewPlistCodecReadWriter(deviceConn.Reader(), deviceConn.Writer()),
	}, nil
}

// ListImages returns a list with signatures of installed developer images
func (conn *developerDiskImageMounter) ListImages() ([][]byte, error) {
	return listImages(conn.plistRw, "Developer", conn.version)
}

// MountImage installs a .dmg image from imagePath after checking that it is present and valid.
func (conn *developerDiskImageMounter) MountImage(imagePath string) error {
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

func (conn *developerDiskImageMounter) mountImage(signatureBytes []byte) error {
	req := map[string]interface{}{
		"Command":        "MountImage",
		"ImageSignature": signatureBytes,
		"ImageType":      "Developer",
	}
	log.Debugf("sending: %+v", req)
	err := conn.plistRw.Write(req)
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

// Close closes the underlying UsbMuxConnection
func (conn *developerDiskImageMounter) Close() {
	conn.deviceConn.Close()
}

func (conn *developerDiskImageMounter) sendUploadRequest(signatureBytes []byte, fileSize uint64) error {
	req := map[string]interface{}{
		"Command":        "ReceiveBytes",
		"ImageSignature": signatureBytes,
		"ImageSize":      fileSize,
		"ImageType":      "Developer",
	}
	log.Debugf("sending: %+v", req)
	err := conn.plistRw.Write(req)
	if err != nil {
		return err
	}
	return nil
}

func (conn *developerDiskImageMounter) checkUploadResponse() error {
	var plist map[string]interface{}
	err := conn.plistRw.Read(&plist)
	if err != nil {
		return err
	}
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

func (conn *developerDiskImageMounter) waitForUploadComplete() error {
	var plist map[string]interface{}
	err := conn.plistRw.Read(&plist)
	if err != nil {
		return err
	}
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

func (conn *developerDiskImageMounter) hangUp() error {
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

func MountImage(device ios.DeviceEntry, path string) error {
	conn, err := New(device)
	if err != nil {
		return fmt.Errorf("failed connecting to image mounter: %v", err)
	}

	signatures, err := conn.ListImages()
	if err != nil {
		return fmt.Errorf("failed getting image list: %v", err)
	}
	if len(signatures) != 0 {
		log.Warn("there is already a developer image mounted, reboot the device if you want to remove it. aborting.")
		return nil
	}
	return conn.MountImage(path)
}

func listImages(prw ios.PlistCodecReadWriter, imageType string, v *semver.Version) ([][]byte, error) {
	err := prw.Write(map[string]interface{}{
		"Command":   "LookupImage",
		"ImageType": imageType,
	})
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	err = prw.Read(&resp)
	if err != nil {
		return nil, err
	}
	deviceError, ok := resp["Error"]
	if ok {
		return nil, fmt.Errorf("device error: %v", deviceError)
	}

	signatures, ok := resp["ImageSignature"]
	if !ok {
		if v.LessThan(ios.IOS14()) {
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
