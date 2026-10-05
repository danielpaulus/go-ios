package imagemounter

import (
	"bytes"
	"errors"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The replies below were captured from real devices, so that the tests assert against what
// devices actually send rather than invented shapes.
var (
	// iOS 26.5.2, mounting with Developer Mode turned off.
	developerModeDisabledReply = map[string]interface{}{
		"Error": "ImageMountFailed",
		"DetailedError": `Error Domain=com.apple.MobileStorage.ErrorDomain Code=-2 "Failed to mount /private/var/.../CAKBOG.dmg." ` +
			`UserInfo={NSUnderlyingError=Code=-4 "Developer mode is not enabled."}`,
	}
	// iOS 26.5.2, unmounting while the device is locked.
	deviceLockedReply = map[string]interface{}{"Error": "DeviceLocked"}
	// iOS 14.3, unmounting with nothing mounted.
	nothingMountedReply = map[string]interface{}{
		"Error": "InternalError",
		"DetailedError": `Error Domain=com.apple.MobileStorage.ErrorDomain Code=-2 "Failed to unmount /Developer." ` +
			`UserInfo={NSLocalizedDescription=There is no matching entry in the device map for /Developer.}`,
	}
	// iOS 26.6, mounting right after boot: the personalized DDI had already
	// auto-remounted (Developer Mode persists it across reboots) faster than
	// ListImages() could observe it.
	alreadyMountedReply = map[string]interface{}{
		"Error": "ImageMountFailed",
		"DetailedError": `Error Domain=com.apple.MobileStorage.ErrorDomain Code=-2 "Failed to mount /private/.../ltFdWi.dmg." ` +
			`UserInfo={NSUnderlyingError=Code=-2 "Invalid value for MountPath: Error Domain=com.apple.MobileStorage.ErrorDomain Code=-3 ` +
			`\"A disk image of type Personalized/DeveloperDiskImage is already mounted at /System/Developer.\""}`,
	}
)

// encodePlistMessage frames a plist message the same way the device does, so that it can be read
// back through the regular codec.
func encodePlistMessage(t *testing.T, msg map[string]interface{}) *bytes.Buffer {
	t.Helper()
	buf := new(bytes.Buffer)
	require.NoError(t, ios.NewPlistCodecReadWriter(nil, buf).Write(msg))
	return buf
}

func readerFor(t *testing.T, reply map[string]interface{}) ios.PlistCodecReadWriter {
	t.Helper()
	return ios.NewPlistCodecReadWriter(encodePlistMessage(t, reply), nil)
}

func TestReadImageMounterResponse(t *testing.T) {
	tests := []struct {
		name           string
		reply          map[string]interface{}
		expectedStatus string
		wantErr        bool
		errContains    []string
	}{
		{
			name:           "mount completed",
			reply:          map[string]interface{}{"Status": "Complete"},
			expectedStatus: "Complete",
		},
		{
			name:           "upload acknowledged",
			reply:          map[string]interface{}{"Status": "ReceiveBytesAck"},
			expectedStatus: "ReceiveBytesAck",
		},
		{
			name:           "developer mode is not enabled",
			reply:          developerModeDisabledReply,
			expectedStatus: "Complete",
			wantErr:        true,
			errContains:    []string{"ImageMountFailed", "Developer mode is not enabled."},
		},
		{
			name:           "device is locked",
			reply:          deviceLockedReply,
			expectedStatus: "Complete",
			wantErr:        true,
			errContains:    []string{"DeviceLocked"},
		},
		{
			name:           "nothing mounted",
			reply:          nothingMountedReply,
			expectedStatus: "Complete",
			wantErr:        true,
			errContains:    []string{"InternalError", "no matching entry in the device map"},
		},
		{
			name:           "wrong status for the command",
			reply:          map[string]interface{}{"Status": "ReceiveBytesAck"},
			expectedStatus: "Complete",
			wantErr:        true,
			errContains:    []string{"unexpected response"},
		},
		{
			name:           "empty response",
			reply:          map[string]interface{}{},
			expectedStatus: "Complete",
			wantErr:        true,
			errContains:    []string{"unexpected response"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := readImageMounterResponse(readerFor(t, tt.reply), "udid", "MountImage", tt.expectedStatus)

			if !tt.wantErr {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			for _, want := range tt.errContains {
				assert.Contains(t, err.Error(), want)
			}
		})
	}
}

func TestAlreadyMounted(t *testing.T) {
	err := readImageMounterResponse(readerFor(t, alreadyMountedReply), "udid", "MountImage", "Complete")
	require.Error(t, err)
	assert.True(t, alreadyMounted(err))

	assert.False(t, alreadyMounted(readImageMounterResponse(readerFor(t, developerModeDisabledReply), "udid", "MountImage", "Complete")))
	assert.False(t, alreadyMounted(readImageMounterResponse(readerFor(t, deviceLockedReply), "udid", "MountImage", "Complete")))
	assert.False(t, alreadyMounted(nil))
	assert.False(t, alreadyMounted(errors.New("some other error")))
}

// Devices that predate 'UnmountImage' answer with UnknownCommand. Unmounting stays best-effort
// there instead of turning into a hard failure.
func TestReadUnmountResponseToleratesUnknownCommand(t *testing.T) {
	unknown := map[string]interface{}{"Error": "UnknownCommand"}

	require.NoError(t, readUnmountResponse(readerFor(t, unknown), "udid"))
	require.Error(t, readUnmountResponse(readerFor(t, deviceLockedReply), "udid"))
}

// A device that refuses a mount or unmount must not be reported as a success.
func TestMountersReportDeviceErrors(t *testing.T) {
	t.Run("personalized mount", func(t *testing.T) {
		written := new(bytes.Buffer)
		mounter := PersonalizedDeveloperDiskImageMounter{
			plistRw: ios.NewPlistCodecReadWriter(encodePlistMessage(t, developerModeDisabledReply), written),
		}

		err := mounter.mountPersonalizedImage([]byte("signature"), []byte("trust-cache"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "Developer mode is not enabled.")
		assert.Contains(t, written.String(), "MountImage")
	})

	t.Run("developer disk image mount", func(t *testing.T) {
		written := new(bytes.Buffer)
		mounter := DeveloperDiskImageMounter{
			plistRw: ios.NewPlistCodecReadWriter(encodePlistMessage(t, deviceLockedReply), written),
		}

		err := mounter.mountImage([]byte("signature"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "DeviceLocked")
		assert.Contains(t, written.String(), "MountImage")
	})

	t.Run("personalized unmount", func(t *testing.T) {
		written := new(bytes.Buffer)
		mounter := PersonalizedDeveloperDiskImageMounter{
			plistRw: ios.NewPlistCodecReadWriter(encodePlistMessage(t, deviceLockedReply), written),
		}

		err := mounter.UnmountImage()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "DeviceLocked")
		assert.Contains(t, written.String(), "UnmountImage")
	})

	t.Run("developer disk image unmount", func(t *testing.T) {
		written := new(bytes.Buffer)
		mounter := DeveloperDiskImageMounter{
			plistRw: ios.NewPlistCodecReadWriter(encodePlistMessage(t, nothingMountedReply), written),
		}

		err := mounter.UnmountImage()

		require.Error(t, err)
		assert.Contains(t, err.Error(), "no matching entry in the device map")
		assert.Contains(t, written.String(), "UnmountImage")
	})
}
