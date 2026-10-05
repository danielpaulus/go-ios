package imagemounter

import (
	"errors"
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
)

// ios17ImageMounter combines the mounters that can be used for iOS 17+ devices. Images are
// mounted with the first mounter that succeeds, while listing, checking and unmounting
// consider all of them, as an image could have been mounted with any of them.
type ios17ImageMounter struct {
	// mounters in the order they are tried for mounting an image
	mounters []namedImageMounter
	udid     string
}

type namedImageMounter struct {
	name string
	ImageMounter
}

func newIOS17ImageMounter(device ios.DeviceEntry, version *semver.Version) (ImageMounter, error) {
	udid := device.Properties.SerialNumber
	var mounters []namedImageMounter
	var errs []error

	personalized, err := NewPersonalizedDeveloperDiskImageMounter(device, version)
	if err == nil {
		mounters = append(mounters, namedImageMounter{name: "personalized", ImageMounter: personalized})
	} else {
		golog.Debug("personalized image mounter is not available", "module", logModule, "udid", udid, "err", err)
		errs = append(errs, err)
	}

	if SupportsCryptexDDI(device) {
		cryptex, err := NewCryptexDeveloperDiskImageMounter(device)
		if err == nil {
			mounters = append(mounters, namedImageMounter{name: "cryptex", ImageMounter: cryptex})
		} else {
			golog.Debug("cryptex image mounter is not available", "module", logModule, "udid", udid, "err", err)
			errs = append(errs, err)
		}
	}

	if len(mounters) == 0 {
		return nil, fmt.Errorf("newIOS17ImageMounter: no image mounter available: %w", errors.Join(errs...))
	}
	return &ios17ImageMounter{mounters: mounters, udid: udid}, nil
}

// ListImages returns the images of all mounters. Mounters that fail to list their images are
// skipped, unless all of them fail.
func (m *ios17ImageMounter) ListImages() ([][]byte, error) {
	var images [][]byte
	var errs []error
	for _, mounter := range m.mounters {
		list, err := mounter.ListImages()
		if err != nil {
			golog.Debug("failed listing images", "module", logModule, "udid", m.udid, "mounter", mounter.name, "err", err)
			errs = append(errs, fmt.Errorf("%s: %w", mounter.name, err))
			continue
		}
		images = append(images, list...)
	}
	if len(errs) == len(m.mounters) {
		return nil, fmt.Errorf("ListImages: %w", errors.Join(errs...))
	}
	return images, nil
}

// MountImage mounts the image at imagePath with the first mounter that succeeds
func (m *ios17ImageMounter) MountImage(imagePath string) error {
	var errs []error
	for _, mounter := range m.mounters {
		err := mounter.MountImage(imagePath)
		if err == nil {
			golog.Info("mounted developer disk image", "module", logModule, "udid", m.udid, "mounter", mounter.name, "imagePath", imagePath)
			return nil
		}
		golog.Info("failed mounting developer disk image", "module", logModule, "udid", m.udid, "mounter", mounter.name, "imagePath", imagePath, "err", err)
		errs = append(errs, fmt.Errorf("%s: %w", mounter.name, err))
	}
	return fmt.Errorf("MountImage: %w", errors.Join(errs...))
}

// UnmountImage unmounts the images of every mounter that has an image mounted
func (m *ios17ImageMounter) UnmountImage() error {
	unmounted := false
	var errs []error
	for _, mounter := range m.mounters {
		images, err := mounter.ListImages()
		if err != nil {
			golog.Debug("failed listing images", "module", logModule, "udid", m.udid, "mounter", mounter.name, "err", err)
			continue
		}
		if len(images) == 0 {
			continue
		}
		if err := mounter.UnmountImage(); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", mounter.name, err))
			continue
		}
		golog.Info("unmounted developer disk image", "module", logModule, "udid", m.udid, "mounter", mounter.name)
		unmounted = true
	}
	if len(errs) > 0 {
		return fmt.Errorf("UnmountImage: %w", errors.Join(errs...))
	}
	if !unmounted {
		return errors.New("UnmountImage: no developer disk image is mounted")
	}
	return nil
}

// IsImageMounted reports whether any of the mounters has the image at imagePath mounted
func (m *ios17ImageMounter) IsImageMounted(imagePath string) (bool, error) {
	var errs []error
	for _, mounter := range m.mounters {
		mounted, err := mounter.IsImageMounted(imagePath)
		if err != nil {
			golog.Debug("failed checking if image is mounted", "module", logModule, "udid", m.udid, "mounter", mounter.name, "imagePath", imagePath, "err", err)
			errs = append(errs, fmt.Errorf("%s: %w", mounter.name, err))
			continue
		}
		if mounted {
			return true, nil
		}
	}
	if len(errs) == len(m.mounters) {
		return false, fmt.Errorf("IsImageMounted: %w", errors.Join(errs...))
	}
	return false, nil
}

// Close closes all mounters
func (m *ios17ImageMounter) Close() error {
	var errs []error
	for _, mounter := range m.mounters {
		errs = append(errs, mounter.Close())
	}
	return errors.Join(errs...)
}
