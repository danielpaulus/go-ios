package afc

import (
	"bytes"
	"encoding/binary"
	"os"
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

	unsafe := []string{"", ".", "..", "../evil", "a/../../escape", "/etc/passwd", "foo/../bar"}
	for _, name := range unsafe {
		assert.True(t, unsafeEntryName(name), "expected %q to be unsafe", name)
	}
}

func TestPullSingleFileDoesNotFollowLeafSymlink(t *testing.T) {
	base := t.TempDir()
	victim := filepath.Join(base, "victim")
	destination := filepath.Join(base, "destination")
	if err := os.WriteFile(victim, []byte("unchanged"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(victim, destination); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	client := &Client{connection: &rwc{r: bytes.NewReader(filePullResponses("pulled"))}}
	if err := client.PullSingleFile("/remote", destination); err != nil {
		t.Fatalf("PullSingleFile: %v", err)
	}

	if contents, err := os.ReadFile(victim); err != nil || string(contents) != "unchanged" {
		t.Fatalf("symlink target changed: contents=%q err=%v", contents, err)
	}
	if contents, err := os.ReadFile(destination); err != nil || string(contents) != "pulled" {
		t.Fatalf("pulled destination = %q, %v", contents, err)
	}
}

func TestPullDoesNotFollowIntermediateSymlink(t *testing.T) {
	base := t.TempDir()
	destination := filepath.Join(base, "destination")
	outside := filepath.Join(base, "outside")
	if err := os.MkdirAll(destination, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(outside, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(outside, filepath.Join(destination, "sub")); err != nil {
		t.Skipf("symlinks are unavailable: %v", err)
	}

	var responses []byte
	responses = append(responses, fileInfoResponse(S_IFDIR)...)
	responses = append(responses, encodeAfcPacket(readDir, nil, []byte("sub\x00\x00"))...)
	responses = append(responses, fileInfoResponse(S_IFDIR)...)
	responses = append(responses, encodeAfcPacket(readDir, nil, []byte("file\x00\x00"))...)
	responses = append(responses, fileInfoResponse(S_IFMT)...)
	responses = append(responses, openedFileResponse()...)
	responses = append(responses, encodeAfcPacket(fileRead, nil, []byte("pulled"))...)
	responses = append(responses, encodeAfcPacket(fileRead, nil, nil)...)
	responses = append(responses, successResponse()...)

	client := &Client{connection: &rwc{r: bytes.NewReader(responses)}}
	if err := client.Pull("/remote", destination); err != nil {
		t.Fatalf("Pull: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outside, "file")); !os.IsNotExist(err) {
		t.Fatalf("pull wrote through intermediate symlink: %v", err)
	}
	if contents, err := os.ReadFile(filepath.Join(destination, "sub", "file")); err != nil || string(contents) != "pulled" {
		t.Fatalf("pulled descendant = %q, %v", contents, err)
	}
}

func filePullResponses(contents string) []byte {
	var responses []byte
	responses = append(responses, fileInfoResponse(S_IFMT)...)
	responses = append(responses, openedFileResponse()...)
	responses = append(responses, encodeAfcPacket(fileRead, nil, []byte(contents))...)
	responses = append(responses, encodeAfcPacket(fileRead, nil, nil)...)
	responses = append(responses, successResponse()...)
	return responses
}

func fileInfoResponse(fileType FileType) []byte {
	payload := []byte("st_ifmt\x00" + string(fileType) + "\x00\x00")
	return encodeAfcPacket(fileInfo, nil, payload)
}

func openedFileResponse() []byte {
	handle := make([]byte, 8)
	binary.LittleEndian.PutUint64(handle, 1)
	return encodeAfcPacket(fileOpenResult, handle, nil)
}

func successResponse() []byte {
	code := make([]byte, 8)
	binary.LittleEndian.PutUint64(code, errSuccess)
	return encodeAfcPacket(status, code, nil)
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
