package ncm

import (
	"github.com/Masterminds/semver"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestUsbmuxDVersion(t *testing.T) {
	const usbmuxVersion = "usbmuxd 1.1.1-56-g360619c"
	version := strings.Replace(usbmuxVersion, "usbmuxd ", "", -1)
	v, err := semver.NewVersion(version)
	assert.Nil(t, err)
	ok := v.Equal(v) || v.GreaterThan(v)
	print(ok)
}
