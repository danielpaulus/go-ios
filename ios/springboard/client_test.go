//go:build !fast
// +build !fast

package springboard

import (
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
)

func TestListIcons(t *testing.T) {
	list, err := ios.ListDevices()
	assert.NoError(t, err)
	if len(list.DeviceList) == 0 {
		t.Skip("No devices found")
		return
	}

	device := list.DeviceList[0]

	client, err := NewClient(device)

	assert.NoError(t, err)

	screens, err := client.ListIcons()

	assert.NoError(t, err)

	assert.Greater(t, len(screens), 0)

	defer client.Close()
}
