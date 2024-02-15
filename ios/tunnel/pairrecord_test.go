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
