package ios_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestPlistCodec(t *testing.T) {
	codec := ios.NewPlistCodec()
	testCases := map[string]struct {
		data     interface{}
		fileName string
	}{
		"BasebandKeyHashInformationType example": {ios.BasebandKeyHashInformationType{5, make([]byte, 1), 4}, "sample-plist-plistcodec-basebandkeyhashinfotype"},
	}

	for _, tc := range testCases {
		golden := filepath.Join("test-fixture", tc.fileName+".plist")
		actual, err := codec.Encode(tc.data)
		if assert.NoError(t, err) {
			if *update {
				err := os.WriteFile(golden, []byte(actual), 0o644)
				if err != nil {
					log.Error(err)
					t.Fail()
				}
			}
			expected, _ := os.ReadFile(golden)
			assert.Equal(t, expected, actual)

			// simple test to check that decode(encode(x))==x

			result, err := codec.Decode(bytes.NewReader(actual))
			assert.NoError(t, err)
			assert.Equal(t, ios.ToPlist(tc.data), string(result))

		}
	}
}
