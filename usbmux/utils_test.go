package usbmux_test

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"
	"usbmuxd/usbmux"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "update golden files")

type SampleData struct {
	StringValue string
	IntValue    int
	FloatValue  float64
}

func TestPlistConversion(t *testing.T) {
	testCases := map[string]struct {
		data     SampleData
		fileName string
	}{
		"foo": {SampleData{"d", 4, 0.2}, "sample-plist-primitives"},
	}

	for _, tc := range testCases {

		actual := usbmux.ToPlist(tc.data)
		println(actual)
		golden := filepath.Join("test-fixture", tc.fileName+".plist")
		if *update {
			err := ioutil.WriteFile(golden, []byte(actual), 0644)
			if err != nil {
				log.Fatal(err)
			}
		}
		expected, _ := ioutil.ReadFile(golden)
		assert.Equal(t, actual, string(expected))
	}

}
