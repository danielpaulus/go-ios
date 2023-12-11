package imagemounter

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"path"
)

type PersonalizedDeveloperDiskImageMounter struct {
	deviceConn ios.DeviceConnectionInterface
	plistRw    ios.PlistCodecReadWriter
	version    *semver.Version
	tss        tssClient
	ecid       uint64
}

func NewPersonalizedDeveloperDiskImageMounter(entry ios.DeviceEntry, version *semver.Version) (PersonalizedDeveloperDiskImageMounter, error) {
	values, err := ios.GetValuesPlist(entry)
	if err != nil {
		return PersonalizedDeveloperDiskImageMounter{}, nil
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

func (p PersonalizedDeveloperDiskImageMounter) Close() error {
	return p.deviceConn.Close()
}

func (p PersonalizedDeveloperDiskImageMounter) ListImages() ([][]byte, error) {
	return listImages(p.plistRw, "Personalized", p.version)
}

func (p PersonalizedDeveloperDiskImageMounter) MountImage(imagePath string) error {
	manifest, err := loadBuildManifest(path.Join(imagePath, "BuildManifest.plist"))
	if err != nil {
		return fmt.Errorf("failed to load build manifest. %w", err)
	}

	identifiers, err := p.queryIdentifiers()
	if err != nil {
		return fmt.Errorf("failed to query personalization identifiers. %w", err)
	}
	nonce, err := p.queryPersonalizedImageNonce()
	if err != nil {
		return fmt.Errorf("failed to get nonce. %w", err)
	}

	identity, err := manifest.findIdentity(identifiers)
	if err != nil {
		return err
	}

	signature, err := p.tss.getSignature(identity, identifiers, nonce, p.ecid)
	if err != nil {
		return fmt.Errorf("failed to get signature from Apple. %w", err)
	}

	dmgPath := path.Join(imagePath, identity.Manifest.PersonalizedDmg.Info.Path)

	imageSize, err := getFileSize(dmgPath)

	err = sendUploadRequest(p.plistRw, "Personalized", signature, imageSize)
	if err != nil {
		return err
	}
	imageFile, err := os.Open(dmgPath)
	if err != nil {
		return err
	}
	defer imageFile.Close()
	n, err := io.Copy(p.deviceConn.Writer(), imageFile)
	log.Debugf("%d bytes written", n)
	if err != nil {
		return err
	}
	err = waitForUploadComplete(p.plistRw)
	if err != nil {
		return err
	}

	trustCache, err := os.ReadFile(path.Join(imagePath, identity.Manifest.LoadableTrustCache.Info.Path))
	if err != nil {
		return fmt.Errorf("could not load trust-cache. %w", err)
	}

	err = p.mountPersonalizedImage(signature, trustCache)
	if err != nil {
		return err
	}

	return hangUp(p.plistRw)
}

func (p PersonalizedDeveloperDiskImageMounter) queryPersonalizedImageNonce() ([]byte, error) {
	err := p.plistRw.Write(map[string]interface{}{
		"Command":               "QueryNonce",
		"HostProcessName":       "CoreDeviceService",
		"PersonalizedImageType": "DeveloperDiskImage",
	})
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	err = p.plistRw.Read(&resp)
	if err != nil {
		return nil, err
	}
	log.WithField("response", resp).Info("got response")
	if nonce, ok := resp["PersonalizationNonce"].([]byte); ok {
		return nonce, nil
	}
	return nil, nil
}

func (p PersonalizedDeveloperDiskImageMounter) queryIdentifiers() (personalizationIdentifiers, error) {
	err := p.plistRw.Write(map[string]interface{}{
		"Command":               "QueryPersonalizationIdentifiers",
		"PersonalizedImageType": "DeveloperDiskImage",
	})
	if err != nil {
		return personalizationIdentifiers{}, err
	}

	var resp map[string]interface{}
	err = p.plistRw.Read(&resp)
	log.WithField("response", resp).Info("QueryPersonalizationIdentifiers")

	x := resp["PersonalizationIdentifiers"].(map[string]interface{})

	var identifiers personalizationIdentifiers

	if board, ok := x["BoardId"].(uint64); ok {
		identifiers.BoardId = int(board)
	}
	if chip, ok := x["ChipID"].(uint64); ok {
		identifiers.ChipID = int(chip)
	}
	if secDom, ok := x["SecurityDomain"].(uint64); ok {
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
		return err
	}

	var res map[string]interface{}
	err = p.plistRw.Read(&res)
	if err != nil {
		return err
	}

	log.WithField("response", res).Info("got response")
	return nil
}

func getFileSize(p string) (uint64, error) {
	info, err := os.Stat(p)
	if err != nil {
		return 0, err
	}
	if info.IsDir() {
		return 0, fmt.Errorf("expected a file, but got a directory: '%s'", p)
	}
	return uint64(info.Size()), nil
}
