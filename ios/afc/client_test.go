package afc

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"slices"
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestAfc writes to and deletes from the device's AFC media directory, so it only
// runs against the UDIDs listed in GO_IOS_E2E_DEVICES, never against whatever
// happens to be plugged in.
func TestAfc(t *testing.T) {
	udids := strings.TrimSpace(os.Getenv("GO_IOS_E2E_DEVICES"))
	if udids == "" {
		t.Skip("GO_IOS_E2E_DEVICES not set")
	}

	for _, udid := range strings.Split(udids, ",") {
		udid = strings.TrimSpace(udid)
		if udid == "" {
			continue
		}
		device, err := ios.GetDevice(udid)
		if err != nil {
			t.Fatalf("GetDevice(%s): %v", udid, err)
		}

		t.Run(fmt.Sprintf("device %s", device.Properties.SerialNumber), func(t *testing.T) {

			client, err := New(device)
			assert.NoError(t, err)

			defer client.Close()

			t.Run("list /tmp", func(t *testing.T) {
				_, err := client.List("/")
				assert.NoError(t, err)
			})

			t.Run("list invalid folder returns error", func(t *testing.T) {
				_, err := client.List("/invalid123")
				assert.Error(t, err)
			})

			t.Run("create file", func(t *testing.T) {
				files, err := client.List(".")
				assert.NoError(t, err)
				hasFile := slices.ContainsFunc(files, func(s string) bool {
					return strings.Contains(s, "test-file")
				})
				if hasFile {
					err = client.Remove("./test-file")
					assert.NoError(t, err)
				}
				f, err := client.Open("./test-file", READ_WRITE_CREATE_TRUNC)
				assert.NoError(t, err)

				err = f.Close()
				assert.NoError(t, err)

				err = client.Remove("./test-file")
				assert.NoError(t, err)
			})

			t.Run("write to file and check size", func(t *testing.T) {
				f, err := client.Open("./test-file", READ_WRITE_CREATE_TRUNC)
				assert.NoError(t, err)

				n, err := f.Write([]byte("test"))
				assert.NoError(t, err)
				assert.Equal(t, 4, n)

				err = f.Close()
				assert.NoError(t, err)

				info, err := client.Stat("./test-file")

				assert.EqualValues(t, 4, info.Size)

				err = client.Remove("./test-file")
				assert.NoError(t, err)
			})

			t.Run("write to file and read it back", func(t *testing.T) {
				f, err := client.Open("./test-file", READ_WRITE_CREATE_TRUNC)
				assert.NoError(t, err)

				n, err := f.Write([]byte("test"))
				assert.NoError(t, err)
				assert.Equal(t, 4, n)

				err = f.Close()
				assert.NoError(t, err)

				f, err = client.Open("./test-file", READ_ONLY)
				assert.NoError(t, err)

				b := make([]byte, 8)
				n, err = f.Read(b)
				assert.NoError(t, err)
				assert.Equal(t, []byte("test"), b[:n])

				err = client.Remove("./test-file")
				assert.NoError(t, err)
			})

			t.Run("create and delete nested directory", func(t *testing.T) {
				err = client.MkDir("./some/nested/directory")
				assert.NoError(t, err)

				var info FileInfo
				info, err = client.Stat("./some")
				assert.NoError(t, err)
				assert.Equal(t, S_IFDIR, info.Type)

				info, err = client.Stat("./some/nested")
				assert.NoError(t, err)
				assert.Equal(t, S_IFDIR, info.Type)

				info, err = client.Stat("./some/nested/directory")
				assert.NoError(t, err)
				assert.Equal(t, S_IFDIR, info.Type)

				err = client.RemoveAll("./some")
				assert.NoError(t, err)

				_, err = client.Stat("./some")
				assert.Error(t, err)
			})

			t.Run("walk dir", func(t *testing.T) {
				basePath := path.Join("./", uuid.New().String())
				mustCreateDir(client, basePath)
				t.Cleanup(func() {
					assert.NoError(t, client.RemoveAll(basePath))
				})
				mustCreateDir(client, path.Join(basePath, "a-dir"))
				mustCreateDir(client, path.Join(basePath, "a-dir", "subdir"))
				mustCreateFile(client, path.Join(basePath, "a-dir", "file"))
				mustCreateDir(client, path.Join(basePath, "c-dir"))

				t.Run("visit all", func(t *testing.T) {
					var visited []string
					err = client.WalkDir(basePath, func(path string, info FileInfo, err error) error {
						visited = append(visited, path)
						return nil
					})

					assert.NoError(t, err)
					assert.Equal(t, []string{
						path.Join(basePath, "a-dir"),
						path.Join(basePath, "a-dir/file"),
						path.Join(basePath, "a-dir/subdir"),
						path.Join(basePath, "c-dir"),
					}, visited)
				})

				t.Run("skip dir", func(t *testing.T) {
					var visited []string
					err = client.WalkDir(basePath, func(p string, info FileInfo, err error) error {
						visited = append(visited, p)
						if path.Base(p) == "a-dir" {
							return fs.SkipDir
						}
						return nil
					})

					assert.NoError(t, err)
					assert.Equal(t, []string{
						path.Join(basePath, "a-dir"),
						path.Join(basePath, "c-dir"),
					}, visited)
				})

				t.Run("skip all", func(t *testing.T) {
					var visited []string
					err = client.WalkDir(basePath, func(p string, info FileInfo, err error) error {
						visited = append(visited, p)
						return fs.SkipAll
					})

					assert.NoError(t, err)
					assert.Equal(t, []string{
						path.Join(basePath, "a-dir"),
					}, visited)
				})

				t.Run("return error stops walkdir", func(t *testing.T) {
					var visited []string
					walkDirErr := errors.New("stop walkdir")
					err = client.WalkDir(basePath, func(p string, info FileInfo, err error) error {
						visited = append(visited, p)
						return walkDirErr
					})
					assert.Len(t, visited, 1)
					assert.Equal(t, walkDirErr, err)
				})

				t.Run("device info", func(t *testing.T) {
					_, err := client.DeviceInfo()
					assert.NoError(t, err)
				})
			})
		})
	}
}

func mustCreateDir(c *Client, dir string) {
	err := c.MkDir(dir)
	if err != nil {
		panic(err)
	}
}

func mustCreateFile(c *Client, path string) {
	f, err := c.Open(path, READ_WRITE_CREATE)
	if err != nil {
		panic(err)
	}
	_ = f.Close()
}
