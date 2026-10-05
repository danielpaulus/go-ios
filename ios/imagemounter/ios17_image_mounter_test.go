package imagemounter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeImageMounter struct {
	images    [][]byte
	listErr   error
	mountErr  error
	mounted   []string
	unmounted bool
}

func (f *fakeImageMounter) ListImages() ([][]byte, error) { return f.images, f.listErr }
func (f *fakeImageMounter) MountImage(imagePath string) error {
	if f.mountErr != nil {
		return f.mountErr
	}
	f.mounted = append(f.mounted, imagePath)
	return nil
}
func (f *fakeImageMounter) UnmountImage() error {
	f.unmounted = true
	return nil
}
func (f *fakeImageMounter) IsImageMounted(string) (bool, error) { return len(f.images) > 0, nil }
func (f *fakeImageMounter) Close() error                        { return nil }

func newFakeIOS17Mounter(personalized, cryptex *fakeImageMounter) *ios17ImageMounter {
	return &ios17ImageMounter{mounters: []namedImageMounter{
		{name: "personalized", ImageMounter: personalized},
		{name: "cryptex", ImageMounter: cryptex},
	}}
}

func TestIOS17MounterPrefersPersonalized(t *testing.T) {
	personalized, cryptex := &fakeImageMounter{}, &fakeImageMounter{}
	require.NoError(t, newFakeIOS17Mounter(personalized, cryptex).MountImage("img"))
	assert.Equal(t, []string{"img"}, personalized.mounted)
	assert.Empty(t, cryptex.mounted)
}

func TestIOS17MounterFallsBackToCryptex(t *testing.T) {
	personalized, cryptex := &fakeImageMounter{mountErr: errors.New("failed")}, &fakeImageMounter{}
	require.NoError(t, newFakeIOS17Mounter(personalized, cryptex).MountImage("img"))
	assert.Equal(t, []string{"img"}, cryptex.mounted)
}

func TestIOS17MounterFailsIfAllMountersFail(t *testing.T) {
	personalized := &fakeImageMounter{mountErr: errors.New("personalized failed")}
	cryptex := &fakeImageMounter{mountErr: errors.New("cryptex failed")}
	err := newFakeIOS17Mounter(personalized, cryptex).MountImage("img")
	assert.ErrorContains(t, err, "personalized failed")
	assert.ErrorContains(t, err, "cryptex failed")
}

func TestIOS17MounterUnmountsOnlyMountedImages(t *testing.T) {
	personalized, cryptex := &fakeImageMounter{}, &fakeImageMounter{images: [][]byte{[]byte(ddiCryptexIdentifier)}}
	require.NoError(t, newFakeIOS17Mounter(personalized, cryptex).UnmountImage())
	assert.False(t, personalized.unmounted)
	assert.True(t, cryptex.unmounted)
}

func TestIOS17MounterUnmountsAllMountedImages(t *testing.T) {
	personalized := &fakeImageMounter{images: [][]byte{{1}}}
	cryptex := &fakeImageMounter{images: [][]byte{[]byte(ddiCryptexIdentifier)}}
	require.NoError(t, newFakeIOS17Mounter(personalized, cryptex).UnmountImage())
	assert.True(t, personalized.unmounted)
	assert.True(t, cryptex.unmounted)
}

func TestIOS17MounterUnmountFailsIfNothingIsMounted(t *testing.T) {
	assert.Error(t, newFakeIOS17Mounter(&fakeImageMounter{}, &fakeImageMounter{}).UnmountImage())
}

func TestIOS17MounterListsImagesOfAllMounters(t *testing.T) {
	personalized := &fakeImageMounter{images: [][]byte{{1}}}
	cryptex := &fakeImageMounter{images: [][]byte{{2}}}
	images, err := newFakeIOS17Mounter(personalized, cryptex).ListImages()
	require.NoError(t, err)
	assert.Equal(t, [][]byte{{1}, {2}}, images)

	personalized.listErr = errors.New("failed")
	images, err = newFakeIOS17Mounter(personalized, cryptex).ListImages()
	require.NoError(t, err)
	assert.Equal(t, [][]byte{{2}}, images)
}
