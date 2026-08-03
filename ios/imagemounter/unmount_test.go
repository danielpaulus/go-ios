package imagemounter

import (
	"bytes"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// encodeUnmountResponse frames a plist message the same way the device does, so that it can be
// read back through the regular codec.
func encodeUnmountResponse(t *testing.T, msg map[string]interface{}) *bytes.Buffer {
	t.Helper()
	buf := new(bytes.Buffer)
	require.NoError(t, ios.NewPlistCodecReadWriter(nil, buf).Write(msg))
	return buf
}

func TestReadUnmountResponse(t *testing.T) {
	tests := []struct {
		name        string
		response    map[string]interface{}
		wantErr     bool
		errContains string
	}{
		{
			name:     "unmount completed",
			response: map[string]interface{}{"Status": "Complete"},
		},
		{
			name:        "device is locked",
			response:    map[string]interface{}{"Error": "DeviceLocked"},
			wantErr:     true,
			errContains: "DeviceLocked",
		},
		{
			name:        "no image mounted",
			response:    map[string]interface{}{"Error": "ImageNotMounted"},
			wantErr:     true,
			errContains: "ImageNotMounted",
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
			err := readUnmountResponse(ios.NewPlistCodecReadWriter(encodeUnmountResponse(t, tt.response), nil))

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errContains)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// A device that refuses to unmount must not be reported as a successful unmount.
func TestUnmountImageReportsDeviceError(t *testing.T) {
	deviceLocked := map[string]interface{}{"Error": "DeviceLocked"}

	t.Run("personalized image mounter", func(t *testing.T) {
		written := new(bytes.Buffer)
		mounter := PersonalizedDeveloperDiskImageMounter{
			plistRw: ios.NewPlistCodecReadWriter(encodeUnmountResponse(t, deviceLocked), written),
		}

		err := mounter.UnmountImage()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "DeviceLocked")
		assert.Contains(t, written.String(), "UnmountImage", "the 'UnmountImage' command should have been sent")
	})

	t.Run("developer disk image mounter", func(t *testing.T) {
		written := new(bytes.Buffer)
		mounter := DeveloperDiskImageMounter{
			plistRw: ios.NewPlistCodecReadWriter(encodeUnmountResponse(t, deviceLocked), written),
		}

		err := mounter.UnmountImage()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "DeviceLocked")
		assert.Contains(t, written.String(), "UnmountImage", "the 'UnmountImage' command should have been sent")
	})
}
