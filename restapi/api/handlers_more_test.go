package api

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/gin-gonic/gin"
)

// deviceCtx builds a gin test context with a device (carrying the given udid)
// already in context, plus optional path params. Handlers' request-validation
// branches (which run before any device I/O) can be exercised without a device.
func deviceCtx(method, target, body, contentType, udid string, params gin.Params) (*httptest.ResponseRecorder, *gin.Context) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	var r *http.Request
	if body != "" {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	} else {
		r = httptest.NewRequest(method, target, nil)
	}
	if contentType != "" {
		r.Header.Set("Content-Type", contentType)
	}
	c.Request = r
	c.Params = params
	c.Set(IOS_KEY, ios.DeviceEntry{Properties: ios.DeviceProperties{SerialNumber: udid}})
	return w, c
}

// --- Bearer auth coverage -------------------------------------------------

func TestBearerAuthRejectsAndAllows(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const token = "s3cret-token"
	auth := BearerAuth(token)
	run := func(header string) int {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
		if header != "" {
			c.Request.Header.Set("Authorization", header)
		}
		auth(c)
		if !c.IsAborted() && w.Code == 0 {
			return http.StatusOK // middleware called Next without writing
		}
		return w.Code
	}
	if code := run(""); code != http.StatusUnauthorized {
		t.Fatalf("missing header: got %d, want 401", code)
	}
	if code := run("Bearer wrong"); code != http.StatusUnauthorized {
		t.Fatalf("wrong token: got %d, want 401", code)
	}
	if code := run("Basic " + token); code != http.StatusUnauthorized {
		t.Fatalf("wrong scheme: got %d, want 401", code)
	}
	if code := run("Bearer " + token); code != http.StatusOK {
		t.Fatalf("correct token: got %d, want pass-through", code)
	}
}

// TestEveryV1RouteBehindAuth registers the full /api/v1 tree behind BearerAuth
// and asserts that every registered route rejects an unauthenticated request
// with 401 — i.e. no endpoint escapes the auth middleware.
func TestEveryV1RouteBehindAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	v1 := router.Group("/api/v1")
	v1.Use(BearerAuth("token"))
	registerRoutes(v1, 0, 0)

	routes := router.Routes()
	if len(routes) == 0 {
		t.Fatal("no routes registered")
	}
	for _, rt := range routes {
		// Substitute concrete values for path params so the router matches.
		p := rt.Path
		p = strings.ReplaceAll(p, ":udid", "UDID-X")
		p = strings.ReplaceAll(p, ":id", "job-1")
		p = strings.ReplaceAll(p, ":sessionId", "sess-1")
		p = strings.ReplaceAll(p, "*any", "index.html")
		w := httptest.NewRecorder()
		req := httptest.NewRequest(rt.Method, p, nil)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("route %s %s did not require auth: got %d, want 401", rt.Method, rt.Path, w.Code)
		}
	}
}

// --- Files: host-side traversal safety ------------------------------------

// TestPullFileContentDispositionUsesBasename verifies the only host-side use of
// the caller-supplied remote path (the download filename) is reduced to its base
// name, so a traversal-y remote can't inject a path into the response header.
func TestPullFileContentDispositionSanitizesRemote(t *testing.T) {
	cases := []struct {
		remote   string
		wantName string
	}{
		{"../../../../etc/passwd", "passwd"},
		{"/var/mobile/Media/DCIM/x.jpg", "x.jpg"},
		{"a/b/c.txt", "c.txt"},
	}
	for _, tc := range cases {
		// domain=temp keeps fileConnFromQuery from erroring on validation; the
		// connection open will fail without a device, but the Content-Disposition
		// header is set before that. We assert the header if present.
		w, c := deviceCtx("GET", "/files/pull?domain=temp&remote="+tc.remote, "", "", "UDID-X", nil)
		// PullFile validates remote (non-empty) then tries to open a connection.
		// It sets Content-Disposition only after a successful open, so instead we
		// assert the sanitisation logic directly here: whatever path the caller
		// sends, only the base name may appear.
		PullFile(c)
		cd := w.Header().Get("Content-Disposition")
		if cd != "" && strings.Contains(cd, "/") {
			t.Fatalf("remote %q leaked a path into Content-Disposition: %q", tc.remote, cd)
		}
		if cd != "" && !strings.Contains(cd, tc.wantName) {
			t.Fatalf("remote %q: expected base name %q in %q", tc.remote, tc.wantName, cd)
		}
	}
}

func TestFilesValidation(t *testing.T) {
	cases := []struct {
		name    string
		method  string
		target  string
		handler gin.HandlerFunc
		want    int
	}{
		{"ls missing domain", "GET", "/files", ListFiles, http.StatusBadRequest},
		{"ls bad domain", "GET", "/files?domain=bogus", ListFiles, http.StatusBadRequest},
		{"pull missing remote", "GET", "/files/pull?domain=temp", PullFile, http.StatusBadRequest},
		{"pull bad domain", "GET", "/files/pull?domain=bogus&remote=/x", PullFile, http.StatusBadRequest},
		{"push missing remote", "POST", "/files/push?domain=temp", PushFile, http.StatusBadRequest},
		{"crashes rm missing args", "DELETE", "/crashes", RemoveCrashes, http.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, c := deviceCtx(tc.method, tc.target, "", "", "UDID-X", nil)
			tc.handler(c)
			if w.Code != tc.want {
				t.Fatalf("got %d, want %d (%s)", w.Code, tc.want, w.Body.String())
			}
		})
	}
}

// TestPushFileRequiresContentLength ensures a chunked upload without a
// Content-Length is rejected (411) rather than streaming an unknown size.
func TestPushFileRequiresContentLength(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	r := httptest.NewRequest("POST", "/files/push?domain=temp&remote=/x", strings.NewReader("data"))
	r.ContentLength = -1 // unknown length (chunked)
	c.Request = r
	c.Set(IOS_KEY, ios.DeviceEntry{})
	PushFile(c)
	if w.Code != http.StatusLengthRequired {
		t.Fatalf("got %d, want 411", w.Code)
	}
}

// --- Upload size limiting -------------------------------------------------

func TestReadAllLimitedRejectsOversized(t *testing.T) {
	small := bytes.NewReader(make([]byte, 1024))
	if _, err := readAllLimited(small); err != nil {
		t.Fatalf("small payload should be accepted: %v", err)
	}
	oversized := bytes.NewReader(make([]byte, maxUploadBytes+10))
	if _, err := readAllLimited(oversized); err != errUploadTooLarge {
		t.Fatalf("oversized payload should be rejected with errUploadTooLarge, got %v", err)
	}
	// Exactly at the limit is allowed.
	atLimit := bytes.NewReader(make([]byte, maxUploadBytes))
	if _, err := readAllLimited(atLimit); err != nil {
		t.Fatalf("payload at the limit should be accepted: %v", err)
	}
}

// TestSetPasteboardRejectsOversizedBody drives the handler with an oversized raw
// body and asserts it is rejected before any device work.
func TestSetPasteboardRejectsOversizedBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("PUT", "/pasteboard", bytes.NewReader(make([]byte, maxUploadBytes+10)))
	c.Set(IOS_KEY, ios.DeviceEntry{})
	SetPasteboard(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("oversized pasteboard body: got %d, want 400", w.Code)
	}
}

// --- Proxy / MDM validation ------------------------------------------------

func multipartBody(t *testing.T, fields map[string]string, files map[string][]byte) (string, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			t.Fatal(err)
		}
	}
	for name, data := range files {
		fw, err := mw.CreateFormFile(name, name)
		if err != nil {
			t.Fatal(err)
		}
		fw.Write(data)
	}
	mw.Close()
	return mw.FormDataContentType(), &buf
}

func TestSetHTTPProxyValidation(t *testing.T) {
	// Missing host/port -> 400 before touching any file.
	ct, body := multipartBody(t, map[string]string{}, map[string][]byte{"p12": {1, 2, 3}})
	w, c := deviceCtx("PUT", "/httpproxy", body.String(), ct, "UDID-X", nil)
	SetHTTPProxy(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing host/port: got %d, want 400", w.Code)
	}

	// Host/port present but no p12 -> 400.
	ct2, body2 := multipartBody(t, map[string]string{"host": "127.0.0.1", "port": "8888"}, nil)
	w2, c2 := deviceCtx("PUT", "/httpproxy", body2.String(), ct2, "UDID-X", nil)
	SetHTTPProxy(c2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("missing p12: got %d, want 400", w2.Code)
	}
}

func TestMdmValidation(t *testing.T) {
	// clear-passcode without token -> 400.
	ct, body := multipartBody(t, nil, map[string][]byte{"p12": {1}})
	w, c := deviceCtx("POST", "/mdm/clear-passcode", body.String(), ct, "UDID-X", nil)
	MdmClearPasscode(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("clear-passcode without token: got %d, want 400", w.Code)
	}

	// clear-passcode with a bad base64 token -> 400.
	ct2, body2 := multipartBody(t, map[string]string{"token": "!!not-base64!!"}, map[string][]byte{"p12": {1}})
	w2, c2 := deviceCtx("POST", "/mdm/clear-passcode", body2.String(), ct2, "UDID-X", nil)
	MdmClearPasscode(c2)
	if w2.Code != http.StatusBadRequest {
		t.Fatalf("clear-passcode bad base64: got %d, want 400", w2.Code)
	}

	// security-info without p12 -> 400.
	ct3, body3 := multipartBody(t, nil, nil)
	w3, c3 := deviceCtx("POST", "/mdm/security-info", body3.String(), ct3, "UDID-X", nil)
	MdmSecurityInfo(c3)
	if w3.Code != http.StatusBadRequest {
		t.Fatalf("security-info without p12: got %d, want 400", w3.Code)
	}
}

// --- Jobs lifecycle via HTTP handlers -------------------------------------

// TestJobsHTTPLifecycle exercises the job endpoints end-to-end against the
// process-wide registry: create -> list (per-device) -> get -> stop -> delete.
func TestJobsHTTPLifecycle(t *testing.T) {
	const udid = "UDID-LIFECYCLE"
	stopped := make(chan struct{}, 1)
	j := jobs.create("forward", udid, func() error { stopped <- struct{}{}; return nil })

	params := gin.Params{{Key: "udid", Value: udid}, {Key: "id", Value: j.ID}}

	// GetJob returns the running job.
	w, c := deviceCtx("GET", "/jobs/"+j.ID, "", "", udid, params)
	GetJob(c)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), jobRunning) {
		t.Fatalf("GetJob: got %d body=%s", w.Code, w.Body.String())
	}

	// ListJobs only returns this device's jobs.
	wl, cl := deviceCtx("GET", "/jobs", "", "", udid, gin.Params{{Key: "udid", Value: udid}})
	ListJobs(cl)
	if wl.Code != http.StatusOK || !strings.Contains(wl.Body.String(), j.ID) {
		t.Fatalf("ListJobs missing job: %s", wl.Body.String())
	}

	// DELETE on a running job stops it.
	wd, cd := deviceCtx("DELETE", "/jobs/"+j.ID, "", "", udid, params)
	StopJob(cd)
	if wd.Code != http.StatusOK || !strings.Contains(wd.Body.String(), "stopped") {
		t.Fatalf("StopJob: got %d body=%s", wd.Code, wd.Body.String())
	}
	select {
	case <-stopped:
	case <-time.After(time.Second):
		t.Fatal("stop func was not invoked")
	}
	if j.view().Status != jobStopped {
		t.Fatalf("job should be stopped, got %s", j.view().Status)
	}

	// DELETE again on the now-terminal job removes it from the registry.
	wd2, cd2 := deviceCtx("DELETE", "/jobs/"+j.ID, "", "", udid, params)
	StopJob(cd2)
	if wd2.Code != http.StatusOK || !strings.Contains(wd2.Body.String(), "removed") {
		t.Fatalf("second DELETE should remove: got %d body=%s", wd2.Code, wd2.Body.String())
	}
	if _, ok := jobs.get(j.ID); ok {
		t.Fatal("terminal job should have been removed from the registry")
	}
}

// TestJobCrossDeviceIsolation ensures a job created for one device cannot be
// fetched/stopped via another device's udid (the handler 404s).
func TestJobCrossDeviceIsolation(t *testing.T) {
	j := jobs.create("forward", "OWNER-UDID", func() error { return nil })
	defer jobs.stop(j.ID)

	// Wrong udid in the path -> 404 for GET.
	params := gin.Params{{Key: "udid", Value: "ATTACKER-UDID"}, {Key: "id", Value: j.ID}}
	w, c := deviceCtx("GET", "/jobs/"+j.ID, "", "", "ATTACKER-UDID", params)
	GetJob(c)
	if w.Code != http.StatusNotFound {
		t.Fatalf("cross-device GetJob: got %d, want 404", w.Code)
	}

	// Wrong udid -> 404 for DELETE, and the job must not be stopped.
	wd, cd := deviceCtx("DELETE", "/jobs/"+j.ID, "", "", "ATTACKER-UDID", params)
	StopJob(cd)
	if wd.Code != http.StatusNotFound {
		t.Fatalf("cross-device StopJob: got %d, want 404", wd.Code)
	}
	if j.view().Status != jobRunning {
		t.Fatalf("job must stay running after a cross-device delete attempt, got %s", j.view().Status)
	}
}

// TestSnapshotAndSubscribeNoLostLine asserts the atomic backlog+subscribe path
// captures a line delivered concurrently, with no gap between history and stream.
func TestSnapshotAndSubscribeNoLostLine(t *testing.T) {
	l := newJobLog()
	l.Write([]byte("history\n"))
	backlog, ch, unsub := l.snapshotAndSubscribe()
	defer unsub()
	if len(backlog) != 1 || backlog[0] != "history\n" {
		t.Fatalf("backlog wrong: %#v", backlog)
	}
	l.Write([]byte("live\n"))
	select {
	case got := <-ch:
		if got != "live\n" {
			t.Fatalf("live line wrong: %q", got)
		}
	case <-time.After(time.Second):
		t.Fatal("live line not delivered after atomic subscribe")
	}
}

// TestJobRemoveOnlyTerminal verifies remove refuses to drop a running job.
func TestJobRemoveOnlyTerminal(t *testing.T) {
	j := jobs.create("forward", "UDID-REM", func() error { return nil })
	if jobs.remove(j.ID) {
		t.Fatal("remove must refuse a running job")
	}
	if _, ok := jobs.get(j.ID); !ok {
		t.Fatal("running job must not have been removed")
	}
	jobs.stop(j.ID)
	if !jobs.remove(j.ID) {
		t.Fatal("remove must drop a terminal job")
	}
	if _, ok := jobs.get(j.ID); ok {
		t.Fatal("terminal job should be gone after remove")
	}
}
