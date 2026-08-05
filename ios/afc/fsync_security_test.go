package afc

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestListFiltersTraversalEntries covers HIGH-16: a malicious device could
// return directory entries like "../evil" from a readDir reply. These must be
// filtered out by List so they can never be joined onto a host path during a
// Pull. Legitimate names ("foo.txt") must survive unchanged.
func TestListFiltersTraversalEntries(t *testing.T) {
	// Build a readDir payload of null-terminated entry names, including the
	// traversal names that must be dropped and a legitimate name that stays.
	var payload []byte
	for _, name := range []string{".", "..", "../evil", "a/../../escape", "/etc/passwd", "foo.txt"} {
		payload = append(payload, []byte(name)...)
		payload = append(payload, 0)
	}
	raw := encodeAfcPacket(readDir, nil, payload)

	c := &Client{connection: &rwc{r: bytes.NewReader(raw)}}
	list, err := c.List("/some/dir")
	assert.NoError(t, err)

	assert.Equal(t, []string{"foo.txt"}, list)
	for _, entry := range list {
		assert.False(t, unsafeEntryName(entry), "traversal entry survived List: %q", entry)
	}
}

// TestUnsafeEntryName verifies the defense-in-depth name check: legitimate
// single filenames are accepted, while absolute names and names containing a
// ".." path element are rejected before ever being joined onto a host path.
func TestUnsafeEntryName(t *testing.T) {
	safe := []string{"foo.txt", "Documents", "a.b.c", "file with spaces", "..dotfile", "trailingdots.."}
	for _, name := range safe {
		assert.False(t, unsafeEntryName(name), "expected %q to be safe", name)
	}

	unsafe := []string{"", "..", "../evil", "a/../../escape", "/etc/passwd", "foo/../bar"}
	for _, name := range unsafe {
		assert.True(t, unsafeEntryName(name), "expected %q to be unsafe", name)
	}
}

// TestContainedInRejectsEscape verifies the containment check used at the Pull
// join site: a destination that resolves outside the root (via a ".." entry)
// is rejected, while a normal descendant is accepted. This is the host-side
// containment logic that makes the escape deterministic without a real device.
func TestContainedInRejectsEscape(t *testing.T) {
	root := filepath.Clean("/tmp/pull-dest")

	// Legitimate descendant: accepted.
	assert.True(t, containedIn(root, filepath.Join(root, "foo.txt")))
	assert.True(t, containedIn(root, filepath.Join(root, "sub", "bar.txt")))
	// The root itself is contained.
	assert.True(t, containedIn(root, root))

	// Escaping candidates: rejected. filepath.Join collapses the "..".
	escape := filepath.Join(root, "../evil")
	assert.False(t, containedIn(root, escape))
	assert.False(t, containedIn(root, filepath.Join(root, "a/../../escape")))
	// A sibling directory sharing a name prefix must not be considered inside.
	assert.False(t, containedIn(root, root+"-sibling"))
}
