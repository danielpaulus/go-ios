package imagemounter

import (
	"fmt"
	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
	"path"
)

type personalizedDeveloperDiskImageMounter struct {
	deviceConn ios.DeviceConnectionInterface
	plistRw    ios.PlistCodecReadWriter
	version    *semver.Version
	tss        tssClient
	ecid       uint64
}

func (p personalizedDeveloperDiskImageMounter) ListImages() ([][]byte, error) {
	return listImages(p.plistRw, "Personalized", p.version)
}

func (p personalizedDeveloperDiskImageMounter) MountImage(imagePath string) error {
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
	panic("implement me")
}

func (p personalizedDeveloperDiskImageMounter) queryPersonalizedImageNonce() ([]byte, error) {
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

func (p personalizedDeveloperDiskImageMounter) queryIdentifiers() (personalizationIdentifiers, error) {
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
