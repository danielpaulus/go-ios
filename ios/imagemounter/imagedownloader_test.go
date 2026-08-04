package imagemounter_test

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/imagemounter"
	"github.com/elazarl/goproxy"
	"github.com/stretchr/testify/assert"
)

func TestVersionMatching(t *testing.T) {
	assert.Equal(t, "11.2 (15C5092b)", imagemounter.MatchAvailable("11.2.5"))
	assert.Equal(t, "12.2 (16E5191d)", imagemounter.MatchAvailable("12.2.5"))
	assert.Equal(t, "13.5", imagemounter.MatchAvailable("13.6.1"))
	assert.Equal(t, "14.7.1", imagemounter.MatchAvailable("14.7.1"))
	assert.Equal(t, "15.3.1", imagemounter.MatchAvailable("15.3.1"))
	assert.Equal(t, "15.4", imagemounter.MatchAvailable("15.4.1"))
	assert.Equal(t, "15.7", imagemounter.MatchAvailable("15.7.2"))
	assert.Equal(t, "16.6", imagemounter.MatchAvailable("19.4.1"))
}

func TestUsesProxy(t *testing.T) {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	wg := sync.WaitGroup{}
	wg.Add(1)

	proxy.OnRequest().HandleConnectFunc(func(host string, ctx *goproxy.ProxyCtx) (*goproxy.ConnectAction, string) {
		log.Printf("Got request for %s", host)
		wg.Done()
		return goproxy.OkConnect, host
	})

	go func() {
		log.Print(http.ListenAndServe(":60001", proxy))
	}()
	tempDir, err := os.MkdirTemp("", "example")
	if err != nil {
		fmt.Printf("Error creating temporary directory: %v\n", err)
		t.Fail()
		return
	}
	defer os.RemoveAll(tempDir)
	ios.UseHttpProxy("http://localhost:60001")
	path, err := imagemounter.Download17Plus(tempDir, ios.IOS17())
	if !assert.Nil(t, err) {
		t.Fail()
	}
	log.Printf("Downloaded to %s", path)
	wg.Wait()
	d, _ := ios.ListDevices()
	if len(d.DeviceList) == 0 {
		t.Skip("No device attached")
		return
	}
	wg.Add(1)
	m, err := imagemounter.NewPersonalizedDeveloperDiskImageMounter(d.DeviceList[0], ios.IOS17())
	if !assert.Nil(t, err) {
		t.Fail()
	}

	err = m.MountImage(path)
	if !assert.Nil(t, err) {
		t.Fail()
	}
	wg.Wait()
	//mounter.MountImage(path)
}

func TestWorksWithoutProxy(t *testing.T) {

	tempDir, err := os.MkdirTemp("", "example")
	if err != nil {
		fmt.Printf("Error creating temporary directory: %v\n", err)
		return
	}
	defer os.RemoveAll(tempDir)
	ios.UseHttpProxy("")
	path, err := imagemounter.Download17Plus(tempDir, ios.IOS17())
	if !assert.Nil(t, err) {
		t.Fail()
	}
	log.Printf("Downloaded to %s", path)

	d, _ := ios.ListDevices()
	if len(d.DeviceList) == 0 {
		t.Skip("No device attached")
		return
	}
	m, err := imagemounter.NewPersonalizedDeveloperDiskImageMounter(d.DeviceList[0], ios.IOS17())
	if !assert.Nil(t, err) {
		t.Fail()
	}

	err = m.MountImage(path)
	if !assert.Nil(t, err) {
		t.Fail()
	}

}
