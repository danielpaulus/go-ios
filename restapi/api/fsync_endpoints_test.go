package api

import (
	"net/http"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// TestSafeDevicePath exercises the traversal-safety helper directly: it must
// reject any path containing a ".." element (which could escape the AFC root)
// while accepting legitimate absolute and relative device paths.
func TestSafeDevicePath(t *testing.T) {
	cases := []struct {
		in     string
		want   string
		wantOk bool
	}{
		{"", ".", true},
		{".", ".", true},
		{"Documents", "Documents", true},
		{"Documents/logs", "Documents/logs", true},
		{"/DCIM/100APPLE", "/DCIM/100APPLE", true},
		{"a/b/../c", "", false},
		{"../escape", "", false},
		{"..", "", false},
		{"foo/..", "", false},
		{"/../etc", "", false},
		{"a/../../b", "", false},
	}
	for _, tc := range cases {
		got, ok := safeDevicePath(tc.in)
		if ok != tc.wantOk {
			t.Fatalf("safeDevicePath(%q) ok=%v, want %v", tc.in, ok, tc.wantOk)
		}
		if ok && got != tc.want {
			t.Fatalf("safeDevicePath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestFsyncValidationRejections checks the request-validation branches of the
// fsync handlers, which run before any device I/O. Missing required paths and
// traversal attempts must be rejected with a 400 error envelope, without ever
// opening an AFC connection.
func TestFsyncValidationRejections(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		target  string
		handler gin.HandlerFunc
		want    int
	}{
		{"ls with traversal path", "GET", "/fsync/ls?path=../../etc", FsyncLs, http.StatusBadRequest},
		{"tree with traversal path", "GET", "/fsync/tree?path=a/../../b", FsyncTree, http.StatusBadRequest},
		{"pull without path", "GET", "/fsync/pull", FsyncPull, http.StatusBadRequest},
		{"pull with traversal path", "GET", "/fsync/pull?path=../secret", FsyncPull, http.StatusBadRequest},
		{"push without path", "POST", "/fsync/push", FsyncPush, http.StatusBadRequest},
		{"push with traversal path", "POST", "/fsync/push?path=../x", FsyncPush, http.StatusBadRequest},
		{"rm without path", "DELETE", "/fsync/rm", FsyncRm, http.StatusBadRequest},
		{"rm with traversal path", "DELETE", "/fsync/rm?path=..", FsyncRm, http.StatusBadRequest},
		{"mkdir without path", "POST", "/fsync/mkdir", FsyncMkdir, http.StatusBadRequest},
		{"mkdir with traversal path", "POST", "/fsync/mkdir?path=foo/../..", FsyncMkdir, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, c := newHandlerCtx(tc.method, tc.target, "")
			tc.handler(c)
			if w.Code != tc.want {
				t.Fatalf("%s: got %d, want %d (body=%s)", tc.name, w.Code, tc.want, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), `"error"`) {
				t.Fatalf("%s: expected error envelope, got %q", tc.name, w.Body.String())
			}
		})
	}
}
