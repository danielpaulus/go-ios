package afc

import (
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
	if name == "" {
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
	if fileInfo.IsLink() {
		srcPath = fileInfo.LinkTarget
	}
	fd, err := c.Open(srcPath, READ_ONLY)
	if err != nil {
		return err
	}
	defer fd.Close()

	f, err := os.OpenFile(dstPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, fd)
	return err
}

func (conn *Client) Pull(srcPath, dstPath string) error {
	fileInfo, err := conn.Stat(srcPath)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		ret, _ := ios.PathExists(dstPath)
		if !ret {
			err = os.MkdirAll(dstPath, os.ModePerm)
			if err != nil {
				return err
			}
		}
		fileList, err := conn.List(srcPath)
		if err != nil {
			return err
		}
		for _, v := range fileList {
			// v is a device-supplied entry name. Reject names that are
			// absolute or contain ".." elements, and verify the joined host
			// destination stays inside dstPath, so a malicious device cannot
			// escape the destination directory (zip-slip / path traversal).
			if unsafeEntryName(v) {
				return fmt.Errorf("afc: refusing to pull entry with unsafe name %q under %q", v, srcPath)
			}
			sp := path.Join(srcPath, v)
			dp := filepath.Join(dstPath, v)
			if !containedIn(dstPath, dp) {
				return fmt.Errorf("afc: refusing to write %q outside destination %q", dp, dstPath)
			}
			err = conn.Pull(sp, dp)
			if err != nil {
				return err
			}
		}
	} else {
		return conn.PullSingleFile(srcPath, dstPath)
	}
	return nil
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
