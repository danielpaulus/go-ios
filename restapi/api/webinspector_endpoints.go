package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/webinspector"
	"github.com/gin-gonic/gin"
)

// Sentinel errors for webinspector request validation.
var (
	errMissingURL    = errors.New("missing required 'url' (query param or JSON body)")
	errMissingScript = errors.New("missing required 'script' (JSON body)")
	errNoMatchingWIP = errors.New("no matching inspectable page found")
)

// webinspectorConnectTimeout bounds how long an endpoint waits for the device's
// Web Inspector service to report its pages/apps before failing. It mirrors the
// CLI default (5s) so behaviour is consistent across the CLI and the REST API.
const webinspectorConnectTimeout = 5 * time.Second

// registerWebInspectorRoutes registers the non-interactive Web Inspector
// endpoints, mirroring the `ios webinspector list|launch|eval` CLI commands.
// The interactive commands (js-shell, cdp) are intentionally not exposed over
// the REST API. Routes are registered under /device/:udid.
func registerWebInspectorRoutes(device *gin.RouterGroup) {
	group := device.Group("/webinspector")
	group.GET("/pages", WebInspectorPages)
	group.POST("/launch", WebInspectorLaunch)
	group.POST("/eval", WebInspectorEval)
}

// webInspectorLaunchRequest is the JSON body for POST /webinspector/launch.
// url may alternatively be supplied as a query param.
type webInspectorLaunchRequest struct {
	URL      string `json:"url"`
	BundleID string `json:"bundleId"`
}

// webInspectorEvalRequest is the JSON body for POST /webinspector/eval.
// page identifies the inspectable page (its key); when empty the first matching
// web/javascript page is used. bundleId optionally scopes the page selection.
type webInspectorEvalRequest struct {
	Page     string `json:"page"`
	BundleID string `json:"bundleId"`
	Script   string `json:"script"`
}

// newWebInspectorClient connects to the device's Web Inspector service and waits
// for it to report its connected applications. On the well-known "Web Inspector
// not enabled" condition it returns webinspector.ErrWebInspectorDisabled so
// callers can map it to a 4xx. The caller owns closing the returned client.
func newWebInspectorClient(ctx context.Context, device ios.DeviceEntry) (*webinspector.Client, error) {
	client, err := webinspector.New(device)
	if err != nil {
		return nil, err
	}
	if err := client.Connect(ctx); err != nil {
		client.Close()
		return nil, err
	}
	return client, nil
}

// respondWebInspectorError maps webinspector service errors to sensible HTTP
// status codes: the "not enabled on the device" conditions become 424 Failed
// Dependency (the request is well-formed but a device-side prerequisite is
// missing), everything else is a 500.
func respondWebInspectorError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, webinspector.ErrWebInspectorDisabled),
		errors.Is(err, webinspector.ErrRemoteAutomationDisabled):
		RespondError(c, http.StatusFailedDependency, err)
	default:
		RespondError(c, http.StatusInternalServerError, err)
	}
}

// WebInspectorPages lists the inspectable pages reported by the device's Web
// Inspector service (CLI: ios webinspector list).
// @Summary List inspectable Web Inspector pages
// @Param udid path string true "Device UDID"
// @Success 200 {object} []webinspector.ApplicationPage
// @Failure 424 {object} map[string]string "Web Inspector not enabled on the device"
// @Failure 500 {object} map[string]string
// @Router /device/{udid}/webinspector/pages [get]
func WebInspectorPages(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)
	ctx, cancel := context.WithTimeout(c.Request.Context(), webinspectorConnectTimeout)
	defer cancel()

	client, err := newWebInspectorClient(ctx, device)
	if err != nil {
		respondWebInspectorError(c, err)
		return
	}
	defer client.Close()

	pages, err := client.ListPages(ctx, 500*time.Millisecond)
	if err != nil {
		respondWebInspectorError(c, err)
		return
	}
	if pages == nil {
		pages = []webinspector.ApplicationPage{}
	}
	golog.Info("listed webinspector pages", "module", logModule, "udid", device.Properties.SerialNumber, "count", len(pages))
	c.JSON(http.StatusOK, pages)
}

// WebInspectorLaunch opens a URL in a new inspectable page via a remote
// automation session (CLI: ios webinspector launch <url>). bundleId defaults to
// Safari.
// @Summary Open a URL in a new inspectable page
// @Param udid path string true "Device UDID"
// @Param url query string false "URL to open (or in JSON body)"
// @Param body body webInspectorLaunchRequest false "launch request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 424 {object} map[string]string "Web Inspector / Remote Automation not enabled"
// @Failure 500 {object} map[string]string
// @Router /device/{udid}/webinspector/launch [post]
func WebInspectorLaunch(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	var req webInspectorLaunchRequest
	// Body is optional (url may come from the query), so ignore bind errors and
	// fall back to the query param.
	_ = c.ShouldBindJSON(&req)
	url := strings.TrimSpace(req.URL)
	if url == "" {
		url = strings.TrimSpace(c.Query("url"))
	}
	if url == "" {
		RespondError(c, http.StatusBadRequest, errMissingURL)
		return
	}
	bundleID := strings.TrimSpace(req.BundleID)
	if bundleID == "" {
		bundleID = strings.TrimSpace(c.Query("bundleId"))
	}
	if bundleID == "" {
		bundleID = webinspector.SafariBundleID
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), webinspectorConnectTimeout)
	defer cancel()

	client, err := newWebInspectorClient(ctx, device)
	if err != nil {
		respondWebInspectorError(c, err)
		return
	}
	defer client.Close()

	app, err := client.OpenApp(ctx, bundleID)
	if err != nil {
		respondWebInspectorError(c, err)
		return
	}
	session, err := client.AutomationSession(ctx, app)
	if err != nil {
		respondWebInspectorError(c, err)
		return
	}
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer stopCancel()
		_ = session.Stop(stopCtx)
	}()
	if err := session.Start(ctx); err != nil {
		respondWebInspectorError(c, err)
		return
	}
	if err := session.Navigate(ctx, url); err != nil {
		respondWebInspectorError(c, err)
		return
	}
	currentURL, _ := session.CurrentURL(ctx)
	title, _ := session.Title(ctx)
	golog.Info("launched webinspector url", "module", logModule, "udid", device.Properties.SerialNumber, "bundleId", bundleID, "url", currentURL)
	c.JSON(http.StatusOK, gin.H{"bundleId": bundleID, "url": currentURL, "title": title})
}

// WebInspectorEval evaluates JavaScript in an inspectable page and returns the
// result (CLI: ios webinspector eval). page identifies the target page by key;
// when omitted the first matching web/javascript page (optionally scoped by
// bundleId) is used.
// @Summary Evaluate JavaScript in an inspectable page
// @Param udid path string true "Device UDID"
// @Param body body webInspectorEvalRequest true "eval request"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string "no matching page"
// @Failure 424 {object} map[string]string "Web Inspector not enabled on the device"
// @Failure 500 {object} map[string]string
// @Router /device/{udid}/webinspector/eval [post]
func WebInspectorEval(c *gin.Context) {
	device := c.MustGet(IOS_KEY).(ios.DeviceEntry)

	var req webInspectorEvalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(req.Script) == "" {
		RespondError(c, http.StatusBadRequest, errMissingScript)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), webinspectorConnectTimeout)
	defer cancel()

	client, err := newWebInspectorClient(ctx, device)
	if err != nil {
		respondWebInspectorError(c, err)
		return
	}
	defer client.Close()

	app, page, err := resolveWebInspectorPage(ctx, client, req.Page, req.BundleID)
	if err != nil {
		if errors.Is(err, errNoMatchingWIP) {
			RespondError(c, http.StatusNotFound, err)
			return
		}
		respondWebInspectorError(c, err)
		return
	}

	result, err := client.Evaluate(ctx, app, page, req.Script)
	if err != nil {
		respondWebInspectorError(c, err)
		return
	}
	golog.Info("evaluated webinspector script", "module", logModule, "udid", device.Properties.SerialNumber, "page", page.Key)
	c.JSON(http.StatusOK, gin.H{"page": page.Key, "result": result})
}

// resolveWebInspectorPage selects the inspectable page to operate on, mirroring
// the CLI's selection logic: an exact page key is preferred; otherwise the first
// web/javascript page (optionally scoped by bundleId) is returned. Returns
// errNoMatchingWIP when nothing matches.
func resolveWebInspectorPage(ctx context.Context, client *webinspector.Client, pageID string, bundleID string) (webinspector.Application, webinspector.Page, error) {
	if pageID != "" {
		if app, page, ok := client.FindPage(pageID); ok {
			return app, page, nil
		}
	}
	pages, err := client.ListPages(ctx, 500*time.Millisecond)
	if err != nil {
		return webinspector.Application{}, webinspector.Page{}, err
	}
	for _, candidate := range pages {
		if candidate.Page.Type != webinspector.WIRTypeWeb &&
			candidate.Page.Type != webinspector.WIRTypeWebPage &&
			candidate.Page.Type != webinspector.WIRTypeJavaScript {
			continue
		}
		if bundleID != "" && candidate.Application.BundleID != bundleID {
			continue
		}
		if pageID != "" && candidate.Page.Key != pageID {
			continue
		}
		return candidate.Application, candidate.Page, nil
	}
	return webinspector.Application{}, webinspector.Page{}, errNoMatchingWIP
}
