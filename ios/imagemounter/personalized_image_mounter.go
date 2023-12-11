package imagemounter

import (
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
}

func (p personalizedDeveloperDiskImageMounter) ListImages() ([][]byte, error) {
	return listImages(p.plistRw, "Personalized", p.version)
}

func (p personalizedDeveloperDiskImageMounter) MountImage(imagePath string) error {
	manifest, err := loadBuildManifest(path.Join(imagePath, "BuildManifest.plist"))
	//TODO implement me
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
