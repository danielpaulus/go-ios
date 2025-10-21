package zipconduit

import (
	"archive/zip"
	"bytes"
	"encoding/binary"
	"hash"
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

// those permissions were observed by capturing Xcode traffic, and we use exactly the same values.
// we also tried using only the last three numbers in octal representation. This worked fine, but we still use the same
// values as Xcode
const (
	stdDirPerm  = 16877  // 0o40755 -> 0o755
	stdFilePerm = -32348 // 0o37777700644 -> 0o644
)

// Connection exposes functions to interoperate with zipconduit
type Connection struct {
	deviceConn io.ReadWriteCloser
	plistCodec ios.PlistCodecReadWriter
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
		plistCodec: ios.NewPlistCodecReadWriter(deviceConn.Reader(), deviceConn.Writer()),
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
		plistCodec: ios.NewPlistCodecReadWriter(deviceConn.Reader(), deviceConn.Writer()),
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

	metainfFolder, metainfFile, err := addMetaInf(tmpDir, len(unzippedFiles), uint64(totalBytes))
	if err != nil {
		return err
	}

	init := newInitTransfer(dir + ".ipa")
	log.Debugf("sending inittransfer %+v", init)
	err = conn.plistCodec.Write(init)
	if err != nil {
		return err
	}

	hasher := crc32.NewIEEE()

	log.Debug("writing meta inf")
	err = addFileToZip(conn.deviceConn, metainfFolder, tmpDir, hasher)
	if err != nil {
		return err
	}
	err = addFileToZip(conn.deviceConn, metainfFile, tmpDir, hasher)
	if err != nil {
		return err
	}
	log.Debug("meta inf send successfully")

	log.Debug("sending files....")

	for _, file := range unzippedFiles {
		err := addFileToZip(conn.deviceConn, file, dir, hasher)
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
	ipa, err := zip.OpenReader(ipaFile)
	if err != nil {
		return err
	}
	defer ipa.Close()

	totalBytes, numFiles := zipFilesSize(ipa)

	init := newInitTransfer(ipaFile)
	log.Debugf("sending inittransfer %+v", init)
	err = conn.plistCodec.Write(init)
	if err != nil {
		return err
	}

	err = transferDirectory(conn.deviceConn, "META-INF/")
	if err != nil {
		return err
	}

	metaInfBytes := createMetaInfFile(numFiles, totalBytes)
	crc, err := calculateCrc32(bytes.NewReader(metaInfBytes), crc32.NewIEEE())
	if err != nil {
		return err
	}

	copyBuffer := make([]byte, 32*1024)

	err = transferFile(conn.deviceConn, bytes.NewReader(metaInfBytes), crc, uint32(len(metaInfBytes)), path.Join("META-INF", metainfFileName), copyBuffer)
	if err != nil {
		return err
	}

	for _, f := range ipa.File {
		if f.FileInfo().IsDir() {
			err := transferDirectory(conn.deviceConn, f.Name)
			if err != nil {
				return err
			}
			continue
		}

		uncompressedFile, err := f.Open()
		if err != nil {
			return err
		}

		err = transferFile(conn.deviceConn, uncompressedFile, f.CRC32, uint32(f.UncompressedSize64), f.Name, copyBuffer)
		_ = uncompressedFile.Close()
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
		var plist map[string]interface{}
		err := conn.plistCodec.Read(&plist)
		if err != nil {
			return err
		}
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

func addMetaInf(metainfPath string, numFiles int, totalBytes uint64) (string, string, error) {
	folderPath := path.Join(metainfPath, "META-INF/")
	ret, _ := ios.PathExists(folderPath)
	if !ret {
		err := os.Mkdir(folderPath, 0o777)
		if err != nil {
			return "", "", err
		}
	}
	metaBytes := createMetaInfFile(numFiles, totalBytes)
	filePath := path.Join(metainfPath, "META-INF", metainfFileName)
	err := os.WriteFile(filePath, metaBytes, 0o777)
	if err != nil {
		return "", "", err
	}
	return folderPath, filePath, nil
}

func createMetaInfFile(numFiles int, totalBytes uint64) []byte {
	meta := metadata{RecordCount: 2 + numFiles, StandardDirectoryPerms: stdDirPerm, StandardFilePerms: stdFilePerm, TotalUncompressedBytes: totalBytes, Version: 2}
	return ios.ToPlistBytes(meta)
}

func addFileToZip(writer io.Writer, filename string, tmpdir string, hasher hash.Hash32) error {
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
		_, err = writer.Write(name)
		if err != nil {
			return err
		}
		if _, werr := writer.Write(extra); werr != nil {
			return werr
		}
		return err
	}

	crc, err := calculateCrc32(fileToZip, hasher)
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
	_, err = writer.Write(name)
	if err != nil {
		return err
	}
	_, err = writer.Write(extra)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, fileToZip)
	return err
}

func calculateCrc32(reader io.Reader, hasher hash.Hash32) (uint32, error) {
	hasher.Reset()
	if _, err := io.Copy(hasher, reader); err != nil {
		return 0, err
	}
	return hasher.Sum32(), nil
}

// zipFilesSize counts all the files that are stored in the zip archive and adds up their uncompressed size
func zipFilesSize(r *zip.ReadCloser) (size uint64, numFiles int) {
	for _, f := range r.File {
		size += f.UncompressedSize64
		numFiles++
	}
	return
}

func transferFile(dst io.Writer, src io.Reader, crc uint32, uncompressedSize uint32, dstFilePath string, buffer []byte) error {
	header, name, extra := newZipHeader(uncompressedSize, crc, dstFilePath)
	err := binary.Write(dst, binary.LittleEndian, header)
	if err != nil {
		return err
	}
	_, err = dst.Write(name)
	if err != nil {
		return err
	}
	_, err = dst.Write(extra)
	if err != nil {
		return err
	}
	_, err = io.CopyBuffer(dst, src, buffer)
	return err
}

func transferDirectory(writer io.Writer, dstDirPath string) error {
	// write our "zip" header for a directory
	header, name, extra := newZipHeaderDir(dstDirPath)
	err := binary.Write(writer, binary.LittleEndian, header)
	if err != nil {
		return err
	}
	_, err = writer.Write(name)
	if err != nil {
		return err
	}
	_, err = writer.Write(extra)
	return err
}
