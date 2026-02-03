package codesign

import (
	"archive/zip"
	"io"
	"os"
	"strings"
)

//CompressToIpa compresses all files and directories in the given root folder to a zip
//and writes it to the out io.Writer.
func CompressToIpa(root string, out io.Writer) error {
	zipWriter := zip.NewWriter(out)
	defer zipWriter.Close()

	files, err := GetFiles(root)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err = addFileToZip(zipWriter, file, root); err != nil {
			return err
		}
	}
	return nil
}

func addFileToZip(zipWriter *zip.Writer, filename string, root string) error {
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

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	//we remove the root from each of the filenames before zipping
	//if root does not have a trailing slash, all files will start with /
	//which will create a broken zip
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}

	// Using FileInfoHeader() above only uses the basename of the file. If we want
	// to preserve the folder structure we can overwrite this with the full path.
	header.Name = strings.Replace(filename, root, "", 1)

	//To properly store empty directories, this code is needed
	if info.IsDir() {
		header.Name += "/"
	} else {
		// Change to deflate to gain better compression
		// see http://golang.org/pkg/archive/zip/#pkg-constants
		header.Method = zip.Deflate
	}

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	//if it is a dir, not need to write a file
	if info.IsDir() {
		return nil
	}

	_, err = io.Copy(writer, fileToZip)
	return err
}
