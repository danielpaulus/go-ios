package tunnel

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/ed25519"
	"howett.net/plist"
)

func TestPairRecordManager(t *testing.T) {
	tmp := t.TempDir()

	pm, err := NewPairRecordManager(tmp)

	require.NoError(t, err)

	t.Run("self identity is created", func(t *testing.T) {
		_, err := os.Stat(path.Join(tmp, "selfIdentity.plist"))
		assert.NoError(t, err)
	})
	t.Run("read key equals the stored one", func(t *testing.T) {
		siPath := path.Join(tmp, "selfIdentity.plist")
		b, err := os.ReadFile(siPath)
		require.NoError(t, err)

		var si selfIdentity
		_, err = plist.Unmarshal(b, &si)
		require.NoError(t, err)

		private := ed25519.NewKeyFromSeed(si.PrivateKey)

		assert.True(t, private.Equal(pm.selfId.privateKey()))
	})
}

func TestGetOrCreateSelfIdentityRejectsDanglingSymlink(t *testing.T) {
	tmp := t.TempDir()
	target := path.Join(tmp, "target")
	link := path.Join(tmp, "selfIdentity.plist")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	_, err := getOrCreateSelfIdentity(link)
	require.ErrorContains(t, err, "refusing symbolic link")
	_, err = os.Stat(target)
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestCreateSelfIdentityDoesNotReplaceExistingFile(t *testing.T) {
	identityPath := path.Join(t.TempDir(), "selfIdentity.plist")
	require.NoError(t, os.WriteFile(identityPath, []byte("sentinel"), 0o600))

	_, err := createSelfIdentity(identityPath)
	require.ErrorIs(t, err, os.ErrExist)
	content, err := os.ReadFile(identityPath)
	require.NoError(t, err)
	require.Equal(t, []byte("sentinel"), content)
}
