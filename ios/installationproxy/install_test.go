package installationproxy

import (
	"bytes"
	"testing"

	ios "github.com/danielpaulus/go-ios/ios"
	"github.com/stretchr/testify/assert"
	"howett.net/plist"
)

// fakeConn is an in-memory stand-in for the device connection: it serves
// canned responses from 'in' and records everything the client sends in 'out'.
type fakeConn struct {
	in  *bytes.Buffer
	out *bytes.Buffer
}

func (f *fakeConn) Read(p []byte) (int, error)  { return f.in.Read(p) }
func (f *fakeConn) Write(p []byte) (int, error) { return f.out.Write(p) }
func (f *fakeConn) Close() error                { return nil }

// cannedResponses encodes the given dicts in the length-prefixed plist wire
// format that installation_proxy uses.
func cannedResponses(t *testing.T, responses ...map[string]interface{}) *bytes.Buffer {
	t.Helper()
	codec := ios.NewPlistCodec()
	buf := &bytes.Buffer{}
	for _, r := range responses {
		b, err := codec.Encode(r)
		assert.NoError(t, err)
		buf.Write(b)
	}
	return buf
}

func newTestConnection(in *bytes.Buffer) (*Connection, *fakeConn) {
	rwc := &fakeConn{in: in, out: &bytes.Buffer{}}
	return &Connection{
		deviceConn: ios.NewDeviceConnectionWithRWC(rwc),
		plistCodec: ios.NewPlistCodec(),
	}, rwc
}

func TestInstallCommandRequestPlist(t *testing.T) {
	cmd := installCommand("PublicStaging/wda.ipa", nil)
	assert.Equal(t, "Install", cmd["Command"])
	assert.Equal(t, "PublicStaging/wda.ipa", cmd["PackagePath"])
	// ClientOptions must always be present, an empty dict when no options are given
	assert.Equal(t, map[string]interface{}{}, cmd["ClientOptions"])

	cmd = installCommand("PublicStaging/app.ipa", map[string]interface{}{"PackageType": "Developer"})
	assert.Equal(t, map[string]interface{}{"PackageType": "Developer"}, cmd["ClientOptions"])
}

func TestInstallSendsRequestAndStreamsProgress(t *testing.T) {
	conn, rwc := newTestConnection(cannedResponses(t,
		map[string]interface{}{"Status": "CreatingStagingDirectory", "PercentComplete": 5},
		map[string]interface{}{"Status": "InstallingApplication", "PercentComplete": 60},
		map[string]interface{}{"Status": "Complete"},
	))

	err := conn.Install("PublicStaging/wda.ipa", nil)
	assert.NoError(t, err)

	// decode the request that was sent over the wire and verify the plist content
	sent, err := ios.NewPlistCodec().Decode(bytes.NewReader(rwc.out.Bytes()))
	assert.NoError(t, err)
	var request map[string]interface{}
	_, err = plist.Unmarshal(sent, &request)
	assert.NoError(t, err)
	assert.Equal(t, "Install", request["Command"])
	assert.Equal(t, "PublicStaging/wda.ipa", request["PackagePath"])
	assert.Equal(t, map[string]interface{}{}, request["ClientOptions"])
}

func TestInstallErrorResponse(t *testing.T) {
	conn, _ := newTestConnection(cannedResponses(t,
		map[string]interface{}{"Status": "VerifyingApplication", "PercentComplete": 40},
		map[string]interface{}{"Error": "ApplicationVerificationFailed", "ErrorDescription": "invalid signature"},
	))

	err := conn.Install("PublicStaging/wda.ipa", nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ApplicationVerificationFailed")
	assert.Contains(t, err.Error(), "invalid signature")
}

func TestEvaluateInstallProgressFromCannedPlists(t *testing.T) {
	progressPlist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
	<key>PercentComplete</key><integer>30</integer>
	<key>Status</key><string>VerifyingApplication</string>
</dict></plist>`
	dict, err := ios.ParsePlist([]byte(progressPlist))
	assert.NoError(t, err)
	done, percent, status, err := evaluateInstallProgress(dict)
	assert.NoError(t, err)
	assert.False(t, done)
	assert.Equal(t, 30, percent)
	assert.Equal(t, "VerifyingApplication", status)

	completePlist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
	<key>Status</key><string>Complete</string>
</dict></plist>`
	dict, err = ios.ParsePlist([]byte(completePlist))
	assert.NoError(t, err)
	done, percent, _, err = evaluateInstallProgress(dict)
	assert.NoError(t, err)
	assert.True(t, done)
	assert.Equal(t, 100, percent)

	errorPlist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
	<key>Error</key><string>ApplicationVerificationFailed</string>
	<key>ErrorDescription</key><string>failed to verify code signature</string>
</dict></plist>`
	dict, err = ios.ParsePlist([]byte(errorPlist))
	assert.NoError(t, err)
	_, _, _, err = evaluateInstallProgress(dict)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ApplicationVerificationFailed")
	assert.Contains(t, err.Error(), "failed to verify code signature")
}

func TestEvaluateInstallProgressUnknownUpdate(t *testing.T) {
	_, _, _, err := evaluateInstallProgress(map[string]interface{}{"Foo": "Bar"})
	assert.Error(t, err)
}
