package imagemounter

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

// PersonalizedDeveloperDiskImageMounter allows mounting personalized developer disk images
// that are used starting with iOS 17.
// For personalized developer disk images a nonce gets queried from the device and needs to
// be signed by Apple to be able to mount the developer disk
type PersonalizedDeveloperDiskImageMounter struct {
	deviceConn ios.DeviceConnectionInterface
	plistRw    ios.PlistCodecReadWriter
	version    *semver.Version
	tss        tssClient
	ecid       uint64
}

// NewPersonalizedDeveloperDiskImageMounter creates a PersonalizedDeveloperDiskImageMounter for the device entry
func NewPersonalizedDeveloperDiskImageMounter(entry ios.DeviceEntry, version *semver.Version) (PersonalizedDeveloperDiskImageMounter, error) {
	values, err := ios.GetValuesPlist(entry)
	if err != nil {
		return PersonalizedDeveloperDiskImageMounter{}, fmt.Errorf("NewPersonalizedDeveloperDiskImageMounter: could not read lockdown values: %w", err)
	}
	var ecid uint64
	if e, ok := values["UniqueChipID"].(uint64); ok {
		ecid = e
	} else {
		return PersonalizedDeveloperDiskImageMounter{}, fmt.Errorf("could not get ECID from device")
	}
	deviceConn, err := ios.ConnectToService(entry, serviceName)
	if err != nil {
		return PersonalizedDeveloperDiskImageMounter{}, err
	}
	return PersonalizedDeveloperDiskImageMounter{
		deviceConn: deviceConn,
		plistRw:    ios.NewPlistCodecReadWriter(deviceConn.Reader(), deviceConn.Writer()),
		version:    version,
		tss:        newTssClient(),
		ecid:       ecid,
	}, nil
}

// Close closes the connection to the image mounter service
func (p PersonalizedDeveloperDiskImageMounter) Close() error {
	return p.deviceConn.Close()
}

// ListImages provides a list of signatures of the mounted personalized developer disk images
func (p PersonalizedDeveloperDiskImageMounter) ListImages() ([][]byte, error) {
	return listImages(p.plistRw, "Personalized", p.version)
}

// MountImage mounts the personalized developer disk image present at imagePath.
// imagePath needs to point to the 'Restore' directory of the personalized developer disk image.
//
// MountImage gets device identifiers and a nonce from the device first, which needs to be signed by Apple
// and after that the developer disk image is sent to the device with this signature to be able to mount it.
func (p PersonalizedDeveloperDiskImageMounter) MountImage(imagePath string) error {
	manifest, err := loadBuildManifest(path.Join(imagePath, "BuildManifest.plist"))
	if err != nil {
		return fmt.Errorf("MountImage: failed to load build manifest: %w", err)
	}

	identifiers, err := p.queryIdentifiers()
	if err != nil {
		return fmt.Errorf("MountImage: failed to query personalization identifiers: %w", err)
	}
	nonce, err := p.queryPersonalizedImageNonce()
	if err != nil {
		return fmt.Errorf("MountImage: failed to get nonce: %w", err)
	}

	identity, err := manifest.findIdentity(identifiers)
	if err != nil {
		return fmt.Errorf("MountImage: could not find identity for identifiers %+v: %w", identifiers, err)
	}

	signature, err := p.tss.getSignature(identity, identifiers, nonce, p.ecid)
	if err != nil {
		return fmt.Errorf("MountImage: failed to get signature from Apple: %w", err)
	}

	dmgPath := path.Join(imagePath, identity.Manifest.PersonalizedDmg.Info.Path)

	imageSize, err := getFileSize(dmgPath)

	err = sendUploadRequest(p.plistRw, "Personalized", signature, imageSize)
	if err != nil {
		return fmt.Errorf("MountImage: failed to send upload request for image: %w", err)
	}
	imageFile, err := os.Open(dmgPath)
	if err != nil {
		return fmt.Errorf("MountImage: failed to open developer disk dmg file '%s': %w", dmgPath, err)
	}
	defer imageFile.Close()
	n, err := io.Copy(p.deviceConn.Writer(), imageFile)
	log.Debugf("%d bytes written", n)
	if err != nil {
		return fmt.Errorf("MountImage: could not copy developer disk image to the device: %w", err)
	}
	err = waitForUploadComplete(p.plistRw)
	if err != nil {
		return err
	}

	trustCache, err := os.ReadFile(path.Join(imagePath, identity.Manifest.LoadableTrustCache.Info.Path))
	if err != nil {
		return fmt.Errorf("MountImage: could not load trust-cache. %w", err)
	}

	err = p.mountPersonalizedImage(signature, trustCache)
	if err != nil {
		return fmt.Errorf("MountImage: mount command failed: %w", err)
	}

	err = hangUp(p.plistRw)
	if err != nil {
		return fmt.Errorf("MountImage: HangUp command failed: %w", err)
	}
	return nil
}

func (p PersonalizedDeveloperDiskImageMounter) UnmountImage() error {
	req := map[string]interface{}{
		"Command":   "UnmountImage",
		"MountPath": "/System/Developer",
	}
	log.Debugf("sending: %+v", req)
	err := p.plistRw.Write(req)
	if err != nil {
		return err
	}
	return nil
}

func (p PersonalizedDeveloperDiskImageMounter) queryPersonalizedImageNonce() ([]byte, error) {
	err := p.plistRw.Write(map[string]interface{}{
		"Command":               "QueryNonce",
		"HostProcessName":       "CoreDeviceService",
		"PersonalizedImageType": "DeveloperDiskImage",
	})
	if err != nil {
		return nil, fmt.Errorf("queryPersonalizedImageNonce: failed to write 'QueryNonce' command: %w", err)
	}

	var resp map[string]interface{}
	err = p.plistRw.Read(&resp)
	if err != nil {
		return nil, fmt.Errorf("queryPersonalizedImageNonce: failed to read response for 'QueryNonce': %w", err)
	}
	if nonce, ok := resp["PersonalizationNonce"].([]byte); ok {
		return nonce, nil
	}
	return nil, fmt.Errorf("queryPersonalizedImageNonce: could not get nonce from response %+v", resp)
}

func (p PersonalizedDeveloperDiskImageMounter) queryIdentifiers() (personalizationIdentifiers, error) {
	err := p.plistRw.Write(map[string]interface{}{
		"Command":               "QueryPersonalizationIdentifiers",
		"PersonalizedImageType": "DeveloperDiskImage",
	})
	if err != nil {
		return personalizationIdentifiers{}, fmt.Errorf("queryIdentifiers: failed to write 'QueryPersonalizationIdentifiers' command: %w", err)
	}

	var resp map[string]interface{}
	err = p.plistRw.Read(&resp)
	if err != nil {
		return personalizationIdentifiers{}, fmt.Errorf("queryIdentifiers: failed to read response for 'QueryPersonalizationIdentifiers': %w", err)
	}

	var persIdentifiers map[string]interface{}
	var ok bool
	if persIdentifiers, ok = resp["PersonalizationIdentifiers"].(map[string]interface{}); !ok {
		return personalizationIdentifiers{}, fmt.Errorf("queryIdentifiers: response has no 'PersonalizationIdentifiers' entry: %+v", resp)
	}

	identifiers := personalizationIdentifiers{
		AdditionalIdentifiers: map[string]interface{}{},
	}

	for k, v := range persIdentifiers {
		if strings.HasPrefix(k, "Ap,") {
			identifiers.AdditionalIdentifiers[k] = v
		}
	}

	if board, ok := persIdentifiers["BoardId"].(uint64); ok {
		identifiers.BoardId = int(board)
	}
	if chip, ok := persIdentifiers["ChipID"].(uint64); ok {
		identifiers.ChipID = int(chip)
	}
	if secDom, ok := persIdentifiers["SecurityDomain"].(uint64); ok {
		identifiers.SecurityDomain = int(secDom)
	}

	return identifiers, nil
}

func (p PersonalizedDeveloperDiskImageMounter) mountPersonalizedImage(signatureBytes []byte, trustCache []byte) error {
	err := p.plistRw.Write(map[string]interface{}{
		"Command":         "MountImage",
		"ImageSignature":  signatureBytes,
		"ImageType":       "Personalized",
		"ImageTrustCache": trustCache,
	})
	if err != nil {
		return fmt.Errorf("mountPersonalizedImage: failed to write 'MountImage' command: %w", err)
	}

	var res map[string]interface{}
	err = p.plistRw.Read(&res)
	if err != nil {
		return fmt.Errorf("mountPersonalizedImage: failed to read response for 'MountImage': %w", err)
	}
	return nil
}

func getFileSize(p string) (uint64, error) {
	info, err := os.Stat(p)
	if err != nil {
		return 0, fmt.Errorf("getFileSize: could not get file stats for '%s': %w", p, err)
	}
	if info.IsDir() {
		return 0, fmt.Errorf("getFileSize: expected a file, but got a directory: '%s'", p)
	}
	return uint64(info.Size()), nil
}
