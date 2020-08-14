package usbmux_test

/*
import (
	"bytes"
	"github.com/danielpaulus/go-ios/usbmux"
	"io/ioutil"
	"path/filepath"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPlistCodec(t *testing.T) {
	responseChannel := make(chan []byte)
	codec := usbmux.NewPlistCodec(responseChannel)
	testCases := map[string]struct {
		data     interface{}
		fileName string
	}{
		"BasebandKeyHashInformationType example": {usbmux.BasebandKeyHashInformationType{5, make([]byte, 1), 4}, "sample-plist-plistcodec-basebandkeyhashinfotype"},
	}

	for _, tc := range testCases {
		golden := filepath.Join("test-fixture", tc.fileName+".plist")
		actual, err := codec.Encode(tc.data)
		if assert.NoError(t, err) {
			if *update {
				err := ioutil.WriteFile(golden, []byte(actual), 0644)
				if err != nil {
					log.Fatal(err)
				}
			}
			expected, _ := ioutil.ReadFile(golden)
			assert.Equal(t, expected, actual)

			//simple test to check that decode(encode(x))==x
			go func() {
				err := codec.Decode(bytes.NewReader(actual))
				assert.NoError(t, err)
			}()
			result := <-codec.ResponseChannel
			assert.Equal(t, usbmux.ToPlist(tc.data), string(result))

		}
	}

}
*/
