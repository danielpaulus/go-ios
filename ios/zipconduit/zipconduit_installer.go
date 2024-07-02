package zipconduit

import (
	"encoding/binary"
	"hash/crc32"
	"io"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
)

/*
*
Typical weird iOS service :-D
It is a kind of special "zip" format that XCode uses to send files&folder to devices.
Sadly it is not compliant with all standard zip libraries, in particular it does not work
with the golang zipWriter implementation... OF COURSE ;-)
This is why I had to hack my own "zip" encoding together. Here is how zip_conduit works:

 1. Send PLIST "InitTransfer" in standard 4byte length + Plist format
 2. Start sending binary zip stream next
 3. Since zip does not support streaming,
    we first generate a metainf file inside a metainf directory. It contains number of files and
    total byte sizes among other things (check the struct). Probably to make streaming work we also send
    it as the first file
 4. Starting with metainf for each file:
    send a ZipFileHeader with compression set to STORE (so no compression at all)
    this also means uncompressedSize==compressedSize btw.
    be sure not to use DataDescriptors (https://en.wikipedia.org/wiki/ZIP_(file_format)#Local_file_header)
    I guess they have disabled them as it would make streaming harder. This is why golang's zip implementation
    does not work.
 5. Send the standard central directory header but not a central directory (obviously)
 6. wait for a bunch of PLISTs to be received that indicate progress and completion of installation
*/
const (
	usbmuxdServiceName string = "com.apple.streaming_zip_conduit"
	shimServiceName    string = "com.apple.streaming_zip_conduit.shim.remote"
)

// Connection exposes functions to interoperate with zipconduit
type Connection struct {
	deviceConn io.ReadWriteCloser
	plistCodec ios.PlistCodec
}

// New returns a new ZipConduit Connection for the given DeviceID and Udid
func New(device ios.DeviceEntry) (*Connection, error) {
	if !device.SupportsRsd() {
		return NewWithUsbmuxdConnection(device)
	}
	return NewWithShimConnection(device)
}

// NewWithUsbmuxdConnection connects to the streaming_zip_conduit service on the device over the usbmuxd socket
func NewWithUsbmuxdConnection(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, usbmuxdServiceName)
	if err != nil {
		return &Connection{}, err
	}

	return &Connection{
		deviceConn: deviceConn,
		plistCodec: ios.NewPlistCodec(),
	}, nil
}

// NewWithShimConnection connects to the streaming_zip_conduit service over a tunnel interface and the service port
// is obtained from remote service discovery
func NewWithShimConnection(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToShimService(device, shimServiceName)
	if err != nil {
		return &Connection{}, err
	}

	return &Connection{
		deviceConn: deviceConn,
		plistCodec: ios.NewPlistCodec(),
	}, nil
}

// SendFile will send either a zipFile or an unzipped directory to the device.
// If you specify appFilePath to a file, it will try to Unzip it to a temp dir first and then send.
// If appFilePath points to a directory, it will try to install the dir contents as an app.
func (conn Connection) SendFile(appFilePath string) error {
	openedFile, err := os.Open(appFilePath)
	if err != nil {
		return err
	}

	// Get the file information
	info, err := openedFile.Stat()
	openedFile.Close()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return conn.sendDirectory(appFilePath)
	}
	return conn.sendIpaFile(appFilePath)
}

func (conn Connection) Close() error {
	return conn.deviceConn.Close()
}

func (conn Connection) sendDirectory(dir string) error {
	tmpDir, err := os.MkdirTemp("", "prefix")
	if err != nil {
		return err
	}
	log.Debugf("created tempdir: %s", tmpDir)
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			log.WithFields(log.Fields{"dir": tmpDir}).Warn("failed removing tempdir")
		}
	}()
	var totalBytes int64
	var unzippedFiles []string
	err = filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			totalBytes += info.Size()
			unzippedFiles = append(unzippedFiles, path)
			return nil
		})
	if err != nil {
		return err
	}

	metainfFolder, metainfFile, err := addMetaInf(tmpDir, unzippedFiles, uint64(totalBytes))
	if err != nil {
		return err
	}

	init := newInitTransfer(dir + ".ipa")
	log.Debugf("sending inittransfer %+v", init)
	bytes, err := conn.plistCodec.Encode(init)
	if err != nil {
		return err
	}

	_, err = conn.deviceConn.Write(bytes)
	if err != nil {
		return err
	}

	log.Debug("writing meta inf")
	err = AddFileToZip(conn.deviceConn, metainfFolder, tmpDir)
	if err != nil {
		return err
	}
	err = AddFileToZip(conn.deviceConn, metainfFile, tmpDir)
	if err != nil {
		return err
	}
	log.Debug("meta inf send successfully")

	log.Debug("sending files....")

	for _, file := range unzippedFiles {
		err := AddFileToZip(conn.deviceConn, file, dir)
		if err != nil {
			return err
		}
	}
	log.Debug("files sent, sending central header....")
	_, err = conn.deviceConn.Write(centralDirectoryHeader)
	if err != nil {
		return err
	}

	return conn.waitForInstallation()
}

func (conn Connection) sendIpaFile(ipaFile string) error {
	tmpDir, err := os.MkdirTemp("", "prefix")
	if err != nil {
		return err
	}
	log.Debugf("created tempdir: %s", tmpDir)
	defer func() {
		err := os.RemoveAll(tmpDir)
		if err != nil {
			log.WithFields(log.Fields{"dir": tmpDir}).Warn("failed removing tempdir")
		}
	}()
	log.Debug("unzipping..")
	unzippedFiles, totalBytes, err := ios.Unzip(ipaFile, tmpDir)
	if err != nil {
		return err
	}

	metainfFolder, metainfFile, err := addMetaInf(tmpDir, unzippedFiles, totalBytes)
	if err != nil {
		return err
	}

	init := newInitTransfer(ipaFile)
	log.Debugf("sending inittransfer %+v", init)
	bytes, err := conn.plistCodec.Encode(init)
	if err != nil {
		return err
	}

	_, err = conn.deviceConn.Write(bytes)
	if err != nil {
		return err
	}

	log.Debug("writing meta inf")
	err = AddFileToZip(conn.deviceConn, metainfFolder, tmpDir)
	if err != nil {
		return err
	}
	err = AddFileToZip(conn.deviceConn, metainfFile, tmpDir)
	if err != nil {
		return err
	}
	log.Debug("meta inf send successfully")

	log.Debug("sending files....")

	for _, file := range unzippedFiles {
		err := AddFileToZip(conn.deviceConn, file, tmpDir)
		if err != nil {
			return err
		}
	}
	log.Debug("files sent, sending central header....")
	_, err = conn.deviceConn.Write(centralDirectoryHeader)
	if err != nil {
		return err
	}

	return conn.waitForInstallation()
}

func (conn Connection) waitForInstallation() error {
	for {
		msg, _ := conn.plistCodec.Decode(conn.deviceConn)
		plist, _ := ios.ParsePlist(msg)
		log.Debugf("%+v", plist)
		done, percent, status, err := evaluateProgress(plist)
		if err != nil {
			return err
		}
		if done {
			log.Info("installation successful")
			return nil
		}
		log.WithFields(log.Fields{"status": status, "percentComplete": percent}).Info("installing")
	}
}

const metainfFileName = "com.apple.ZipMetadata.plist"

func addMetaInf(metainfPath string, files []string, totalBytes uint64) (string, string, error) {
	folderPath := path.Join(metainfPath, "META-INF")
	ret, _ := ios.PathExists(folderPath)
	if !ret {
		err := os.Mkdir(folderPath, 0o777)
		if err != nil {
			return "", "", err
		}
	}
	// recordcount == files + meta-inf + metainffile
	meta := metadata{RecordCount: 2 + len(files), StandardDirectoryPerms: 16877, StandardFilePerms: -32348, TotalUncompressedBytes: totalBytes, Version: 2}
	metaBytes := ios.ToPlistBytes(meta)
	filePath := path.Join(metainfPath, "META-INF", metainfFileName)
	err := os.WriteFile(filePath, metaBytes, 0o777)
	if err != nil {
		return "", "", err
	}
	return folderPath, filePath, nil
}

func AddFileToZip(writer io.Writer, filename string, tmpdir string) error {
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	var filenameForZip string
	if runtime.GOOS == "windows" {
		filenameForZip = strings.Replace(ios.FixWindowsPaths(filename), ios.FixWindowsPaths(tmpdir)+"/", "", 1)
		if info.IsDir() && !strings.HasSuffix(filenameForZip, "/") {
			filenameForZip += "/"
		}
	} else {
		filenameForZip = strings.Replace(filename, tmpdir+"/", "", 1)
		if info.IsDir() && !strings.HasSuffix(filenameForZip, "/") {
			filenameForZip += "/"
		}
	}

	if info.IsDir() {
		// write our "zip" header for a directory
		header, name, extra := newZipHeaderDir(filenameForZip)
		err := binary.Write(writer, binary.LittleEndian, header)
		if err != nil {
			return err
		}
		err = binary.Write(writer, binary.BigEndian, name)
		if err != nil {
			return err
		}
		err = binary.Write(writer, binary.BigEndian, extra)
		return err
	}

	crc, err := calculateCrc32(fileToZip)
	if err != nil {
		return err
	}
	fileToZip.Seek(0, io.SeekStart)
	// write our "zip" file header
	header, name, extra := newZipHeader(uint32(info.Size()), crc, filenameForZip)
	err = binary.Write(writer, binary.LittleEndian, header)
	if err != nil {
		return err
	}
	err = binary.Write(writer, binary.BigEndian, name)
	if err != nil {
		return err
	}
	err = binary.Write(writer, binary.BigEndian, extra)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func calculateCrc32(reader io.Reader) (uint32, error) {
	hash := crc32.New(crc32.IEEETable)
	if _, err := io.Copy(hash, reader); err != nil {
		return 0, err
	}
	return hash.Sum32(), nil
}
