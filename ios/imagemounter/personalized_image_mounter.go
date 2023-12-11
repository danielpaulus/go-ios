package imagemounter

import (
	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
)

type personalizedDeveloperDiskImageMounter struct {
	deviceConn ios.DeviceConnectionInterface
	plistRw    ios.PlistCodecReadWriter
	version    *semver.Version
}

func (p personalizedDeveloperDiskImageMounter) ListImages() ([][]byte, error) {
	return listImages(p.plistRw, "Personalized", p.version)
}

func (p personalizedDeveloperDiskImageMounter) MountImage(imagePath string) error {
	//TODO implement me
	panic("implement me")
}
