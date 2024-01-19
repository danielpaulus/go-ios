package opack

import (
	"bytes"
	"encoding/base64"
	"github.com/stretchr/testify/assert"
	"howett.net/plist"
	"io"
	"os/exec"
	"testing"
)

var decoderInitialized = false

func TestEncode(t *testing.T) {
	if !decoderInitialized {
		t.SkipNow()
	}
	tests := []struct {
		name  string
		input map[string]interface{}
	}{
		{
			name:  "empty dict",
			input: map[string]interface{}{},
		},
		{
			name: "single string",
			input: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name: "middle string",
			input: map[string]interface{}{
				"key": "some longer string",
			},
		},
		{
			name: "data",
			input: map[string]interface{}{
				"key": []byte{0xCD, 0xFF, 0xAB},
			},
		},
		{
			name: "device info",
			input: map[string]interface{}{
				"accountID":                   "BB559933-AA88-4499-BB88-442266EECCFF",
				"altIRK":                      []byte{0x5e, 0xca, 0x81, 0x91, 0x92, 0x02, 0x82, 0x00, 0x11, 0x22, 0x33, 0x44, 0xbb, 0xf2, 0x4a, 0xc8},
				"btAddr":                      "FF:DD:99:66:BB:AA",
				"mac":                         []byte{0xff, 0x44, 0x88, 0x66, 0x33, 0x99},
				"model":                       "MacBookPro18,3",
				"name":                        "host-name",
				"remotepairing_serial_number": "YY9944YY99",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := Encode(tt.input)
			assert.NoError(t, err)
			reference := decodeWithReference(encoded)
			assert.Equal(t, reference, tt.input)
		})
	}
}

func decodeWithReference(b []byte) map[string]interface{} {
	cmd := exec.Command("./decode")
	w, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}
	str := base64.StdEncoding.EncodeToString(b)
	r, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	_, err = w.Write([]byte(str))
	if err != nil {
		panic(err)
	}
	_ = w.Close()
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}
	err = cmd.Wait()
	if err != nil {
		panic(err)
	}
	if cmd.ProcessState.ExitCode() != 0 {
		panic("not 0")
	}

	buf := bytes.NewReader(out)
	dec := plist.NewDecoder(buf)
	var decoded map[string]interface{}
	err = dec.Decode(&decoded)
	if err != nil {
		panic(err)
	}
	return decoded
}

func init() {
	cmd := exec.Command("xcrun",
		"clang",
		"-DDECODE",
		"-fobjc-arc",
		"-fmodules",
		"-F", "/System/Library/PrivateFrameworks/",
		"-framework", "CoreUtils",
		"main.m", "--output", "decode",
	)
	err := cmd.Run()
	if err != nil {
		return
	}
	decoderInitialized = true
}
