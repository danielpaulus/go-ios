package imagemounter

import (
	"bytes"
	"crypto/sha512"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
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
	entry      ios.DeviceEntry
}

type personalizedImageInfo struct {
	identifiers personalizationIdentifiers
	identity    buildIdentity
	dmgPath     string
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
		entry:      entry,
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
// MountImage first tries to reuse an existing device-side manifest via QueryPersonalizationManifest
// to avoid unnecessary Apple TSS requests on re-mounts. If no manifest exists, it falls back
// to querying a nonce and getting a new signature from Apple's TSS server.
func (p PersonalizedDeveloperDiskImageMounter) MountImage(imagePath string) error {
	image, err := p.findDmgFile(imagePath)
	if err != nil {
		return fmt.Errorf("MountImage: could not find .dmg-file for image path %+v: %w", imagePath, err)
	}

	signature, err := p.queryPersonalizationManifest(image.dmgPath)
	if err != nil {
		golog.Info("no existing device-side manifest, requesting new signature from Apple TSS", "module", logModule, "udid", p.entry.Properties.SerialNumber, "imagePath", imagePath)

		p, err = p.closeAndReconnect()
		if err != nil {
			return fmt.Errorf("MountImage: failed to reconnect after manifest query: %w", err)
		}

		nonce, err := p.queryPersonalizedImageNonce()
		if err != nil {
			return fmt.Errorf("MountImage: failed to get nonce: %w", err)
		}

		signature, err = p.tss.getSignature(image.identity, image.identifiers, nonce, p.ecid)
		if err != nil {
			return fmt.Errorf("MountImage: failed to get signature from Apple: %w", err)
		}
	} else {
		golog.Info("reusing existing device-side manifest, skipping Apple TSS", "module", logModule, "udid", p.entry.Properties.SerialNumber, "imagePath", imagePath)
	}

	imageSize, err := getFileSize(image.dmgPath)
	if err != nil {
		return fmt.Errorf("MountImage: %w", err)
	}

	err = sendUploadRequest(p.plistRw, p.entry.Properties.SerialNumber, "Personalized", signature, imageSize)
	if err != nil {
		return fmt.Errorf("MountImage: failed to send upload request for image: %w", err)
	}
	imageFile, err := os.Open(image.dmgPath)
	if err != nil {
		return fmt.Errorf("MountImage: failed to open developer disk dmg file '%s': %w", image.dmgPath, err)
	}
	defer imageFile.Close()
	n, err := io.Copy(p.deviceConn.Writer(), imageFile)
	golog.Debug("bytes written", "module", logModule, "udid", p.entry.Properties.SerialNumber, "dmgPath", image.dmgPath, "count", n)
	if err != nil {
		return fmt.Errorf("MountImage: could not copy developer disk image to the device: %w", err)
	}
	err = waitForUploadComplete(p.plistRw, p.entry.Properties.SerialNumber)
	if err != nil {
		return err
	}

	trustCache, err := os.ReadFile(path.Join(imagePath, image.identity.trustCachePath()))
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
	golog.Debug("sending", "module", logModule, "udid", p.entry.Properties.SerialNumber, "request", req)
	err := p.plistRw.Write(req)
	if err != nil {
		return err
	}
	return readUnmountResponse(p.plistRw, p.entry.Properties.SerialNumber)
}

func (p PersonalizedDeveloperDiskImageMounter) queryPersonalizationManifest(dmgPath string) ([]byte, error) {
	digest, err := sha384FileHash(dmgPath)
	if err != nil {
		return nil, fmt.Errorf("queryPersonalizationManifest: failed to hash DMG: %w", err)
	}

	err = p.plistRw.Write(map[string]interface{}{
		"Command":               "QueryPersonalizationManifest",
		"PersonalizedImageType": "DeveloperDiskImage",
		"ImageType":             "DeveloperDiskImage",
		"ImageSignature":        digest,
	})
	if err != nil {
		return nil, fmt.Errorf("queryPersonalizationManifest: failed to write command: %w", err)
	}

	var resp map[string]interface{}
	err = p.plistRw.Read(&resp)
	if err != nil {
		return nil, fmt.Errorf("queryPersonalizationManifest: failed to read response: %w", err)
	}

	if sig, ok := resp["ImageSignature"].([]byte); ok {
		return sig, nil
	}
	return nil, fmt.Errorf("queryPersonalizationManifest: no ImageSignature in response %+v", resp)
}

func (p PersonalizedDeveloperDiskImageMounter) closeAndReconnect() (PersonalizedDeveloperDiskImageMounter, error) {
	p.deviceConn.Close()

	deviceConn, err := ios.ConnectToService(p.entry, serviceName)
	if err != nil {
		return p, fmt.Errorf("closeAndReconnect: failed to reconnect to %s: %w", serviceName, err)
	}

	return PersonalizedDeveloperDiskImageMounter{
		deviceConn: deviceConn,
		plistRw:    ios.NewPlistCodecReadWriter(deviceConn.Reader(), deviceConn.Writer()),
		version:    p.version,
		tss:        p.tss,
		ecid:       p.ecid,
		entry:      p.entry,
	}, nil
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

	return readImageMounterResponse(p.plistRw, p.entry.Properties.SerialNumber, "MountImage", "Complete")
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

// sha384FileHash for creating the hash of the .dmg image that is used for creating the signature, as well as to
// compare the mounted image on a device to the one on disk (the device returns this hash in the ListImages call)
func sha384FileHash(p string) ([]byte, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, fmt.Errorf("sha384FileHash: failed to open DMG: %w", err)
	}
	defer f.Close()

	h := sha512.New384()
	if _, err := io.Copy(h, f); err != nil {
		return nil, fmt.Errorf("sha384FileHash: failed to hash DMG: %w", err)
	}
	digest := h.Sum(nil)
	return digest, nil
}

// findDmgFile uses the device identifiers to find the matching .dmg file that needs to be mounted to the device
func (p PersonalizedDeveloperDiskImageMounter) findDmgFile(imagePath string) (personalizedImageInfo, error) {
	manifest, err := loadBuildManifest(path.Join(imagePath, "BuildManifest.plist"))
	if err != nil {
		return personalizedImageInfo{}, fmt.Errorf("findDmgFile: failed to load build manifest: %w", err)
	}

	identifiers, err := p.queryIdentifiers()
	if err != nil {
		return personalizedImageInfo{}, fmt.Errorf("findDmgFile: failed to query personalization identifiers: %w", err)
	}

	identity, err := manifest.findIdentity(identifiers)
	if err != nil {
		return personalizedImageInfo{}, fmt.Errorf("findDmgFile: could not find identity for identifiers %+v: %w", identifiers, err)
	}

	dmgPath := path.Join(imagePath, identity.dmgPath())
	return personalizedImageInfo{
		identifiers: identifiers,
		identity:    identity,
		dmgPath:     dmgPath,
	}, nil
}

// IsImageMounted verifies if there is currently an image mounted on the device, and if that is the case, it compares
// it against the image at imagePath to be the same
func (p PersonalizedDeveloperDiskImageMounter) IsImageMounted(imagePath string) (bool, error) {
	mountedSignatures, err := p.ListImages()
	if err != nil {
		return false, fmt.Errorf("IsImageMounted: failed to list images: %w", err)
	}
	if len(mountedSignatures) == 0 {
		return false, nil
	}
	image, err := p.findDmgFile(imagePath)
	if err != nil {
		return false, err
	}

	imageHash, err := sha384FileHash(image.dmgPath)
	if err != nil {
		return false, fmt.Errorf("IsImageMounted: failed to calculate sha384: %w", err)
	}

	for _, signature := range mountedSignatures {
		if bytes.Compare(signature, imageHash) == 0 {
			return true, nil
		}
	}
	return false, nil
}
