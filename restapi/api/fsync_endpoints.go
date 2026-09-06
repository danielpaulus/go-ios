package api

import (
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/afc"
	"github.com/danielpaulus/go-ios/ios/house_arrest"
	"github.com/gin-gonic/gin"
)

// registerFsyncRoutes registers the AFC (Apple File Conduit) filesystem endpoints
// mirroring `ios fsync`. All routes live under /device/:udid.
//
// This is the AFC-based surface (works over the classic com.apple.afc lockdown
// service via ios/afc), distinct from the iOS 17+ fileservice surface exposed by
// /files. By default it operates on the media directory (afc.New). Passing
// ?bundleID=<id> scopes operations to that app's data container via
// house_arrest, mirroring `ios fsync --app=<bundleID>`.
//
// The caller-supplied ?path= is a device-side path. We reject any path with a
// ".." element via safeDevicePath so a request cannot traverse out of the AFC
// root (media dir or vended app container). On pull, entry names returned by the
// device are additionally filtered by the afc package itself.
func registerFsyncRoutes(device *gin.RouterGroup) {
	device.GET("/fsync/ls", FsyncLs)
	device.GET("/fsync/tree", FsyncTree)
	device.GET("/fsync/pull", FsyncPull)
	device.POST("/fsync/push", FsyncPush)
	device.DELETE("/fsync/rm", FsyncRm)
	device.POST("/fsync/mkdir", FsyncMkdir)
}

// afcClientFromQuery opens an AFC client for the request. If ?bundleID= is set it
// vends the app's data container via house_arrest; otherwise it connects to the
// device media directory. The caller is responsible for closing the client.
func afcClientFromQuery(c *gin.Context) (*afc.Client, error) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	bundleID := c.Query("bundleID")
	if bundleID != "" {
		return house_arrest.New(device, bundleID)
	}
	return afc.New(device)
}

// safeDevicePath validates a caller-supplied device path. It rejects paths that
// contain a ".." element, which could otherwise escape the AFC root (the media
// directory or a vended app container). Empty paths default to ".". A leading
// "/" is permitted since AFC paths are rooted at the AFC domain, not the host
// filesystem. Returns the cleaned path and whether it is safe.
func safeDevicePath(p string) (string, bool) {
	if p == "" {
		return ".", true
	}
	// Split on "/" (AFC uses POSIX separators) and reject any ".." element.
	for _, part := range strings.Split(p, "/") {
		if part == ".." {
			return "", false
		}
	}
	return path.Clean(p), true
}

// pathFromQuery reads and validates the ?path= query param, responding with a
// 400 on an unsafe path. It returns the cleaned path and ok=false when it has
// already written an error response.
func pathFromQuery(c *gin.Context, allowEmpty bool) (string, bool) {
	raw := c.Query("path")
	if raw == "" && !allowEmpty {
		RespondError(c, http.StatusBadRequest, errMissingPath)
		return "", false
	}
	p, ok := safeDevicePath(raw)
	if !ok {
		RespondError(c, http.StatusBadRequest, errUnsafePath)
		return "", false
	}
	return p, true
}

// FsyncLs lists a directory over AFC (CLI: ios fsync ls).
// @Summary List a device directory over AFC
// @Produce json
// @Param udid path string true "Device UDID"
// @Param bundleID query string false "app bundle id to scope to its container (else media dir)"
// @Param path query string false "directory path (default '.')"
// @Success 200 {object} map[string]interface{}
// @Router /device/{udid}/fsync/ls [get]
func FsyncLs(c *gin.Context) {
	p, ok := pathFromQuery(c, true)
	if !ok {
		return
	}
	client, err := afcClientFromQuery(c)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer client.Close()
	files, err := client.List(p)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"path": p, "files": files, "count": len(files)})
}

// fsyncTreeEntry is one node in the recursive tree response.
type fsyncTreeEntry struct {
	Path  string `json:"path"`
	Name  string `json:"name"`
	IsDir bool   `json:"isDir"`
	Size  int64  `json:"size"`
}

// FsyncTree walks a directory recursively over AFC (CLI: ios fsync tree).
// @Summary Recursively list a device directory over AFC
// @Produce json
// @Param udid path string true "Device UDID"
// @Param bundleID query string false "app bundle id to scope to its container (else media dir)"
// @Param path query string false "root path (default '.')"
// @Success 200 {object} map[string]interface{}
// @Router /device/{udid}/fsync/tree [get]
func FsyncTree(c *gin.Context) {
	p, ok := pathFromQuery(c, true)
	if !ok {
		return
	}
	client, err := afcClientFromQuery(c)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer client.Close()
	entries := make([]fsyncTreeEntry, 0)
	err = client.WalkDir(p, func(entryPath string, info afc.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		entries = append(entries, fsyncTreeEntry{
			Path:  entryPath,
			Name:  info.Name,
			IsDir: info.IsDir(),
			Size:  info.Size,
		})
		return nil
	})
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"path": p, "entries": entries, "count": len(entries)})
}

// FsyncPull streams a device file back as the response body (CLI: ios fsync pull).
// @Summary Download a file from the device over AFC
// @Produce application/octet-stream
// @Param udid path string true "Device UDID"
// @Param bundleID query string false "app bundle id to scope to its container (else media dir)"
// @Param path query string true "remote file path on the device"
// @Success 200 {file} binary
// @Router /device/{udid}/fsync/pull [get]
func FsyncPull(c *gin.Context) {
	p, ok := pathFromQuery(c, false)
	if !ok {
		return
	}
	client, err := afcClientFromQuery(c)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer client.Close()
	info, err := client.Stat(p)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	if info.IsDir() {
		RespondError(c, http.StatusBadRequest, errPullIsDir)
		return
	}
	src := p
	if info.IsLink() && info.LinkTarget != "" {
		src = info.LinkTarget
	}
	fd, err := client.Open(src, afc.READ_ONLY)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer fd.Close()
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename=\""+path.Base(p)+"\"")
	if _, err := io.Copy(c.Writer, fd); err != nil {
		// Headers may already be sent; only surface an error if nothing was written.
		if !c.Writer.Written() {
			RespondError(c, http.StatusInternalServerError, err)
		}
		return
	}
}

// FsyncPush uploads the request body to a device path over AFC (CLI: ios fsync
// push). Accepts either a raw request body or a multipart form with a "file"
// field. Uploads are streamed to the device and bounded by maxUploadBytes.
// @Summary Upload a file to the device over AFC
// @Accept application/octet-stream
// @Accept multipart/form-data
// @Param udid path string true "Device UDID"
// @Param bundleID query string false "app bundle id to scope to its container (else media dir)"
// @Param path query string true "destination path on the device"
// @Success 200 {object} map[string]interface{}
// @Router /device/{udid}/fsync/push [post]
func FsyncPush(c *gin.Context) {
	p, ok := pathFromQuery(c, false)
	if !ok {
		return
	}

	// Resolve the upload source: a multipart "file" field if present, else the
	// raw request body. Either way the reader is bounded by maxUploadBytes so an
	// authenticated client cannot exhaust server memory while we stream to the
	// device.
	var body io.Reader
	if f, _, err := c.Request.FormFile("file"); err == nil {
		defer f.Close()
		body = f
	} else {
		body = c.Request.Body
	}
	bounded := io.LimitReader(body, maxUploadBytes+1)

	client, err := afcClientFromQuery(c)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer client.Close()

	cw := &countingReader{r: bounded}
	if err := client.WriteToFile(cw, p); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	if cw.n > maxUploadBytes {
		RespondError(c, http.StatusRequestEntityTooLarge, errUploadTooLarge)
		return
	}
	c.JSON(http.StatusOK, gin.H{"path": p, "size": cw.n})
}

// countingReader counts bytes read so a bounded upload can be checked against
// the size cap after streaming without buffering the whole payload.
type countingReader struct {
	r io.Reader
	n int64
}

func (cr *countingReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	cr.n += int64(n)
	return n, err
}

// FsyncRm removes a file or directory over AFC (CLI: ios fsync rm). Pass
// ?recursive=true to delete a non-empty directory.
// @Summary Remove a file or directory on the device over AFC
// @Param udid path string true "Device UDID"
// @Param bundleID query string false "app bundle id to scope to its container (else media dir)"
// @Param path query string true "path to remove"
// @Param recursive query bool false "remove directory contents recursively"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/fsync/rm [delete]
func FsyncRm(c *gin.Context) {
	p, ok := pathFromQuery(c, false)
	if !ok {
		return
	}
	client, err := afcClientFromQuery(c)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer client.Close()
	if c.Query("recursive") == "true" {
		err = client.RemoveAll(p)
	} else {
		err = client.Remove(p)
	}
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "removed", "path": p})
}

// FsyncMkdir creates a directory over AFC (CLI: ios fsync mkdir).
// @Summary Create a directory on the device over AFC
// @Param udid path string true "Device UDID"
// @Param bundleID query string false "app bundle id to scope to its container (else media dir)"
// @Param path query string true "directory path to create"
// @Success 200 {object} map[string]string
// @Router /device/{udid}/fsync/mkdir [post]
func FsyncMkdir(c *gin.Context) {
	p, ok := pathFromQuery(c, false)
	if !ok {
		return
	}
	client, err := afcClientFromQuery(c)
	if err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	defer client.Close()
	if err := client.MkDir(p); err != nil {
		RespondError(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "created", "path": p})
}
