package codesign

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ExtractIpa takes a io.ReaderAt and a length to extract a zip archive to a
// temporary directory that will be returned as a string path.
// It automatically skips "__MACOSX" resource fork folders, which mac os sometimes adds to zip files.
// Zipping those will break ipa files.
// It returns duration of the process, the temp directory containing the extracted files
// or an error.
// It is the callers responsibility to clean up the temp dir.
func ExtractIpa(zipFile io.ReaderAt, length int64) (time.Duration, string, error) {
	destination, err := os.MkdirTemp("", "goios-ipa-extract")
	if err != nil {
		return 0, "", err
	}

	start := time.Now()
	r, err := zip.NewReader(zipFile, length)
	if err != nil {
		return 0, "", err
	}

	for _, zf := range r.File {
		if isMacOsResourceForkFolder(zf.Name) {
			continue
		}
		if err := unzipFile(zf, destination); err != nil {
			return 0, "", err
		}
	}

	return time.Since(start), destination, nil
}

func isMacOsResourceForkFolder(name string) bool {
	return strings.Contains(name, "__MACOSX")
}

func unzipFile(zf *zip.File, destination string) error {
	if strings.HasSuffix(zf.Name, "/") {
		return mkdir(filepath.Join(destination, zf.Name))
	}

	rc, err := zf.Open()
	if err != nil {
		return fmt.Errorf("%s: open compressed file: %v", zf.Name, err)
	}
	defer rc.Close()

	return writeNewFile(filepath.Join(destination, zf.Name), rc, zf.FileInfo().Mode())
}

func writeNewFile(fpath string, in io.Reader, fm os.FileMode) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	out, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("%s: creating new file: %v", fpath, err)
	}
	defer out.Close()

	err = out.Chmod(fm)
	if err != nil && runtime.GOOS != "windows" {
		return fmt.Errorf("%s: changing file mode: %v", fpath, err)
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("%s: writing file: %v", fpath, err)
	}
	return nil
}

func writeNewSymbolicLink(fpath string, target string) error {
	err := os.MkdirAll(filepath.Dir(fpath), 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory for file: %v", fpath, err)
	}

	err = os.Symlink(target, fpath)
	if err != nil {
		return fmt.Errorf("%s: making symbolic link for: %v", fpath, err)
	}

	return nil
}

func mkdir(dirPath string) error {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("%s: making directory: %v", dirPath, err)
	}
	return nil
}
