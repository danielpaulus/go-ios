package xpc

import (
	"bytes"
	"encoding/base64"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestEmptyDictionary(t *testing.T) {
	b, _ := os.ReadFile(path.Join("xpc_empty_dict.bin"))

	res, err := DecodeMessage(bytes.NewReader(b))
	assert.NoError(t, err)
	assert.Equal(t, Message{
		Flags: AlwaysSetFlag,
		Body:  map[string]interface{}{},
	}, res)
}

func TestDictionary(t *testing.T) {
	b, _ := os.ReadFile(path.Join("xpc_dict.bin"))

	res, err := DecodeMessage(bytes.NewReader(b))
	assert.NoError(t, err)
	assert.Equal(t, Message{
		Flags: AlwaysSetFlag | DataFlag | HeartbeatRequestFlag,
		Body: map[string]interface{}{
			"CoreDevice.CoreDeviceDDIProtocolVersion": int64(0),
			"CoreDevice.action":                       map[string]interface{}{},
			"CoreDevice.coreDeviceVersion": map[string]interface{}{
				"components":              []interface{}{uint64(0x15c), uint64(0x1), uint64(0x0), uint64(0x0), uint64(0x0)},
				"originalComponentsCount": int64(2),
				"stringValue":             "348.1",
			},
			"CoreDevice.deviceIdentifier":  "A7DD28AC-2911-4549-811D-85917B9AC72F",
			"CoreDevice.featureIdentifier": "com.apple.coredevice.feature.launchapplication",
			"CoreDevice.input": map[string]interface{}{
				"applicationSpecifier": map[string]interface{}{
					"bundleIdentifier": map[string]interface{}{
						"_0": "xxx.xxxxxxxxx.xxxxxxxx",
					},
				},
				"options": map[string]interface{}{
					"arguments": []interface{}{},
					"environmentVariables": map[string]interface{}{
						"TERM": "xterm-256color",
					},
					"platformSpecificOptions":       base64Decode("YnBsaXN0MDDQCAAAAAAAAAEBAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAJ"),
					"standardIOUsesPseudoterminals": true,
					"startStopped":                  false,
					"terminateExisting":             false,
					"user": map[string]interface{}{
						"active": true,
					},
					"workingDirectory": nil,
				},
				"standardIOIdentifiers": map[string]interface{}{},
			},
			"CoreDevice.invocationIdentifier": "62419FC1-5ABF-4D96-BCA8-7A5F6F9A69EE",
		},
	}, res)
}

func base64Decode(s string) []byte {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(s)))
	_, err := base64.StdEncoding.Decode(dst, []byte(s))
	if err != nil {
		panic(err)
	}
	return dst
}

func TestEncodeDecode(t *testing.T) {
	tests := []struct {
		name          string
		input         map[string]interface{}
		expectedFlags uint32
	}{
		{
			name:          "empty dict",
			input:         map[string]interface{}{},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name:          "no xpc body",
			input:         nil,
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name: "keys without padding",
			input: map[string]interface{}{
				"key":     "value",
				"key-key": "value",
			},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name: "nested values",
			input: map[string]interface{}{
				"key1": "string-val",
				"nested-dict": map[string]interface{}{
					"bool":   true,
					"int64":  int64(123),
					"uint64": uint64(321),
					"data":   []byte{0x1},
					"double": float64(1.2),
				},
			},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name: "null entry",
			input: map[string]interface{}{
				"null": nil,
			},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name: "dictionary with array",
			input: map[string]interface{}{
				"array": []interface{}{uint64(1), uint64(2), uint64(3)},
			},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name: "encode uuid",
			input: map[string]interface{}{
				"uuidvalue": func() uuid.UUID {
					u, _ := uuid.FromBytes(base64Decode("RYjS2yNAbEG+Y0WWxq5/4w=="))
					return u
				}(),
			},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
		{
			name: "encode uuid",
			input: map[string]interface{}{
				"uuidvalue": func() uuid.UUID {
					u, _ := uuid.FromBytes(base64Decode("RYjS2yNAbEG+Y0WWxq5/4w=="))
					return u
				}(),
			},
			expectedFlags: AlwaysSetFlag | DataFlag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(nil)
			err := EncodeMessage(buf, Message{
				Flags: AlwaysSetFlag | DataFlag,
				Body:  tt.input,
				Id:    0,
			})
			assert.NoError(t, err)
			res, err := DecodeMessage(buf)
			assert.NoError(t, err)
			assert.Equal(t, tt.input, res.Body)
			assert.Equal(t, tt.expectedFlags, res.Flags)
		})
	}
}
