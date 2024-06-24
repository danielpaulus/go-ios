package ios_test

import (
	"context"
	"log"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/tunnel"
)

func TestDiscover(t *testing.T) {

	const amod = "_apple-mobdev2._tcp"
	const r = "_remotepairing._tcp"
	const rmp = "_remotepairing-manual-pairing._tcp.local."
	var c = func(a string, p int) (ios.RsdService, error) {
		return tunnel.NewTCP(a, p)
	}
	ios.NewTCP = c
	rsp, err := ios.FindDevicesForService(context.Background(), r)
	if err != nil {
		t.Fatal(err)
	}
	//d := rsp[0]

	log.Fatalf("%v", rsp)
}
