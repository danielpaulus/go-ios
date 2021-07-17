package zipconduit

import (
	"archive/zip"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
	"hash/crc32"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

/**
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
const serviceName string = "com.apple.streaming_zip_conduit"

//Connection exposes functions to interoperate with zipconduit
type Connection struct {
	deviceConn ios.DeviceConnectionInterface
	plistCodec ios.PlistCodec
}

//New returns a new ZipConduit Connection for the given DeviceID and Udid
func New(device ios.DeviceEntry) (*Connection, error) {
	deviceConn, err := ios.ConnectToService(device, serviceName)
	if err != nil {
		return &Connection{}, err
	}

	return &Connection{
		deviceConn: deviceConn,
		plistCodec: ios.NewPlistCodec(),
	}, nil
}

func (conn Connection) SendFile(ipaFile string) error {
	init := newInitTransfer(ipaFile)
	log.Info("sending inittransfer")
	bytes, err := conn.plistCodec.Encode(init)
	if err != nil {
		return err
	}
	println(hex.Dump(bytes))
	err = conn.deviceConn.Send(bytes)
	if err != nil {
		return err
	}

	tmpDir, err := ioutil.TempDir("", "prefix")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)
	unzippedFiles, totalBytes, err := Unzip(ipaFile, tmpDir)
	if err != nil {
		return err
	}
	deviceStream := conn.deviceConn.Writer()

	err = os.Mkdir(path.Join(tmpDir, "META-INF"), 0777)
	fileMetaNAme := "com.apple.ZipMetadata.plist"

	meta := metadata{RecordCount: 2 + len(unzippedFiles), StandardDirectoryPerms: 16877, StandardFilePerms: -32348, TotalUncompressedBytes: totalBytes, Version: 2}
	metaBytes := ios.ToPlistBytes(meta)
	log.Infof("%x", metaBytes)
	println(hex.Dump(metaBytes))
	ioutil.WriteFile(path.Join(tmpDir, "META-INF", fileMetaNAme), metaBytes, 0777)
	if err != nil {
		return err
	}

	log.Info("writing meta inf")
	err = AddFileToZip(deviceStream, path.Join(tmpDir, "META-INF"), tmpDir)
	if err != nil {
		return err
	}
	err = AddFileToZip(deviceStream, path.Join(tmpDir, "META-INF", fileMetaNAme), tmpDir)
	if err != nil {
		return err
	}

	log.Info("Writing..")

	for _, file := range unzippedFiles {
		log.Info(file)
		err := AddFileToZip(deviceStream, file, tmpDir)
		if err != nil {
			return err
		}
	}

	_, err = conn.deviceConn.Writer().Write(centralDirectoryHeader)

	if err != nil {
		return err
	}

	for {
		msg, _ := conn.plistCodec.Decode(conn.deviceConn.Reader())
		plist, _ := ios.ParsePlist(msg)
		log.Infof("%+v", plist)
	}
	return nil
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
	filenameForZip := strings.Replace(filename, tmpdir+"/", "", 1)
	if info.IsDir() && !strings.HasSuffix(filenameForZip, "/") {
		filenameForZip += "/"
	}
	if info.IsDir() {
		header, name, extra := newZipHeaderDir(filenameForZip)
		err := binary.Write(writer, binary.LittleEndian, header)
		binary.Write(writer, binary.BigEndian, name)
		binary.Write(writer, binary.BigEndian, extra)
		return err
	}
	filebytes, err := io.ReadAll(fileToZip)
	if err != nil {
		log.Fatal("err reading file")
	}
	crc := crc32.ChecksumIEEE(filebytes)
	header, name, extra := newZipHeader(uint32(len(filebytes)), crc, filenameForZip)
	err = binary.Write(writer, binary.LittleEndian, header)
	binary.Write(writer, binary.BigEndian, name)
	binary.Write(writer, binary.BigEndian, extra)
	if err != nil {
		return err
	}
	_, err = writer.Write(filebytes)
	//_, err = io.Copy(writer, fileToZip)
	return err
}



func Unzip(src string, dest string) ([]string, uint64, error) {
	var overallSize uint64
	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, 0, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, 0, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, 0, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, 0, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, 0, err
		}

		_, err = io.Copy(outFile, rc)
		//sizeStat, err := outFile.Stat()
		overallSize += f.UncompressedSize64
		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, 0, err
		}
	}
	return filenames, overallSize, nil
}
