package ncm

import (
	"strings"
	"testing"

	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/assert"
)

func TestNetworkCommandArgumentsAreValidated(t *testing.T) {
	tests := []struct {
		name string
		call func() error
	}{
		{"interface shell metacharacters", func() error { _, err := SetInterfaceUp("eth0;id"); return err }},
		{"interface whitespace", func() error { _, err := AddInterface("eth0 bad", "fc00::fb/64"); return err }},
		{"IPv6 shell metacharacters", func() error { _, err := AddInterface("eth0", "fc00::fb/64;id"); return err }},
		{"IPv4 prefix", func() error { _, err := AddInterface("eth0", "192.0.2.1/24"); return err }},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Error(t, test.call())
		})
	}
}

func TestUsbmuxDVersion(t *testing.T) {
	const usbmuxVersion = "usbmuxd 1.1.1-56-g360619c"
	version := strings.Replace(usbmuxVersion, "usbmuxd ", "", -1)
	v, err := semver.NewVersion(version)
	assert.Nil(t, err)
	ok := v.Equal(v) || v.GreaterThan(v)
	print(ok)
}
