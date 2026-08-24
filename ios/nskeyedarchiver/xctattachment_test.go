package nskeyedarchiver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"howett.net/plist"
)

// iOS 27 sends XCUITest attachments with no userInfo, which NSKeyedArchiver
// encodes as the string "$null" rather than a dictionary. NewXCTAttachment used
// to assert a dictionary there, so decoding panicked; Unarchive recovered it as
// "Unarchive: interface conversion: interface {} is string, not
// map[string]interface {}", which failed the whole XCUITest run and left devices
// stuck in cleaning.
func TestNewXCTAttachmentWithNullUserInfo(t *testing.T) {
	objects := []interface{}{
		"$null",              // 0
		"public.plain-text",  // 1 uniformTypeIdentifier
		"override.txt",       // 2 fileNameOverride
		float64(770474977.9), // 3 timestamp
		"Screenshot",         // 4 name
		[]uint8{0x61, 0x62},  // 5 payload
	}
	object := map[string]interface{}{
		"lifetime":              uint64(1),
		"uniformTypeIdentifier": plist.UID(1),
		"fileNameOverride":      plist.UID(2),
		"timestamp":             plist.UID(3),
		"name":                  plist.UID(4),
		"payload":               plist.UID(5),
		"userInfo":              plist.UID(0), // points at "$null"
	}

	var decoded interface{}
	require.NotPanics(t, func() { decoded = NewXCTAttachment(object, objects) })

	attachment, ok := decoded.(XCTAttachment)
	require.True(t, ok)
	assert.Equal(t, "public.plain-text", attachment.UniformTypeIdentifier)
	assert.Equal(t, "Screenshot", attachment.Name)
	assert.Equal(t, []uint8{0x61, 0x62}, attachment.Payload)
	assert.Nil(t, attachment.userInfo, "a missing userInfo should decode as empty, not panic")
}

// A real userInfo dictionary must still be decoded.
func TestNewXCTAttachmentWithUserInfoDictionary(t *testing.T) {
	objects := []interface{}{
		"$null",
		"public.png",
		"shot.png",
		float64(1),
		"Shot",
		[]uint8{0x01},
		map[string]interface{}{"NS.keys": []interface{}{}, "NS.objects": []interface{}{}},
	}
	object := map[string]interface{}{
		"lifetime":              uint64(2),
		"uniformTypeIdentifier": plist.UID(1),
		"fileNameOverride":      plist.UID(2),
		"timestamp":             plist.UID(3),
		"name":                  plist.UID(4),
		"payload":               plist.UID(5),
		"userInfo":              plist.UID(6),
	}

	var decoded interface{}
	require.NotPanics(t, func() { decoded = NewXCTAttachment(object, objects) })
	attachment, ok := decoded.(XCTAttachment)
	require.True(t, ok)
	assert.Equal(t, "public.png", attachment.UniformTypeIdentifier)
	assert.NotNil(t, attachment.userInfo)
}

// Malformed archives must not panic either: missing keys, references that are
// not references, and references past the end of the object table.
func TestNewXCTAttachmentToleratesMalformedArchives(t *testing.T) {
	objects := []interface{}{"$null", "public.png"}

	tests := []struct {
		name   string
		object map[string]interface{}
	}{
		{name: "empty object", object: map[string]interface{}{}},
		{
			name: "reference out of range",
			object: map[string]interface{}{
				"lifetime": uint64(1),
				"name":     plist.UID(99),
			},
		},
		{
			name: "value is not a reference",
			object: map[string]interface{}{
				"lifetime": uint64(1),
				"name":     "not-a-uid",
				"userInfo": 42,
			},
		},
		{
			name: "lifetime has the wrong type",
			object: map[string]interface{}{
				"lifetime": "not-a-number",
				"name":     plist.UID(1),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotPanics(t, func() { NewXCTAttachment(tt.object, objects) })
		})
	}
}

func TestResolveRef(t *testing.T) {
	objects := []interface{}{"zero", "one"}

	assert.Equal(t, "one", resolveRef(map[string]interface{}{"k": plist.UID(1)}, "k", objects))
	assert.Nil(t, resolveRef(map[string]interface{}{}, "missing", objects))
	assert.Nil(t, resolveRef(map[string]interface{}{"k": "not-a-uid"}, "k", objects))
	assert.Nil(t, resolveRef(map[string]interface{}{"k": plist.UID(5)}, "k", objects),
		"a reference past the end of the object table must not index out of range")
}

// The same "$null" shape reaches XCTIssue and its nested source code context:
// an issue without a source-code context, or a context without a location,
// panicked in the same way and accounted for the rest of the failures.
func TestNewXCTIssueWithNullSourceCodeContext(t *testing.T) {
	objects := []interface{}{"$null", "compact", "detailed"}
	object := map[string]interface{}{
		"runtimeIssueSeverity": uint64(1),
		"compact-description":  plist.UID(1),
		"detailed-description": plist.UID(2),
		"source-code-context":  plist.UID(0), // "$null"
	}

	var decoded interface{}
	require.NotPanics(t, func() { decoded = NewXCTIssue(object, objects) })

	issue, ok := decoded.(XCTIssue)
	require.True(t, ok)
	assert.Equal(t, "compact", issue.CompactDescription)
	assert.Equal(t, "detailed", issue.DetailedDescription)
}

func TestNewXCTSourceCodeContextWithNullLocation(t *testing.T) {
	objects := []interface{}{"$null"}

	require.NotPanics(t, func() {
		NewXCTSourceCodeContext(map[string]interface{}{"location": plist.UID(0)}, objects)
	})
	require.NotPanics(t, func() {
		NewXCTSourceCodeContext(map[string]interface{}{}, objects)
	})
}

func TestNewXCTSourceCodeLocationToleratesMissingFileURL(t *testing.T) {
	objects := []interface{}{"$null", map[string]interface{}{"NS.relative": plist.UID(0)}}

	require.NotPanics(t, func() {
		NewXCTSourceCodeLocation(map[string]interface{}{"file-url": plist.UID(0), "line-number": uint64(3)}, objects)
	})
	require.NotPanics(t, func() {
		NewXCTSourceCodeLocation(map[string]interface{}{"file-url": plist.UID(1)}, objects)
	})
}
