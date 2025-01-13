package installationproxy

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/afc"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

func BenchmarkName(b *testing.B) {
	devices, err := ios.ListDevices()
	if err != nil {
		b.Fatal(err)
	}
	if len(devices.DeviceList) == 0 {
		return
	}
	device := devices.DeviceList[0]
	_ = device
	for i := 0; i < b.N; i++ {
		log.Printf("iteration %d", i)
		b.StartTimer()
		pathOnDevice := path.Join("/", uuid.New().String()+".ipa")
		f, err := os.Open("/Users/dmissmann/Library/Developer/Xcode/DerivedData/LargeBinary-cxufsuxkyzqaxzcycsbpioynmvlo/Build/Products/Debug-iphoneos/large.ipa")
		if err != nil {
			b.Fatal(err)
		}
		defer f.Close()
		//buf := make([]byte, 4*1024*1024)
		//err = afc.PushBuffer(device, pathOnDevice, f, buf)
		err = afc.Push(device, pathOnDevice, f)
		if err != nil {
			b.Fatal(err)
		}
		i, err := New(device)
		if err != nil {
			b.Fatal(err)
		}
		err = i.Install(pathOnDevice)
		if err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
		c, err := New(device)
		if err != nil {
			b.Fatal(err)
		}
		_ = c.Uninstall("com.saucelabs.LargeBinary")
		time.Sleep(100 * time.Millisecond)
	}
}
