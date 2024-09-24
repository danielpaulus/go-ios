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

// DeveloperDiskImageMounter to mobile image mounter
type DeveloperDiskImageMounter struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
	version    *semver.Version
	plistRw    ios.PlistCodecReadWriter
}

// ImageMounter mounts developer disk images to an iOS device, and give a list of already mounted images
type ImageMounter interface {
	ListImages() ([][]byte, error)
	MountImage(imagePath string) error
	UnmountImage() error
	io.Closer
}

// NewDeveloperDiskImageMounter returns a new mobile image mounter DeveloperDiskImageMounter for the given device
func NewDeveloperDiskImageMounter(device ios.DeviceEntry, version *semver.Version) (*DeveloperDiskImageMounter, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return nil, err
	}
	return &DeveloperDiskImageMounter{
		deviceConn: deviceConn,
		plistCodec: ios.NewPlistCodec(),
		version:    version,
		plistRw:    ios.NewPlistCodecReadWriter(deviceConn.Reader(), deviceConn.Writer()),
	}, nil
}

// NewImageMounter creates a new ImageMounter depending on the version of the given device.
// For iOS 17+ devices a PersonalizedDeveloperDiskImageMounter is created, and for all other devices
// a DeveloperDiskImageMounter gets created
func NewImageMounter(device ios.DeviceEntry) (ImageMounter, error) {
	version, err := ios.GetProductVersion(device)
	if err != nil {
		return nil, fmt.Errorf("NewImageMounter: failed to get device version. %w", err)
	}
	if version.Major() < 17 {
		return NewDeveloperDiskImageMounter(device, version)
	} else {
		return NewPersonalizedDeveloperDiskImageMounter(device, version)
	}
}

// ListImages returns a list with signatures of installed developer images
func (conn *DeveloperDiskImageMounter) ListImages() ([][]byte, error) {
	return listImages(conn.plistRw, "Developer", conn.version)
}

// MountImage installs a .dmg image from imagePath after checking that it is present and valid.
func (conn *DeveloperDiskImageMounter) MountImage(imagePath string) error {
	signatureBytes, imageSize, err := validatePathAndLoadSignature(imagePath)
	if err != nil {
		return err
	}
	err = sendUploadRequest(conn.plistRw, "Developer", signatureBytes, uint64(imageSize))
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
	err = waitForUploadComplete(conn.plistRw)
	if err != nil {
		return err
	}
	err = conn.mountImage(signatureBytes)
	if err != nil {
		return err
	}

	return hangUp(conn.plistRw)
}

func (conn *DeveloperDiskImageMounter) UnmountImage() error {
	req := map[string]interface{}{
		"Command":   "UnmountImage",
		"MountPath": "/Developer",
	}
	log.Debugf("sending: %+v", req)
	err := conn.plistRw.Write(req)
	if err != nil {
		return err
	}
	return nil
}

func (conn *DeveloperDiskImageMounter) mountImage(signatureBytes []byte) error {
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
func (conn *DeveloperDiskImageMounter) Close() error {
	return conn.deviceConn.Close()
}

func waitForUploadComplete(plistRw ios.PlistCodecReadWriter) error {
	var plist map[string]interface{}
	err := plistRw.Read(&plist)
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

func hangUp(plistRw ios.PlistCodecReadWriter) error {
	req := map[string]interface{}{
		"Command": "Hangup",
	}
	log.Debugf("sending: %+v", req)
	return plistRw.Write(req)
}

func MountImage(device ios.DeviceEntry, path string) error {
	conn, err := NewImageMounter(device)
	if err != nil {
		return fmt.Errorf("failed connecting to image mounter: %v", err)
	}
	defer conn.Close()

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

func UnmountImage(device ios.DeviceEntry) error {
	conn, err := NewImageMounter(device)
	if err != nil {
		return fmt.Errorf("failed connecting to image mounter: %v", err)
	}
	defer conn.Close()

	return conn.UnmountImage()
}

func listImages(prw ios.PlistCodecReadWriter, imageType string, v *semver.Version) ([][]byte, error) {
	err := prw.Write(map[string]interface{}{
		"Command":   "LookupImage",
		"ImageType": imageType,
	})
	if err != nil {
		return nil, fmt.Errorf("listImages: failed to write command 'LookupImage': %w", err)
	}

	var resp map[string]interface{}
	err = prw.Read(&resp)
	if err != nil {
		return nil, fmt.Errorf("listImages: failed to read response for 'LookupImage': %w", err)
	}
	deviceError, ok := resp["Error"]
	if ok {
		return nil, fmt.Errorf("listImages: device responded with error: %v", deviceError)
	}

	signatures, ok := resp["ImageSignature"]
	if !ok {
		if v.LessThan(ios.IOS14()) {
			return [][]byte{}, nil
		}
		return nil, fmt.Errorf("listImages: invalid response: %+v", resp)
	}

	array, ok := signatures.([]interface{})
	result := make([][]byte, len(array))
	for i, intf := range array {
		bytes, ok := intf.([]byte)
		if !ok {
			return nil, fmt.Errorf("listImages: could not convert %+v to byte slice", intf)
		}
		result[i] = bytes
	}
	return result, nil
}

func sendUploadRequest(plistRw ios.PlistCodecReadWriter, imageType string, signatureBytes []byte, fileSize uint64) error {
	req := map[string]interface{}{
		"Command":        "ReceiveBytes",
		"ImageSignature": signatureBytes,
		"ImageSize":      fileSize,
		"ImageType":      imageType,
	}
	log.Debugf("sending: %+v", req)
	err := plistRw.Write(req)
	if err != nil {
		return fmt.Errorf("sendUploadRequest: failed to write command 'ReceiveBytes': %w", err)
	}

	var plist map[string]interface{}
	err = plistRw.Read(&plist)
	if err != nil {
		return fmt.Errorf("sendUploadRequest: failed to read response for 'ReceiveBytes': %w", err)
	}
	log.Debugf("upload response: %+v", plist)
	status, ok := plist["Status"]
	if !ok {
		return fmt.Errorf("sendUploadRequest: unexpected response: %+v", plist)
	}
	if "ReceiveBytesAck" != status {
		return fmt.Errorf("sendUploadRequest: unexpected status: %+v", plist)
	}
	return nil
}

// Check if developer mode is enabled through the mobile_image_mounter service
func IsDevModeEnabled(device ios.DeviceEntry) (bool, error) {
	conn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return false, fmt.Errorf("IsDevModeEnabled: failed connecting to image mounter service with err: %w", err)
	}
	defer conn.Close()

	reader := conn.Reader()
	request := map[string]interface{}{"Command": "QueryDeveloperModeStatus"}

	plistCodec := ios.NewPlistCodec()

	bytes, err := plistCodec.Encode(request)
	if err != nil {
		return false, fmt.Errorf("IsDevModeEnabled: failed encoding request to service with err: %w", err)
	}

	err = conn.Send(bytes)
	if err != nil {
		return false, fmt.Errorf("IsDevModeEnabled: failed sending request to service with err: %w", err)
	}

	responseBytes, err := plistCodec.Decode(reader)
	if err != nil {
		return false, fmt.Errorf("IsDevModeEnabled: failed decoding service response with err: %w", err)
	}

	plist, err := ios.ParsePlist(responseBytes)
	if err != nil {
		return false, fmt.Errorf("IsDevModeEnabled: failed parsing service response with err: %w", err)
	}

	if val, ok := plist["DeveloperModeStatus"]; ok {
		if assertedVal, ok := val.(bool); ok {
			return assertedVal, nil
		}
		return false, fmt.Errorf("IsDevModeEnabled: failed type assertion on DeveloperModeStatus value from service response")
	}

	return false, nil
}
