package ios_test

import (
	"reflect"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
)

// TestPairRecordfromBytesGarbage ensures that decoding invalid plist bytes
// returns a zero-value PairRecord instead of panicking.
func TestPairRecordfromBytesGarbage(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{name: "garbage bytes", input: []byte{0x00, 0x01, 0x02, 0xff, 0xfe}},
		{name: "empty bytes", input: []byte{}},
		{name: "non plist text", input: []byte("this is not a plist")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ios.PairRecordfromBytes(tt.input)
			if !reflect.DeepEqual(got, ios.PairRecord{}) {
				t.Errorf("PairRecordfromBytes(%v) = %+v, want zero value", tt.input, got)
			}
		})
	}
}
