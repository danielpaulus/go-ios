package nskeyedarchiver_test

import (
	"strings"
	"testing"

	archiver "github.com/danielpaulus/go-ios/ios/nskeyedarchiver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// selfReferentialArrayArchive is a hostile NSKeyedArchiver plist whose root
// object is an NSArray whose NS.objects entry points back at the NSArray itself
// (UID 1 -> UID 1). Before the depth guard in extractObjects, decoding this
// recursed forever and triggered an uncatchable Go "fatal error: stack
// overflow" that the recover() in Unarchive could NOT catch. With the fix it
// returns an error.
const selfReferentialArrayArchive = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>$archiver</key>
	<string>NSKeyedArchiver</string>
	<key>$objects</key>
	<array>
		<string>$null</string>
		<dict>
			<key>$class</key>
			<dict>
				<key>CF$UID</key>
				<integer>2</integer>
			</dict>
			<key>NS.objects</key>
			<array>
				<dict>
					<key>CF$UID</key>
					<integer>1</integer>
				</dict>
			</array>
		</dict>
		<dict>
			<key>$classes</key>
			<array>
				<string>NSArray</string>
				<string>NSObject</string>
			</array>
			<key>$classname</key>
			<string>NSArray</string>
		</dict>
	</array>
	<key>$top</key>
	<dict>
		<key>root</key>
		<dict>
			<key>CF$UID</key>
			<integer>1</integer>
		</dict>
	</dict>
	<key>$version</key>
	<integer>100000</integer>
</dict>
</plist>`

// mutualRecursiveDictArchive is a hostile archive where an NSDictionary value
// (UID 1) resolves to an NSArray (UID 3) whose only element points back at the
// dictionary (UID 1), forming an A->B->A cycle across container types.
const mutualRecursiveDictArchive = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>$archiver</key>
	<string>NSKeyedArchiver</string>
	<key>$objects</key>
	<array>
		<string>$null</string>
		<dict>
			<key>$class</key>
			<dict>
				<key>CF$UID</key>
				<integer>2</integer>
			</dict>
			<key>NS.keys</key>
			<array>
				<dict>
					<key>CF$UID</key>
					<integer>4</integer>
				</dict>
			</array>
			<key>NS.objects</key>
			<array>
				<dict>
					<key>CF$UID</key>
					<integer>3</integer>
				</dict>
			</array>
		</dict>
		<dict>
			<key>$classes</key>
			<array>
				<string>NSDictionary</string>
				<string>NSObject</string>
			</array>
			<key>$classname</key>
			<string>NSDictionary</string>
		</dict>
		<dict>
			<key>$class</key>
			<dict>
				<key>CF$UID</key>
				<integer>5</integer>
			</dict>
			<key>NS.objects</key>
			<array>
				<dict>
					<key>CF$UID</key>
					<integer>1</integer>
				</dict>
			</array>
		</dict>
		<string>key</string>
		<dict>
			<key>$classes</key>
			<array>
				<string>NSArray</string>
				<string>NSObject</string>
			</array>
			<key>$classname</key>
			<string>NSArray</string>
		</dict>
	</array>
	<key>$top</key>
	<dict>
		<key>root</key>
		<dict>
			<key>CF$UID</key>
			<integer>1</integer>
		</dict>
	</dict>
	<key>$version</key>
	<integer>100000</integer>
</dict>
</plist>`

// TestUnarchiveSelfReferentialArrayReturnsError covers CRIT-2: a self-referential
// NS.objects cycle must return an error instead of crashing the process with an
// uncatchable stack overflow.
func TestUnarchiveSelfReferentialArrayReturnsError(t *testing.T) {
	_, err := archiver.Unarchive([]byte(selfReferentialArrayArchive))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max unarchive depth")
}

// TestUnarchiveMutualRecursionReturnsError covers CRIT-2 for an A->B->A cycle
// that crosses container types (NSDictionary -> NSArray -> NSDictionary).
func TestUnarchiveMutualRecursionReturnsError(t *testing.T) {
	_, err := archiver.Unarchive([]byte(mutualRecursiveDictArchive))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "max unarchive depth")
}

// TestUnarchiveWrongTypedHeadersReturnError covers LOW-2: header fields with the
// wrong type must be rejected with a descriptive error rather than panicking
// (previously only caught by the recover backstop).
func TestUnarchiveWrongTypedHeadersReturnError(t *testing.T) {
	testCases := map[string]string{
		"archiver is not a string": `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>$archiver</key>
	<integer>42</integer>
	<key>$objects</key>
	<array><string>$null</string></array>
	<key>$top</key>
	<dict><key>$0</key><dict><key>CF$UID</key><integer>0</integer></dict></dict>
	<key>$version</key>
	<integer>100000</integer>
</dict>
</plist>`,
		"version is not a uint64": `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>$archiver</key>
	<string>NSKeyedArchiver</string>
	<key>$objects</key>
	<array><string>$null</string></array>
	<key>$top</key>
	<dict><key>$0</key><dict><key>CF$UID</key><integer>0</integer></dict></dict>
	<key>$version</key>
	<string>notanumber</string>
</dict>
</plist>`,
		"top is not a dictionary": `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>$archiver</key>
	<string>NSKeyedArchiver</string>
	<key>$objects</key>
	<array><string>$null</string></array>
	<key>$top</key>
	<string>not-a-dict</string>
	<key>$version</key>
	<integer>100000</integer>
</dict>
</plist>`,
	}

	for name, xml := range testCases {
		t.Run(name, func(t *testing.T) {
			_, err := archiver.Unarchive([]byte(xml))
			require.Error(t, err)
			// The error must not be an uncaught panic surfaced via the recover
			// backstop; the comma-ok guards return a descriptive message.
			assert.NotContains(t, err.Error(), "goroutine")
		})
	}
}

// TestUnarchiveValidArchiveStillDecodes pins the happy-path: a well-formed
// archive round-trips identically after the hardening changes.
func TestUnarchiveValidArchiveStillDecodes(t *testing.T) {
	arr := []interface{}{"a", "n", "c"}
	b, err := archiver.ArchiveXML(arr)
	require.NoError(t, err)
	require.False(t, strings.Contains(string(b), "max unarchive depth"))

	result, err := archiver.Unarchive([]byte(b))
	require.NoError(t, err)
	assert.Equal(t, arr, result[0])
}
