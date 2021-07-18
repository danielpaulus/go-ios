package imagemounter

import (
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/zipconduit"
	log "github.com/sirupsen/logrus"
	"io"
	"net/http"
	"os"
	"path"
)

const repo = "https://github.com/haikieu/xcode-developer-disk-image-all-platforms/raw/master/DiskImages/iPhoneOS.platform/DeviceSupport/%s.zip"
const imagepath = "devimages"
const developerDiskImageDmg = "DeveloperDiskImage.dmg"

func DownloadImageFor(device ios.DeviceEntry, baseDir string) (string, error) {
	allValues, err := ios.GetValues(device)
	if err != nil {
		return "", err
	}
	version := allValues.Value.ProductVersion
	imageDownloaded, err := validateBaseDirAndLookForImage(baseDir, version)
	if err != nil {
		return "", err
	}
	if imageDownloaded {
		log.Infof("iOS %s developer image already downloaded from https://github.com/haikieu/", version)
		return "build path here", nil
	}

	log.Infof("getting developer image for iOS %s", version)
	downloadUrl := fmt.Sprintf(repo, version)
	log.Infof("downloading from: %s", downloadUrl)
	log.Info("thank you haikieu for making these images available :-)")
	zipFileName := path.Join(baseDir, imagepath, fmt.Sprintf("%s.zip", version))
	err = downloadFile(zipFileName, downloadUrl)
	if err != nil {
		return "", err
	}
	files, size, err := zipconduit.Unzip(zipFileName, path.Join(baseDir, imagepath))
	if err != nil {
		return "", err
	}
	err = os.Remove(zipFileName)
	if err != nil {
		log.Warnf("failed deleting: '%s' with err: %+v", zipFileName, err)
	}
	log.Infof("downloaded: %+v totalbytes: %d", files, size)
	return "", nil
}

func validateBaseDirAndLookForImage(baseDir string, version string) (bool, error) {
	images := path.Join(baseDir, imagepath)
	dirHandle, err := os.Open(images)
	defer dirHandle.Close()
	if err != nil {
		err := os.MkdirAll(images, 0777)
		if err != nil {
			return false, err
		}
		return false, nil
	}

	dmgPath := path.Join(images, version, developerDiskImageDmg)
	dmgHandle, err := os.Open(dmgPath)
	if err != nil {
		return false, nil
	}
	_, err = dmgHandle.Stat()
	if err != nil {
		return false, nil
	}
	return true, nil
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
// PS: Taken from golangcode.com
func downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
