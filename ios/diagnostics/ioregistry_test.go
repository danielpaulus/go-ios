package diagnostics

import (
	"bytes"
	"encoding/binary"
	"io"
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	plist "howett.net/plist"
)

// fakeDeviceConn records sent messages and replays a canned response.
type fakeDeviceConn struct {
	ios.DeviceConnectionInterface
	sent     [][]byte
	response bytes.Buffer
}

func (f *fakeDeviceConn) Send(message []byte) error {
	f.sent = append(f.sent, message)
	return nil
}

func (f *fakeDeviceConn) Reader() io.Reader { return &f.response }

// cannedIORegistryResponse is a diagnostics relay response as a real device
// sends it for an EntryClass=IOPMPowerSource query.
const cannedIORegistryResponse = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Diagnostics</key>
	<dict>
		<key>IORegistry</key>
		<dict>
			<key>CurrentCapacity</key>
			<integer>85</integer>
			<key>CycleCount</key>
			<integer>321</integer>
			<key>IsCharging</key>
			<true/>
			<key>Temperature</key>
			<integer>3000</integer>
			<key>Voltage</key>
			<integer>4269</integer>
		</dict>
	</dict>
	<key>Status</key>
	<string>Success</string>
</dict>
</plist>`

// framed prefixes payload with the 4 byte big endian length header used by
// plist based lockdown services.
func framed(payload string) []byte {
	buf := new(bytes.Buffer)
	_ = binary.Write(buf, binary.BigEndian, uint32(len(payload)))
	buf.WriteString(payload)
	return buf.Bytes()
}

func newFakeConnection() (*Connection, *fakeDeviceConn) {
	fake := &fakeDeviceConn{}
	fake.response.Write(framed(cannedIORegistryResponse))
	return &Connection{deviceConn: fake, plistCodec: ios.NewPlistCodec()}, fake
}

func TestIORegistryRequestEncoding(t *testing.T) {
	tests := []struct {
		name     string
		plane    string
		entry    string
		class    string
		expected map[string]string
	}{
		{"no filters", "", "", "", map[string]string{
			"Request": "IORegistry",
		}},
		{"plane only", "IODeviceTree", "", "", map[string]string{
			"Request": "IORegistry", "CurrentPlane": "IODeviceTree",
		}},
		{"name only", "", "AppleARMPMUCharger", "", map[string]string{
			"Request": "IORegistry", "EntryName": "AppleARMPMUCharger",
		}},
		{"class only", "", "", "IOPMPowerSource", map[string]string{
			"Request": "IORegistry", "EntryClass": "IOPMPowerSource",
		}},
		{"all", "IOService", "AppleARMPMUCharger", "IOPMPowerSource", map[string]string{
			"Request": "IORegistry", "CurrentPlane": "IOService", "EntryName": "AppleARMPMUCharger", "EntryClass": "IOPMPowerSource",
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			conn, fake := newFakeConnection()
			_, err := conn.IORegistry(tt.plane, tt.entry, tt.class)
			require.NoError(t, err)
			require.Len(t, fake.sent, 1)
			sent := fake.sent[0]
			require.Greater(t, len(sent), 4)
			assert.Equal(t, uint32(len(sent)-4), binary.BigEndian.Uint32(sent[:4]))
			var decoded map[string]string
			_, err = plist.Unmarshal(sent[4:], &decoded)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, decoded)
		})
	}
}

func TestIORegistryResponseDecoding(t *testing.T) {
	conn, _ := newFakeConnection()
	result, err := conn.IORegistry("", "", "IOPMPowerSource")
	require.NoError(t, err)
	assert.Equal(t, "Success", result["Status"])
	diag, ok := result["Diagnostics"].(map[string]interface{})
	require.True(t, ok, "Diagnostics should be a dict")
	ioreg, ok := diag["IORegistry"].(map[string]interface{})
	require.True(t, ok, "IORegistry should be a dict")
	assert.Equal(t, map[string]interface{}{
		"CurrentCapacity": uint64(85),
		"CycleCount":      uint64(321),
		"IsCharging":      true,
		"Temperature":     uint64(3000),
		"Voltage":         uint64(4269),
	}, ioreg)
}
