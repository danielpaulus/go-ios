package ios_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var (
	update      = flag.Bool("update", false, "update golden files")
	integration = flag.Bool("integration", false, "run integration tests")
)

type SampleData struct {
	StringValue string
	IntValue    int
	FloatValue  float64
}

func TestGenericSliceToType(t *testing.T) {
	slice := []interface{}{5, 3, 2}
	v, err := ios.GenericSliceToType[int](slice)
	assert.Nil(t, err)
	assert.Equal(t, 3, v[1])
	_, err = ios.GenericSliceToType[string](slice)
	assert.NotNil(t, err)
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
			err := os.WriteFile(golden, []byte(actual), 0o644)
			if err != nil {
				log.Error(err)
				t.FailNow()
			}
		}
		expected, _ := os.ReadFile(golden)
		assert.Equal(t, removeLineBreaks(string(expected)), removeLineBreaks(actual))
	}
}

// needed for windows support. Without i, we would have different linebreaks with n and with rn
// and the test would fail.
func removeLineBreaks(s string) string {
	s = strings.Replace(s, "\n", "", -1)
	s = strings.Replace(s, "\r", "", -1)
	return s
}
