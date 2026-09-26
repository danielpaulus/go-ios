package afc

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
)

const serviceName = "com.apple.afc"

// unsafeEntryName reports whether a device-supplied directory entry name is
// unsafe to join onto a host path. A legitimate single filename ("foo.txt",
// "Documents") is never affected; only absolute paths or names containing a
// ".." path element (which could escape the destination directory) are
// rejected. This is defense in depth on top of containedIn.
func unsafeEntryName(name string) bool {
	if name == "" || name == "." {
		return true
	}
	if filepath.IsAbs(name) || path.IsAbs(name) {
		return true
	}
	// Split on both separators so a "..\\" style name is caught on every OS.
	for _, part := range strings.FieldsFunc(name, func(r rune) bool {
		return r == '/' || r == filepath.Separator
	}) {
		if part == ".." {
			return true
		}
	}
	return false
}

// containedIn reports whether the host path candidate stays within root after
// cleaning. It returns true when candidate is exactly root or a descendant of
// root. This is the containment check that prevents an AFC pull from writing
// outside the destination directory via a maliciously named device entry.
func containedIn(root, candidate string) bool {
	root = filepath.Clean(root)
	candidate = filepath.Clean(candidate)
	if candidate == root {
		return true
	}
	return strings.HasPrefix(candidate, root+string(os.PathSeparator))
}

func (c *Client) PullSingleFile(srcPath, dstPath string) error {
	fileInfo, err := c.Stat(srcPath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("afc: %q is a directory", srcPath)
	}
	return c.pullAtomically(srcPath, dstPath, fileInfo)
}

func (conn *Client) Pull(srcPath, dstPath string) error {
	fileInfo, err := conn.Stat(srcPath)
	if err != nil {
		return err
	}
	return conn.pullAtomically(srcPath, dstPath, fileInfo)
}

func (c *Client) pullAtomically(srcPath, dstPath string, fileInfo FileInfo) error {
	cleanDestination := filepath.Clean(dstPath)
	parentPath := filepath.Dir(cleanDestination)
	destinationName := filepath.Base(cleanDestination)
	if destinationName == "." || destinationName == string(os.PathSeparator) || !filepath.IsLocal(destinationName) {
		return fmt.Errorf("afc: destination must name an entry below its parent: %s", dstPath)
	}
	if err := os.MkdirAll(parentPath, 0755); err != nil {
		return err
	}
	parent, err := os.OpenRoot(parentPath)
	if err != nil {
		return err
	}
	defer parent.Close()

	stagingName, err := makePullStagingDir(parent, "."+destinationName+".pull-")
	if err != nil {
		return err
	}
	defer parent.RemoveAll(stagingName)
	staging, err := parent.OpenRoot(stagingName)
	if err != nil {
		return err
	}

	stagedPath := stagingName
	if fileInfo.IsDir() {
		err = c.pullEntry(srcPath, staging, ".", fileInfo)
	} else {
		const stagedFile = "content"
		err = c.pullEntry(srcPath, staging, stagedFile, fileInfo)
		stagedPath = filepath.Join(stagingName, stagedFile)
	}
	closeErr := staging.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}

	backupName := ""
	if _, err := parent.Lstat(destinationName); err == nil {
		backupName, err = unusedPullName(parent, "."+destinationName+".backup-")
		if err != nil {
			return err
		}
		if err := parent.Rename(destinationName, backupName); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if err := parent.Rename(stagedPath, destinationName); err != nil {
		if backupName != "" {
			_ = parent.Rename(backupName, destinationName)
		}
		return err
	}
	if backupName != "" {
		if err := parent.RemoveAll(backupName); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) pullEntry(srcPath string, root *os.Root, relativePath string, fileInfo FileInfo) error {
	if fileInfo.IsDir() {
		if relativePath != "." {
			if err := root.MkdirAll(relativePath, 0755); err != nil {
				return err
			}
		}
		fileList, err := c.List(srcPath)
		if err != nil {
			return err
		}
		for _, v := range fileList {
			if unsafeEntryName(v) {
				return fmt.Errorf("afc: refusing to pull entry with unsafe name %q under %q", v, srcPath)
			}
			childPath := filepath.Join(relativePath, filepath.FromSlash(v))
			if !filepath.IsLocal(childPath) {
				return fmt.Errorf("afc: refusing to write unsafe destination %q", childPath)
			}
			childSource := path.Join(srcPath, v)
			childInfo, err := c.Stat(childSource)
			if err != nil {
				return err
			}
			if err := c.pullEntry(childSource, root, childPath, childInfo); err != nil {
				return err
			}
		}
		return nil
	}

	if fileInfo.IsLink() {
		srcPath = fileInfo.LinkTarget
	}
	deviceFile, err := c.Open(srcPath, READ_ONLY)
	if err != nil {
		return err
	}
	hostFile, err := root.OpenFile(relativePath, os.O_CREATE|os.O_WRONLY|os.O_EXCL, 0666)
	if err != nil {
		_ = deviceFile.Close()
		return err
	}
	_, copyErr := io.Copy(hostFile, deviceFile)
	hostCloseErr := hostFile.Close()
	deviceCloseErr := deviceFile.Close()
	if copyErr != nil {
		return copyErr
	}
	if hostCloseErr != nil {
		return hostCloseErr
	}
	return deviceCloseErr
}

func makePullStagingDir(root *os.Root, prefix string) (string, error) {
	for range 100 {
		random := make([]byte, 16)
		if _, err := rand.Read(random); err != nil {
			return "", err
		}
		name := prefix + hex.EncodeToString(random)
		if err := root.Mkdir(name, 0700); err == nil {
			return name, nil
		} else if !os.IsExist(err) {
			return "", err
		}
	}
	return "", fmt.Errorf("afc: failed to create a unique pull staging directory")
}

func unusedPullName(root *os.Root, prefix string) (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", err
	}
	name := prefix + hex.EncodeToString(random)
	if _, err := root.Lstat(name); err == nil {
		return "", os.ErrExist
	} else if !os.IsNotExist(err) {
		return "", err
	}
	return name, nil
}

func (conn *Client) Push(srcPath, dstPath string) error {
	ret, _ := ios.PathExists(srcPath)
	if !ret {
		return fmt.Errorf("%s: no such file.", srcPath)
	}

	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	if fileInfo, err := conn.Stat(dstPath); err == nil {
		if fileInfo.IsDir() {
			dstPath = path.Join(dstPath, filepath.Base(srcPath))
		}
	}

	return conn.WriteToFile(f, dstPath)
}

func (conn *Client) WriteToFile(reader io.Reader, dstPath string) error {
	if fileInfo, err := conn.Stat(dstPath); err == nil {
		if fileInfo.IsDir() {
			return fmt.Errorf("%s is a directory, cannot write to it as file", dstPath)
		}
	}

	fd, err := conn.Open(dstPath, WRITE_ONLY_CREATE_TRUNC)
	if err != nil {
		return err
	}
	defer fd.Close()

	_, err = io.Copy(fd, reader)
	if err != nil {
		return err
	}
	return nil
}
