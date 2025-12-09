package afc

import (
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/danielpaulus/go-ios/ios"
)

const serviceName = "com.apple.afc"

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
			sp := path.Join(srcPath, v)
			dp := path.Join(dstPath, v)
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
