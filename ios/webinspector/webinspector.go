package webinspector

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/notificationproxy"
	"github.com/google/uuid"
)

const logModule = "go-ios/webinspector"

const (
	ServiceName          = "com.apple.webinspector"
	ShimService          = "com.apple.webinspector.shim.remote"
	SafariBundleID       = "com.apple.mobilesafari"
	DisabledNotification = "com.apple.webinspectord.disabled"
)

type WIRType string

const (
	WIRTypeAutomation WIRType = "WIRTypeAutomation"
	WIRTypeJavaScript WIRType = "WIRTypeJavaScript"
	WIRTypeWeb        WIRType = "WIRTypeWeb"
	WIRTypeWebPage    WIRType = "WIRTypeWebPage"
)

const (
	AutomationAvailable    = "WIRAutomationAvailabilityAvailable"
	AutomationNotAvailable = "WIRAutomationAvailabilityNotAvailable"
)

var (
	ErrWebInspectorDisabled     = errors.New("web inspector is not enabled on the device; enable Settings > Safari > Advanced > Web Inspector, then reconnect the device")
	ErrRemoteAutomationDisabled = errors.New("remote automation is not enabled on the device; enable Settings > Safari > Advanced > Remote Automation, then retry")
)

type Application struct {
	ID           string `json:"id"`
	BundleID     string `json:"bundleId"`
	PID          int    `json:"pid"`
	Name         string `json:"name"`
	Availability string `json:"automationAvailability"`
	Active       bool   `json:"active"`
	Proxy        bool   `json:"proxy"`
	Ready        bool   `json:"ready"`
	Host         string `json:"host,omitempty"`
}

type Page struct {
	ID                     int     `json:"id"`
	Key                    string  `json:"key"`
	Type                   WIRType `json:"type"`
	URL                    string  `json:"url,omitempty"`
	Title                  string  `json:"title,omitempty"`
	AutomationIsPaired     bool    `json:"automationIsPaired,omitempty"`
	AutomationName         string  `json:"automationName,omitempty"`
	AutomationVersion      string  `json:"automationVersion,omitempty"`
	AutomationSessionID    string  `json:"automationSessionId,omitempty"`
	AutomationConnectionID string  `json:"automationConnectionId,omitempty"`
}

type ApplicationPage struct {
	Application Application `json:"application"`
	Page        Page        `json:"page"`
}

type Client struct {
	device ios.DeviceEntry
	conn   ios.DeviceConnectionInterface
	plist  ios.PlistCodecReadWriter

	connectionID string
	writeMu      sync.Mutex
	stateMu      sync.Mutex
	state        string
	apps         map[string]Application
	pages        map[string]map[string]Page

	resultMu sync.Mutex
	results  map[int]chan map[string]any
	events   chan map[string]any

	done     chan struct{}
	errs     chan error
	disabled chan error
}

func New(device ios.DeviceEntry) (*Client, error) {
	var (
		conn ios.DeviceConnectionInterface
		err  error
	)
	if device.SupportsRsd() {
		conn, err = ios.ConnectToShimService(device, ShimService)
	} else {
		conn, err = ios.ConnectToService(device, ServiceName)
	}
	if err != nil {
		return nil, err
	}
	return NewWithConnection(device, conn), nil
}

func NewWithConnection(device ios.DeviceEntry, conn ios.DeviceConnectionInterface) *Client {
	return &Client{
		device:       device,
		conn:         conn,
		plist:        ios.NewPlistCodecReadWriter(conn.Reader(), conn.Writer()),
		connectionID: strings.ToUpper(uuid.New().String()),
		apps:         make(map[string]Application),
		pages:        make(map[string]map[string]Page),
		results:      make(map[int]chan map[string]any),
		events:       make(chan map[string]any, 256),
		done:         make(chan struct{}),
		errs:         make(chan error, 1),
		disabled:     make(chan error, 1),
	}
}

func (c *Client) Connect(ctx context.Context) error {
	c.watchDisabledNotification()
	if err := c.sendMessage("_rpc_reportIdentifier:", nil); err != nil {
		return normalizeStartupError(err)
	}
	go c.readLoop()
	return normalizeStartupError(c.GetConnectedApplications(ctx))
}

func (c *Client) Close() error {
	select {
	case <-c.done:
	default:
		close(c.done)
	}
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *Client) State() string {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	return c.state
}

func (c *Client) GetConnectedApplications(ctx context.Context) error {
	if err := c.sendMessage("_rpc_getConnectedApplications:", map[string]any{}); err != nil {
		return err
	}
	return c.waitFor(ctx, func() bool {
		c.stateMu.Lock()
		defer c.stateMu.Unlock()
		return len(c.apps) > 0 || c.state != ""
	})
}

func (c *Client) OpenApp(ctx context.Context, bundleID string) (Application, error) {
	if err := c.sendMessage("_rpc_requestApplicationLaunch:", map[string]any{
		"WIRApplicationBundleIdentifierKey": bundleID,
	}); err != nil {
		return Application{}, err
	}
	var app Application
	err := c.waitFor(ctx, func() bool {
		c.stateMu.Lock()
		defer c.stateMu.Unlock()
		for _, candidate := range c.apps {
			if candidate.BundleID == bundleID {
				app = candidate
				return true
			}
		}
		return false
	})
	return app, err
}

func (c *Client) ApplicationByBundleID(bundleID string) (Application, bool) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	for _, app := range c.apps {
		if app.BundleID == bundleID {
			return app, true
		}
	}
	return Application{}, false
}

func (c *Client) ListPages(ctx context.Context, wait time.Duration) ([]ApplicationPage, error) {
	if err := c.GetConnectedApplications(ctx); err != nil {
		return nil, err
	}
	if wait > 0 {
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
	c.stateMu.Lock()
	defer c.stateMu.Unlock()

	var out []ApplicationPage
	for appID, appPages := range c.pages {
		app, ok := c.apps[appID]
		if !ok {
			continue
		}
		for _, page := range appPages {
			out = append(out, ApplicationPage{Application: app, Page: page})
		}
	}
	return out, nil
}

func (c *Client) FindPage(pageKey string) (Application, Page, bool) {
	c.stateMu.Lock()
	defer c.stateMu.Unlock()
	for appID, appPages := range c.pages {
		page, ok := appPages[pageKey]
		if !ok {
			continue
		}
		app, ok := c.apps[appID]
		return app, page, ok
	}
	return Application{}, Page{}, false
}

func (c *Client) SetupInspectorSocket(sessionID string, app Application, page Page, pause bool) error {
	message := map[string]any{
		"WIRApplicationIdentifierKey":         app.ID,
		"WIRPageIdentifierKey":                page.ID,
		"WIRSenderKey":                        sessionID,
		"WIRMessageDataTypeChunkSupportedKey": 0,
	}
	if !pause {
		message["WIRAutomaticallyPause"] = false
	}
	return c.sendMessage("_rpc_forwardSocketSetup:", message)
}

func (c *Client) RequestAutomationSession(sessionID string, app Application) error {
	return c.sendMessage("_rpc_forwardAutomationSessionRequest:", map[string]any{
		"WIRApplicationIdentifierKey": app.ID,
		"WIRSessionCapabilitiesKey": map[string]any{
			"org.webkit.webdriver.webrtc.allow-insecure-media-capture":     true,
			"org.webkit.webdriver.webrtc.suppress-ice-candidate-filtering": false,
		},
		"WIRSessionIdentifierKey": sessionID,
	})
}

func (c *Client) SendSocketData(sessionID string, app Application, page Page, data map[string]any) error {
	encoded, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.sendMessage("_rpc_forwardSocketData:", map[string]any{
		"WIRApplicationIdentifierKey": app.ID,
		"WIRPageIdentifierKey":        page.ID,
		"WIRSessionIdentifierKey":     sessionID,
		"WIRSenderKey":                sessionID,
		"WIRSocketDataKey":            encoded,
	})
}

func (c *Client) SendCommand(ctx context.Context, sessionID string, app Application, page Page, id int, method string, params map[string]any) (map[string]any, error) {
	result := make(chan map[string]any, 1)
	c.resultMu.Lock()
	c.results[id] = result
	c.resultMu.Unlock()
	defer func() {
		c.resultMu.Lock()
		delete(c.results, id)
		c.resultMu.Unlock()
	}()

	if err := c.SendSocketData(sessionID, app, page, map[string]any{
		"method": method,
		"params": params,
		"id":     id,
	}); err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case response := <-result:
		if errValue, ok := response["error"]; ok {
			return response, fmt.Errorf("webinspector command error: %v", errValue)
		}
		return response, nil
	}
}

func (c *Client) AutomationSession(ctx context.Context, app Application) (*AutomationSession, error) {
	if c.State() == AutomationNotAvailable {
		return nil, ErrRemoteAutomationDisabled
	}
	sessionID := strings.ToUpper(uuid.New().String())
	if err := c.RequestAutomationSession(sessionID, app); err != nil {
		return nil, err
	}
	if err := c.forwardGetListing(app.ID); err != nil {
		return nil, err
	}
	page, err := c.waitForAutomationPage(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if err := c.SetupInspectorSocket(sessionID, app, page, true); err != nil {
		return nil, err
	}
	if err := c.forwardGetListing(app.ID); err != nil {
		return nil, err
	}
	if err := c.waitForAutomationConnection(ctx, sessionID); err != nil {
		return nil, err
	}
	return &AutomationSession{
		client:    c,
		app:       app,
		page:      page,
		sessionID: sessionID,
		nextID:    1,
	}, nil
}

func (c *Client) Evaluate(ctx context.Context, app Application, page Page, expression string) (any, error) {
	sessionID := strings.ToUpper(uuid.New().String())
	if err := c.SetupInspectorSocket(sessionID, app, page, false); err != nil {
		return nil, err
	}

	var targetID string
	for targetID == "" {
		event, err := c.NextEvent(ctx)
		if err != nil {
			return nil, err
		}
		params, _ := event["params"].(map[string]any)
		targetInfo, _ := params["targetInfo"].(map[string]any)
		targetID, _ = targetInfo["targetId"].(string)
	}

	inner := map[string]any{
		"id":     1,
		"method": "Runtime.evaluate",
		"params": map[string]any{
			"expression":                  expression,
			"objectGroup":                 "console",
			"includeCommandLineAPI":       true,
			"returnByValue":               true,
			"generatePreview":             true,
			"userGesture":                 true,
			"awaitPromise":                false,
			"replMode":                    true,
			"allowUnsafeEvalBlockedByCSP": false,
		},
	}
	if _, err := c.SendCommand(ctx, sessionID, app, page, int(time.Now().UnixNano()&0x7fffffff), "Target.sendMessageToTarget", map[string]any{
		"targetId": targetID,
		"message":  mustMarshalString(inner),
	}); err != nil {
		return nil, err
	}

	for {
		event, err := c.NextEvent(ctx)
		if err != nil {
			return nil, err
		}
		message, ok := decodeDispatchMessage(event)
		if !ok {
			continue
		}
		id, _ := numericInt(message["id"])
		if id != 1 {
			continue
		}
		return parseEvaluateResult(message)
	}
}

func (c *Client) NextEvent(ctx context.Context) (map[string]any, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-c.errs:
		return nil, err
	case err := <-c.disabled:
		return nil, err
	case event := <-c.events:
		return event, nil
	}
}

func (c *Client) waitForAutomationPage(ctx context.Context, sessionID string) (Page, error) {
	var page Page
	err := c.waitFor(ctx, func() bool {
		c.stateMu.Lock()
		defer c.stateMu.Unlock()
		for _, appPages := range c.pages {
			for _, candidate := range appPages {
				if candidate.Type == WIRTypeAutomation && candidate.AutomationSessionID == sessionID {
					page = candidate
					return true
				}
			}
		}
		return false
	})
	if errors.Is(err, context.DeadlineExceeded) {
		err = fmt.Errorf("timed out waiting for remote automation page; %w", ErrRemoteAutomationDisabled)
	}
	return page, err
}

func (c *Client) waitForAutomationConnection(ctx context.Context, sessionID string) error {
	err := c.waitFor(ctx, func() bool {
		c.stateMu.Lock()
		defer c.stateMu.Unlock()
		for _, appPages := range c.pages {
			for _, candidate := range appPages {
				if candidate.Type == WIRTypeAutomation && candidate.AutomationSessionID == sessionID && candidate.AutomationConnectionID != "" {
					return true
				}
			}
		}
		return false
	})
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("timed out waiting for remote automation connection; %w", ErrRemoteAutomationDisabled)
	}
	return err
}

func (c *Client) readLoop() {
	for {
		select {
		case <-c.done:
			return
		default:
		}
		var message map[string]any
		if err := c.plist.Read(&message); err != nil {
			err = normalizeStartupError(err)
			if !errors.Is(err, io.EOF) && !errors.Is(err, ErrWebInspectorDisabled) {
				golog.Error("webinspector read failed", "module", logModule, "udid", c.device.Properties.SerialNumber, "error", err)
			}
			select {
			case c.errs <- err:
			default:
			}
			return
		}
		if err := c.handleMessage(message); err != nil {
			golog.Warn("webinspector message ignored", "module", logModule, "udid", c.device.Properties.SerialNumber, "error", err, "message", fmt.Sprintf("%#v", message))
		}
	}
}

func normalizeStartupError(err error) error {
	if err == nil {
		return nil
	}
	message := strings.ToLower(err.Error())
	if errors.Is(err, io.EOF) ||
		strings.Contains(message, "failed to read message length: eof") ||
		strings.Contains(message, "only 0 bytes were written") {
		return ErrWebInspectorDisabled
	}
	return err
}

func (c *Client) handleMessage(message map[string]any) error {
	selector, _ := message["__selector"].(string)
	arg, _ := message["__argument"].(map[string]any)
	switch selector {
	case "_rpc_reportCurrentState:":
		c.stateMu.Lock()
		c.state, _ = arg["WIRAutomationAvailabilityKey"].(string)
		c.stateMu.Unlock()
	case "_rpc_reportConnectedApplicationList:":
		appDict, _ := arg["WIRApplicationDictionaryKey"].(map[string]any)
		c.stateMu.Lock()
		c.apps = make(map[string]Application)
		for key, rawApp := range appDict {
			app, err := parseApplication(rawApp)
			if err != nil {
				c.stateMu.Unlock()
				return err
			}
			c.apps[key] = app
			go func(appID string) {
				if err := c.forwardGetListing(appID); err != nil {
					golog.Warn("webinspector listing request failed", "module", logModule, "udid", c.device.Properties.SerialNumber, "appID", appID, "error", err)
				}
			}(app.ID)
		}
		c.stateMu.Unlock()
	case "_rpc_applicationConnected:", "_rpc_applicationUpdated:":
		app, err := parseApplication(arg)
		if err != nil {
			return err
		}
		c.stateMu.Lock()
		c.apps[app.ID] = app
		c.stateMu.Unlock()
		if err := c.forwardGetListing(app.ID); err != nil {
			return err
		}
	case "_rpc_applicationDisconnected:":
		appID, _ := arg["WIRApplicationIdentifierKey"].(string)
		c.stateMu.Lock()
		delete(c.apps, appID)
		delete(c.pages, appID)
		c.stateMu.Unlock()
	case "_rpc_applicationSentListing:":
		appID, _ := arg["WIRApplicationIdentifierKey"].(string)
		listing, _ := arg["WIRListingKey"].(map[string]any)
		c.stateMu.Lock()
		if c.pages[appID] == nil {
			c.pages[appID] = make(map[string]Page)
		}
		for key, rawPage := range listing {
			page, err := parsePage(key, rawPage)
			if err != nil {
				c.stateMu.Unlock()
				return err
			}
			c.pages[appID][key] = page
		}
		c.stateMu.Unlock()
	case "_rpc_applicationSentData:":
		return c.handleApplicationData(arg)
	case "_rpc_reportConnectedDriverList:":
	default:
		return fmt.Errorf("unknown selector %q", selector)
	}
	return nil
}

func (c *Client) handleApplicationData(arg map[string]any) error {
	rawData := arg["WIRMessageDataKey"]
	var jsonBytes []byte
	switch value := rawData.(type) {
	case []byte:
		jsonBytes = value
	case string:
		jsonBytes = []byte(value)
	default:
		return fmt.Errorf("unexpected WIRMessageDataKey type %T", rawData)
	}
	var payload map[string]any
	if err := json.Unmarshal(jsonBytes, &payload); err != nil {
		return err
	}
	if id, ok := numericInt(payload["id"]); ok {
		c.resultMu.Lock()
		result := c.results[id]
		c.resultMu.Unlock()
		if result != nil {
			select {
			case result <- payload:
			default:
			}
			return nil
		}
	}
	select {
	case c.events <- payload:
	default:
		golog.Warn("webinspector event queue full", "module", logModule, "udid", c.device.Properties.SerialNumber)
	}
	return nil
}

func (c *Client) forwardGetListing(appID string) error {
	return c.sendMessage("_rpc_forwardGetListing:", map[string]any{"WIRApplicationIdentifierKey": appID})
}

func (c *Client) sendMessage(selector string, args map[string]any) error {
	if args == nil {
		args = make(map[string]any)
	}
	args["WIRConnectionIdentifierKey"] = c.connectionID
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.plist.Write(map[string]any{"__selector": selector, "__argument": args})
}

func (c *Client) waitFor(ctx context.Context, predicate func() bool) error {
	ticker := time.NewTicker(25 * time.Millisecond)
	defer ticker.Stop()
	for {
		if predicate() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-c.errs:
			return err
		case err := <-c.disabled:
			return err
		case <-ticker.C:
		}
	}
}

func (c *Client) watchDisabledNotification() {
	go func() {
		conn, err := notificationproxy.New(c.device)
		if err != nil {
			return
		}
		defer conn.Close()
		err = conn.Observe(DisabledNotification, 10*time.Second)
		if err == nil {
			select {
			case c.disabled <- ErrWebInspectorDisabled:
			default:
			}
		}
	}()
}

func parseApplication(raw any) (Application, error) {
	appDict, ok := raw.(map[string]any)
	if !ok {
		return Application{}, fmt.Errorf("application is %T", raw)
	}
	id, _ := appDict["WIRApplicationIdentifierKey"].(string)
	return Application{
		ID:           id,
		BundleID:     stringValue(appDict["WIRApplicationBundleIdentifierKey"]),
		PID:          pidFromApplicationID(id),
		Name:         stringValue(appDict["WIRApplicationNameKey"]),
		Availability: stringValue(appDict["WIRAutomationAvailabilityKey"]),
		Active:       boolValue(appDict["WIRIsApplicationActiveKey"]),
		Proxy:        boolValue(appDict["WIRIsApplicationProxyKey"]),
		Ready:        boolValue(appDict["WIRIsApplicationReadyKey"]),
		Host:         stringValue(appDict["WIRHostApplicationIdentifierKey"]),
	}, nil
}

func parsePage(key string, raw any) (Page, error) {
	pageDict, ok := raw.(map[string]any)
	if !ok {
		return Page{}, fmt.Errorf("page is %T", raw)
	}
	page := Page{
		ID:   intValue(pageDict["WIRPageIdentifierKey"]),
		Key:  key,
		Type: WIRType(stringValue(pageDict["WIRTypeKey"])),
	}
	if page.Type == WIRTypeWeb || page.Type == WIRTypeWebPage {
		page.Title = stringValue(pageDict["WIRTitleKey"])
		page.URL = stringValue(pageDict["WIRURLKey"])
	}
	if page.Type == WIRTypeAutomation {
		page.AutomationIsPaired = boolValue(pageDict["WIRAutomationTargetIsPairedKey"])
		page.AutomationName = stringValue(pageDict["WIRAutomationTargetNameKey"])
		page.AutomationVersion = stringValue(pageDict["WIRAutomationTargetVersionKey"])
		page.AutomationSessionID = stringValue(pageDict["WIRSessionIdentifierKey"])
		page.AutomationConnectionID = stringValue(pageDict["WIRConnectionIdentifierKey"])
	}
	return page, nil
}

func decodeDispatchMessage(event map[string]any) (map[string]any, bool) {
	if event["method"] != "Target.dispatchMessageFromTarget" {
		return nil, false
	}
	params, ok := event["params"].(map[string]any)
	if !ok {
		return nil, false
	}
	message, _ := params["message"].(string)
	if message == "" {
		return nil, false
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(message), &decoded); err != nil {
		return nil, false
	}
	return decoded, true
}

func parseEvaluateResult(message map[string]any) (any, error) {
	if errValue, ok := message["error"]; ok {
		return nil, fmt.Errorf("webinspector evaluate error: %v", errValue)
	}
	result, _ := message["result"].(map[string]any)
	remote, _ := result["result"].(map[string]any)
	if subtype, _ := remote["subtype"].(string); subtype == "error" {
		return nil, fmt.Errorf("webinspector evaluate error: %v", remote["description"])
	}
	if value, ok := remote["value"]; ok {
		return value, nil
	}
	if remote["type"] == "object" {
		if preview, ok := remote["preview"].(map[string]any); ok {
			return formatObjectPreview(stringValue(remote["className"]), preview), nil
		}
	}
	if description, ok := remote["description"].(string); ok {
		return description, nil
	}
	return nil, nil
}

func formatObjectPreview(className string, preview map[string]any) string {
	if className == "" {
		className = "Object"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "[object %s]\n{", className)
	properties, _ := preview["properties"].([]any)
	for _, rawProperty := range properties {
		property, _ := rawProperty.(map[string]any)
		name := stringValue(property["name"])
		if name == "" {
			continue
		}
		value := stringValue(property["value"])
		if value == "" {
			value = "NOT_SUPPORTED_FOR_PREVIEW"
		}
		propertyType := stringValue(property["type"])
		fmt.Fprintf(&b, "\n\t%s: %s", name, value)
		if propertyType != "" {
			fmt.Fprintf(&b, ", // %s", propertyType)
		}
	}
	if boolValue(preview["overflow"]) {
		b.WriteString("\n\t// ...")
	}
	b.WriteString("\n}")
	return b.String()
}

func mustMarshalString(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func pidFromApplicationID(id string) int {
	_, suffix, ok := strings.Cut(id, ":")
	if !ok {
		return 0
	}
	pid, _ := strconv.Atoi(suffix)
	return pid
}

func stringValue(value any) string {
	s, _ := value.(string)
	return s
}

func boolValue(value any) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	if i, ok := numericInt(value); ok {
		return i != 0
	}
	return false
}

func intValue(value any) int {
	i, _ := numericInt(value)
	return i
}

func numericInt(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case int64:
		return int(v), true
	case uint64:
		return int(v), true
	case uint:
		return int(v), true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}
