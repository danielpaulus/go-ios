package ios_test

import (
	"flag"
	"io/ioutil"
	"path/filepath"
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var update = flag.Bool("update", false, "update golden files")
var integration = flag.Bool("integration", false, "run integration tests")

type SampleData struct {
	StringValue string
	IntValue    int
	FloatValue  float64
}

func TestNtohs(t *testing.T) {

	assert.Equal(t, uint16(62078), ios.Ntohs(ios.Lockdownport))
}

func TestPlistConversion(t *testing.T) {
	testCases := map[string]struct {
		data     interface{}
		fileName string
	}{
		"randomData":     {SampleData{"d", 4, 0.2}, "sample-plist-primitives"},
		"UsbMuxResponse": {ios.MuxResponse{"ErrorName", 5}, "usbmuxresponse"},
	}

	for _, tc := range testCases {

		actual := ios.ToPlist(tc.data)

		golden := filepath.Join("test-fixture", tc.fileName+".plist")
		if *update {
			err := ioutil.WriteFile(golden, []byte(actual), 0644)
			if err != nil {
				log.Fatal(err)
			}
		}
		expected, _ := ioutil.ReadFile(golden)
		assert.Equal(t, string(expected), actual)
	}

}
