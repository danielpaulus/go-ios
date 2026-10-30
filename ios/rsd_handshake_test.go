package ios

import (
	"bytes"
	"io"
	"testing"

	"github.com/danielpaulus/go-ios/ios/xpc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// splitStream reads the device's messages from one buffer and records what the
// host writes in another.
type splitStream struct {
	io.Reader
	written bytes.Buffer
}

func (s *splitStream) Write(p []byte) (int, error) { return s.written.Write(p) }

func TestHandshakeAnnouncesTheHostVersionFlags(t *testing.T) {
	const versionFlags = uint64(0x0100000000000006)
	var device bytes.Buffer
	require.NoError(t, xpc.EncodeMessage(&device, xpc.Message{Flags: xpc.AlwaysSetFlag | xpc.DataFlag, Id: 1, Body: map[string]interface{}{
		"MessageType":              "Handshake",
		"MessagingProtocolVersion": uint64(7),
		"Properties": map[string]interface{}{
			"RemoteXPCVersionFlags": versionFlags,
			"UniqueDeviceID":        "00000000-0000000000000000",
		},
		"Services": map[string]interface{}{
			"com.apple.coredevice.screencaptureservice": map[string]interface{}{"Port": "53541"},
		},
		"UUID": uuid.New(),
	}}))
	stream := &splitStream{Reader: &device}
	conn, err := xpc.New(stream, &bytes.Buffer{}, nil, io.NopCloser(nil))
	require.NoError(t, err)

	response, err := RsdService{xpc: conn, c: io.NopCloser(nil)}.Handshake()

	require.NoError(t, err)
	assert.Equal(t, 53541, response.GetPort("com.apple.coredevice.screencaptureservice"))

	// Without the host's flags the device treats it as a legacy peer and traps
	// when a reply carries a large data object, e.g. a screenshot.
	host, err := xpc.DecodeMessage(&stream.written)
	require.NoError(t, err)
	assert.Equal(t, "Handshake", host.Body["MessageType"])
	assert.Equal(t, uint64(7), host.Body["MessagingProtocolVersion"])
	properties, ok := host.Body["Properties"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, versionFlags, properties["RemoteXPCVersionFlags"])
	assert.NotEqual(t, uuid.Nil, host.Body["UUID"])
}
