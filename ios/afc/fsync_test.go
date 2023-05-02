package afc

/*
import (
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	log "github.com/sirupsen/logrus"
	"path"
	"testing"
)

const test_device_udid = "udid_here"

func TestConnection_Remove(t *testing.T) {
	deviceEnrty, _ := ios.GetDevice(test_device_udid)

	conn, err := New(deviceEnrty)
	if err != nil {
		log.Fatalf("connect service failed: %v", err)
	}

	err = conn.Remove("/DCIM/fsync.go")
	if err != nil {
		log.Fatalf("remove failed:%v", err)
	}
}

func TestConnection_RemoveAll(t *testing.T) {
	deviceEnrty, _ := ios.GetDevice(test_device_udid)

	conn, err := New(deviceEnrty)
	if err != nil {
		log.Fatalf("connect service failed: %v", err)
	}

	err = conn.RemoveAll("/DCIM/TestDir")
	if err != nil {
		log.Fatalf("remove failed:%v", err)
	}
}

func TestConnection_Mkdir(t *testing.T) {
	deviceEnrty, _ := ios.GetDevice(test_device_udid)

	conn, err := New(deviceEnrty)
	if err != nil {
		log.Fatalf("connect service failed: %v", err)
	}

	err = conn.MkDir("/DCIM/TestDir")
	if err != nil {
		log.Fatalf("mkdir failed:%v", err)
	}
}

func TestConnection_stat(t *testing.T) {
	deviceEnrty, _ := ios.GetDevice(test_device_udid)

	conn, err := New(deviceEnrty)
	if err != nil {
		log.Fatalf("connect service failed: %v", err)
	}

	si, err := conn.Stat("/DCIM/architecture_diagram.png")
	if err != nil {
		log.Fatalf("get Stat failed:%v", err)
	}
	log.Printf("Stat :%+v", si)
}

func TestConnection_listDir(t *testing.T) {
	deviceEnrty, _ := ios.GetDevice(test_device_udid)

	conn, err := New(deviceEnrty)
	if err != nil {
		log.Fatalf("connect service failed: %v", err)
	}

	flist, err := conn.listDir("/DCIM/")
	if err != nil {
		log.Fatalf("tree view failed:%v", err)
	}
	for _, v := range flist {
		fmt.Printf("path: %+v\n", v)
	}
}

func TestConnection_TreeView(t *testing.T) {
	deviceEnrty, _ := ios.GetDevice(test_device_udid)

	conn, err := New(deviceEnrty)
	if err != nil {
		log.Fatalf("connect service failed: %v", err)
	}

	err = conn.TreeView("/DCIM/", "", true)
	if err != nil {
		log.Fatalf("tree view failed:%v", err)
	}
}

func TestConnection_pullSingleFile(t *testing.T) {
	deviceEnrty, _ := ios.GetDevice(test_device_udid)

	conn, err := New(deviceEnrty)
	if err != nil {
		log.Fatalf("connect service failed: %v", err)
	}

	err = conn.PullSingleFile("/DCIM/architecture_diagram.png", "architecture_diagram.png")
	if err != nil {
		log.Fatalf("pull single file failed:%v", err)
	}
}

func TestConnection_Pull(t *testing.T) {
	deviceEnrty, _ := ios.GetDevice(test_device_udid)

	conn, err := New(deviceEnrty)
	if err != nil {
		log.Fatalf("connect service failed: %v", err)
	}
	srcPath := "/DCIM/"
	dstpath := "TempRecv"
	dstpath = path.Join(dstpath, srcPath)
	err = conn.Pull(srcPath, dstpath)
	if err != nil {
		log.Fatalf("pull failed:%v", err)
	}
}

func TestConnection_Push(t *testing.T) {
	deviceEnrty, _ := ios.GetDevice(test_device_udid)
	conn, err := New(deviceEnrty)
	if err != nil {
		log.Fatalf("connect service failed: %v", err)
	}

	srcPath := "fsync.go"
	dstpath := "/DCIM/"

	err = conn.Push(srcPath, dstpath)
	if err != nil {
		log.Fatalf("push failed:%v", err)
	}
}
*/
