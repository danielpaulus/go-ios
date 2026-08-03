package imagemounter

import (
	"bytes"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// encodePlistMessage frames a plist message the same way the device does, so that it can be
// read back through the regular codec.
func encodePlistMessage(t *testing.T, msg map[string]interface{}) *bytes.Buffer {
	t.Helper()
	buf := new(bytes.Buffer)
	require.NoError(t, ios.NewPlistCodecReadWriter(nil, buf).Write(msg))
	return buf
}

func TestMountPersonalizedImage(t *testing.T) {
	tests := []struct {
		name        string
		response    map[string]interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:     "mount completed",
			response: map[string]interface{}{"Status": "Complete"},
		},
		{
			name: "developer mode is not enabled",
			response: map[string]interface{}{
				"Error":         "ImageMountFailed",
				"DetailedError": `Error Domain=com.apple.MobileStorage Code=-2 "Failed to mount image" UserInfo={NSUnderlyingError=Code=-4 "Developer mode is not enabled."}`,
			},
			wantErr:     true,
			errContains: "Developer mode is not enabled.",
		},
		{
			name:        "device is locked",
			response:    map[string]interface{}{"Error": "DeviceLocked"},
			wantErr:     true,
			errContains: "DeviceLocked",
		},
		{
			name:        "unexpected status",
			response:    map[string]interface{}{"Status": "Failure"},
			wantErr:     true,
			errContains: "Failure",
		},
		{
			name:        "empty response",
			response:    map[string]interface{}{},
			wantErr:     true,
			errContains: "unexpected response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			written := new(bytes.Buffer)
			mounter := PersonalizedDeveloperDiskImageMounter{
				plistRw: ios.NewPlistCodecReadWriter(encodePlistMessage(t, tt.response), written),
			}

			err := mounter.mountPersonalizedImage([]byte("signature"), []byte("trust-cache"))

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}

			assert.Contains(t, written.String(), "MountImage", "the 'MountImage' command should have been sent")
		})
	}
}
