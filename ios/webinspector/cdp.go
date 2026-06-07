package webinspector

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type CDPServer struct {
	client *Client
	host   string
	port   int
	server *http.Server
}

type CDPTarget struct {
	Description          string `json:"description"`
	ID                   string `json:"id"`
	Title                string `json:"title"`
	Type                 string `json:"type"`
	URL                  string `json:"url"`
	WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	DevtoolsFrontendURL  string `json:"devtoolsFrontendUrl"`
}

type cdpPageSession struct {
	server    *CDPServer
	ws        *websocket.Conn
	sessionID string
	app       Application
	page      Page
	targetID  string

	writeMu sync.Mutex
	sendMu  sync.Mutex

	pendingMu sync.Mutex
	pending   map[int]chan map[string]any

	frame              map[string]any
	defaultExecutionID any
	screencast         *cdpScreencast
}

type cdpScreencast struct {
	format     string
	quality    int
	maxWidth   int
	maxHeight  int
	deviceW    int
	deviceH    int
	scale      float64
	frameID    int
	lastAck    int
	cancelFunc context.CancelFunc
}

func NewCDPServer(client *Client, host string, port int) *CDPServer {
	if host == "" {
		host = "127.0.0.1"
	}
	if port == 0 {
		port = 9222
	}
	return &CDPServer{client: client, host: host, port: port}
}

func (s *CDPServer) Addr() string {
	return fmt.Sprintf("%s:%d", s.host, s.port)
}

func (s *CDPServer) Serve(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/json", s.handleTargets)
	mux.HandleFunc("/json/list", s.handleTargets)
	mux.HandleFunc("/json/version", s.handleVersion)
	mux.HandleFunc("/devtools/page/", s.handlePageWebSocket)
	s.server = &http.Server{Addr: s.Addr(), Handler: mux}

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(shutdownCtx); err != nil {
			return err
		}
		return ctx.Err()
	case err := <-errCh:
		if err == http.ErrServerClosed {
			return nil
		}
		return err
	}
}

func (s *CDPServer) handleTargets(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	pages, err := s.client.ListPages(ctx, 250*time.Millisecond)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	targets := make([]CDPTarget, 0, len(pages))
	for _, appPage := range pages {
		page := appPage.Page
		if page.Type != WIRTypeWeb && page.Type != WIRTypeWebPage {
			continue
		}
		wsURL := fmt.Sprintf("ws://%s/devtools/page/%s", s.Addr(), page.Key)
		targets = append(targets, CDPTarget{
			ID:                   page.Key,
			Title:                page.Title,
			Type:                 "page",
			URL:                  page.URL,
			WebSocketDebuggerURL: wsURL,
			DevtoolsFrontendURL:  "/devtools/inspector.html?ws=" + s.Addr() + "/devtools/page/" + page.Key,
		})
	}
	writeJSON(w, targets)
}

func (s *CDPServer) handleVersion(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, map[string]any{
		"Browser":              "Safari",
		"Protocol-Version":     "1.1",
		"User-Agent":           "go-ios",
		"WebKit-Version":       "",
		"webSocketDebuggerUrl": fmt.Sprintf("ws://%s/devtools/browser/%s", s.Addr(), s.client.connectionID),
	})
}

func (s *CDPServer) handlePageWebSocket(w http.ResponseWriter, r *http.Request) {
	pageKey := strings.TrimPrefix(r.URL.Path, "/devtools/page/")
	app, page, ok := s.client.FindPage(pageKey)
	if !ok {
		http.Error(w, "page not found", http.StatusNotFound)
		return
	}

	upgrader := websocket.Upgrader{CheckOrigin: func(*http.Request) bool { return true }}
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	sessionID := strings.ToUpper(uuid.New().String())
	if err := s.client.SetupInspectorSocket(sessionID, app, page, false); err != nil {
		_ = ws.WriteJSON(cdpError(0, err))
		return
	}
	targetID := waitForTargetID(ctx, s.client, ws)
	if targetID == "" {
		return
	}

	session := &cdpPageSession{
		server:    s,
		ws:        ws,
		sessionID: sessionID,
		app:       app,
		page:      page,
		targetID:  targetID,
		pending:   make(map[int]chan map[string]any),
	}
	errCh := make(chan error, 2)
	go func() {
		errCh <- session.forwardDeviceEvents(ctx)
	}()
	go func() {
		errCh <- session.forwardBrowserCommands(ctx)
	}()
	<-errCh
}

func (s *cdpPageSession) writeJSON(value any) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.ws.WriteJSON(value)
}

func (s *cdpPageSession) forwardDeviceEvents(ctx context.Context) error {
	for {
		event, err := s.server.client.NextEvent(ctx)
		if err != nil {
			return err
		}
		if dispatch, ok := unwrapDispatchMessage(event); ok {
			if s.resolvePending(dispatch) {
				continue
			}
			if normalized, drop := s.normalizeCDPEvent(dispatch); !drop {
				if err := s.writeJSON(normalized); err != nil {
					return err
				}
			}
			continue
		}
		if normalized, drop := s.normalizeCDPEvent(event); !drop {
			if err := s.writeJSON(normalized); err != nil {
				return err
			}
		}
	}
}

func (s *cdpPageSession) forwardBrowserCommands(ctx context.Context) error {
	for {
		var message map[string]any
		if err := s.ws.ReadJSON(&message); err != nil {
			return err
		}
		id, _ := numericInt(message["id"])

		if handled, response, err := s.handleSpecialCommand(ctx, message); handled {
			if err != nil {
				response = cdpError(id, err)
			}
			if err := s.writeJSON(response); err != nil {
				return err
			}
			continue
		}

		if handled, response, extra := localCDPResponse(message, s.targetID, s.sessionID, s.page); handled {
			if err := s.writeJSON(response); err != nil {
				return err
			}
			for _, event := range extra {
				if err := s.writeJSON(event); err != nil {
					return err
				}
			}
			continue
		}

		message = translateCDPCommand(message)
		wrapped := map[string]any{
			"method": "Target.sendMessageToTarget",
			"params": map[string]any{
				"targetId": s.targetID,
				"message":  mustJSON(message),
			},
		}
		if _, err := s.server.client.SendCommand(ctx, s.sessionID, s.app, s.page, nextWIRID(), "Target.sendMessageToTarget", wrapped["params"].(map[string]any)); err != nil {
			if writeErr := s.writeJSON(cdpError(id, err)); writeErr != nil {
				return writeErr
			}
		}
	}
}

func waitForTargetID(ctx context.Context, client *Client, ws *websocket.Conn) string {
	for {
		event, err := client.NextEvent(ctx)
		if err != nil {
			_ = ws.WriteJSON(cdpError(0, err))
			return ""
		}
		targetInfo, ok := event["params"].(map[string]any)["targetInfo"].(map[string]any)
		if !ok {
			continue
		}
		targetID, _ := targetInfo["targetId"].(string)
		if normalized, drop := normalizeCDPEvent(event); !drop {
			_ = ws.WriteJSON(normalized)
		}
		return targetID
	}
}

func unwrapDispatchMessage(event map[string]any) (map[string]any, bool) {
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

func (s *cdpPageSession) resolvePending(message map[string]any) bool {
	id, ok := numericInt(message["id"])
	if !ok {
		return false
	}
	s.pendingMu.Lock()
	ch := s.pending[id]
	s.pendingMu.Unlock()
	if ch == nil {
		return false
	}
	select {
	case ch <- message:
	default:
	}
	return true
}

func (s *cdpPageSession) sendInner(ctx context.Context, message map[string]any) error {
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	_, err := s.server.client.SendCommand(ctx, s.sessionID, s.app, s.page, nextWIRID(), "Target.sendMessageToTarget", map[string]any{
		"targetId": s.targetID,
		"message":  mustJSON(message),
	})
	return err
}

func (s *cdpPageSession) sendInnerWait(ctx context.Context, message map[string]any) (map[string]any, error) {
	id, ok := numericInt(message["id"])
	if !ok || id == 0 {
		id = nextWIRID()
		message["id"] = id
	}
	ch := make(chan map[string]any, 1)
	s.pendingMu.Lock()
	s.pending[id] = ch
	s.pendingMu.Unlock()
	defer func() {
		s.pendingMu.Lock()
		delete(s.pending, id)
		s.pendingMu.Unlock()
	}()

	if err := s.sendInner(ctx, message); err != nil {
		return nil, err
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case response := <-ch:
		if errValue, ok := response["error"]; ok {
			return response, fmt.Errorf("cdp inner command error: %v", errValue)
		}
		return response, nil
	}
}

func (s *cdpPageSession) evaluate(ctx context.Context, id int, expression string, returnByValue bool) (map[string]any, error) {
	response, err := s.sendInnerWait(ctx, map[string]any{
		"id":     id,
		"method": "Runtime.evaluate",
		"params": map[string]any{
			"expression":                  expression,
			"objectGroup":                 "console",
			"includeCommandLineAPI":       true,
			"returnByValue":               returnByValue,
			"generatePreview":             true,
			"userGesture":                 true,
			"awaitPromise":                false,
			"replMode":                    true,
			"allowUnsafeEvalBlockedByCSP": false,
		},
	})
	if err != nil {
		return nil, err
	}
	result, _ := response["result"].(map[string]any)
	remote, _ := result["result"].(map[string]any)
	return remote, nil
}

func (s *cdpPageSession) handleSpecialCommand(ctx context.Context, message map[string]any) (bool, map[string]any, error) {
	id, _ := numericInt(message["id"])
	method, _ := message["method"].(string)
	params, _ := message["params"].(map[string]any)
	switch method {
	case "DOM.getNodeForLocation":
		response, err := s.domGetNodeForLocation(ctx, id, params)
		return true, response, err
	case "DOM.getNodesForSubtreeByStyle":
		response, err := s.domGetNodesForSubtreeByStyle(ctx, id, params)
		return true, response, err
	case "DOMDebugger.getEventListeners":
		response, err := s.domDebuggerGetEventListeners(ctx, id, params)
		return true, response, err
	case "Debugger.setBlackboxPatterns":
		response, err := s.debuggerSetBlackboxPatterns(ctx, id, params)
		return true, response, err
	case "Runtime.compileScript":
		response, err := s.runtimeCompileScript(ctx, id, params)
		return true, response, err
	case "Runtime.getIsolateId":
		return true, s.runtimeGetIsolateID(id), nil
	case "Page.getNavigationHistory":
		response, err := s.pageGetNavigationHistory(ctx, id)
		return true, response, err
	case "Page.getResourceTree":
		response, err := s.pageGetResourceTree(ctx, message)
		return true, response, err
	case "Page.startScreencast":
		response, err := s.pageStartScreencast(ctx, id, params)
		return true, response, err
	case "Page.stopScreencast":
		return true, s.pageStopScreencast(id), nil
	case "Page.screencastFrameAck":
		return true, s.pageScreencastFrameAck(id, params), nil
	case "Input.emulateTouchFromMouseEvent":
		response, err := s.inputEmulateTouchFromMouseEvent(ctx, id, params)
		return true, response, err
	case "Input.dispatchKeyEvent":
		response, err := s.inputDispatchKeyEvent(ctx, id, params)
		return true, response, err
	default:
		return false, nil, nil
	}
}

func (s *cdpPageSession) domGetNodeForLocation(ctx context.Context, id int, params map[string]any) (map[string]any, error) {
	x, _ := numericInt(params["x"])
	y, _ := numericInt(params["y"])
	remote, err := s.evaluate(ctx, id, fmt.Sprintf("document.elementFromPoint(%d,%d)", x, y), false)
	if err != nil {
		return nil, err
	}
	objectID := stringValue(remote["objectId"])
	if objectID == "" {
		return cdpResult(id, map[string]any{}), nil
	}
	node, err := s.objectIDToNodeID(ctx, id, objectID)
	if err != nil {
		return nil, err
	}
	return cdpResult(id, map[string]any{"nodeId": node}), nil
}

func (s *cdpPageSession) objectIDToNodeID(ctx context.Context, id int, objectID string) (int, error) {
	response, err := s.sendInnerWait(ctx, map[string]any{
		"id":     id,
		"method": "DOM.requestNode",
		"params": map[string]any{"objectId": objectID},
	})
	if err != nil {
		return 0, err
	}
	result, _ := response["result"].(map[string]any)
	nodeID, _ := numericInt(result["nodeId"])
	return nodeID, nil
}

func (s *cdpPageSession) domGetNodesForSubtreeByStyle(ctx context.Context, id int, params map[string]any) (map[string]any, error) {
	resolve, err := s.sendInnerWait(ctx, map[string]any{
		"id":     id,
		"method": "DOM.resolveNode",
		"params": map[string]any{"nodeId": params["nodeId"]},
	})
	if err != nil {
		return nil, err
	}
	resolveResult, _ := resolve["result"].(map[string]any)
	object, _ := resolveResult["object"].(map[string]any)
	objectID := stringValue(object["objectId"])
	if objectID == "" {
		return cdpResult(id, map[string]any{"nodeIds": []int{}}), nil
	}
	call, err := s.sendInnerWait(ctx, map[string]any{
		"id":     id,
		"method": "Runtime.callFunctionOn",
		"params": map[string]any{
			"objectId":            objectID,
			"functionDeclaration": "function(styles) { const result = new Set(); var all = this.getElementsByTagName('*'); for (var elem_i = 0; elem_i < all.length; elem_i++) { for (var style_i in styles) { if (window.getComputedStyle(all[elem_i]).getPropertyValue(styles[style_i].name) === styles[style_i].value) { result.add(all[elem_i]); break; } } } return result; }",
			"arguments":           []map[string]any{{"value": params["computedStyles"]}},
		},
	})
	if err != nil {
		return nil, err
	}
	callResult, _ := call["result"].(map[string]any)
	remote, _ := callResult["result"].(map[string]any)
	collectionID := stringValue(remote["objectId"])
	if collectionID == "" {
		return cdpResult(id, map[string]any{"nodeIds": []int{}}), nil
	}
	entries, err := s.sendInnerWait(ctx, map[string]any{
		"id":     id,
		"method": "Runtime.getCollectionEntries",
		"params": map[string]any{"objectId": collectionID},
	})
	if err != nil {
		return nil, err
	}
	entriesResult, _ := entries["result"].(map[string]any)
	rawEntries, _ := entriesResult["entries"].([]any)
	nodeIDs := make([]int, 0, len(rawEntries))
	for _, rawEntry := range rawEntries {
		entry, _ := rawEntry.(map[string]any)
		value, _ := entry["value"].(map[string]any)
		nodeObjectID := stringValue(value["objectId"])
		if nodeObjectID == "" {
			continue
		}
		nodeID, err := s.objectIDToNodeID(ctx, id, nodeObjectID)
		if err == nil && nodeID != 0 {
			nodeIDs = append(nodeIDs, nodeID)
		}
	}
	return cdpResult(id, map[string]any{"nodeIds": nodeIDs}), nil
}

func (s *cdpPageSession) domDebuggerGetEventListeners(ctx context.Context, id int, params map[string]any) (map[string]any, error) {
	objectID := stringValue(params["objectId"])
	nodeID, err := s.objectIDToNodeID(ctx, id, objectID)
	if err != nil {
		return nil, err
	}
	response, err := s.sendInnerWait(ctx, map[string]any{
		"id":     id,
		"method": "DOM.getEventListenersForNode",
		"params": map[string]any{"nodeId": nodeID},
	})
	if err != nil {
		return cdpResult(id, map[string]any{"listeners": []map[string]any{}}), nil
	}
	result, _ := response["result"].(map[string]any)
	rawListeners, _ := result["listeners"].([]any)
	listeners := make([]map[string]any, 0, len(rawListeners))
	for _, rawListener := range rawListeners {
		listener, _ := rawListener.(map[string]any)
		out := map[string]any{
			"type":       listener["type"],
			"useCapture": listener["useCapture"],
			"passive":    boolValue(listener["passive"]),
			"once":       boolValue(listener["once"]),
		}
		if location, _ := listener["location"].(map[string]any); location != nil {
			out["scriptId"] = location["scriptId"]
			out["lineNumber"] = location["lineNumber"]
			out["columnNumber"] = location["columnNumber"]
		}
		listeners = append(listeners, out)
	}
	return cdpResult(id, map[string]any{"listeners": listeners}), nil
}

func (s *cdpPageSession) debuggerSetBlackboxPatterns(ctx context.Context, id int, params map[string]any) (map[string]any, error) {
	patterns, _ := params["patterns"].([]any)
	for _, rawPattern := range patterns {
		pattern := stringValue(rawPattern)
		if pattern == "" {
			continue
		}
		_, err := s.sendInnerWait(ctx, map[string]any{
			"id":     id,
			"method": "Debugger.setShouldBlackboxURL",
			"params": map[string]any{"url": pattern, "shouldBlackbox": true},
		})
		if err != nil {
			return nil, err
		}
	}
	return cdpResult(id, map[string]any{}), nil
}

func (s *cdpPageSession) runtimeCompileScript(ctx context.Context, id int, params map[string]any) (map[string]any, error) {
	expression := stringValue(params["expression"])
	response, err := s.sendInnerWait(ctx, map[string]any{
		"id":     id,
		"method": "Runtime.parse",
		"params": map[string]any{"source": expression},
	})
	if err != nil {
		return nil, err
	}
	result, _ := response["result"].(map[string]any)
	if result["result"] == "none" {
		return cdpResult(id, map[string]any{"result": nil}), nil
	}
	parseRange, _ := result["range"].(map[string]any)
	endOffset, _ := numericInt(parseRange["endOffset"])
	if endOffset < 0 || endOffset > len(expression) {
		endOffset = len(expression)
	}
	lines := strings.Split(expression[:endOffset], "\n")
	lineNumber := len(lines) - 1
	columnNumber := 0
	if len(lines) > 0 {
		columnNumber = len(lines[len(lines)-1]) - 1
		if columnNumber < 0 {
			columnNumber = 0
		}
	}
	return cdpResult(id, map[string]any{
		"exceptionDetails": map[string]any{
			"exceptionId":  1,
			"text":         result["message"],
			"lineNumber":   lineNumber,
			"columnNumber": columnNumber,
		},
	}), nil
}

func (s *cdpPageSession) runtimeGetIsolateID(id int) map[string]any {
	executionID := s.defaultExecutionID
	if executionID == nil {
		executionID = 0
	}
	return cdpResult(id, map[string]any{"id": executionID})
}

func (s *cdpPageSession) pageGetNavigationHistory(ctx context.Context, id int) (map[string]any, error) {
	href, err := s.evaluate(ctx, id, "window.location.href", true)
	if err != nil {
		return nil, err
	}
	title, err := s.evaluate(ctx, id, "document.title", true)
	if err != nil {
		return nil, err
	}
	return cdpResult(id, map[string]any{
		"currentIndex": 0,
		"entries": []map[string]any{{
			"id":    0,
			"url":   href["value"],
			"title": title["value"],
		}},
	}), nil
}

func (s *cdpPageSession) pageGetResourceTree(ctx context.Context, message map[string]any) (map[string]any, error) {
	response, err := s.sendInnerWait(ctx, message)
	if err != nil {
		return nil, err
	}
	result, _ := response["result"].(map[string]any)
	frameTree, _ := result["frameTree"].(map[string]any)
	s.frame, _ = frameTree["frame"].(map[string]any)
	return response, nil
}

func (s *cdpPageSession) pageStartScreencast(ctx context.Context, id int, params map[string]any) (map[string]any, error) {
	s.pageStopScreencast(id)
	format := stringValue(params["format"])
	if format == "" {
		format = "jpeg"
	}
	quality, _ := numericInt(params["quality"])
	maxWidth, _ := numericInt(params["maxWidth"])
	maxHeight, _ := numericInt(params["maxHeight"])
	if maxWidth == 0 {
		maxWidth = 1024
	}
	if maxHeight == 0 {
		maxHeight = 768
	}
	remote, err := s.evaluate(ctx, id, "(window.innerWidth > 0 ? window.innerWidth : screen.width) + ',' + (window.innerHeight > 0 ? window.innerHeight : screen.height) + ',' + window.devicePixelRatio", true)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(stringValue(remote["value"]), ",")
	cast := &cdpScreencast{format: format, quality: quality, maxWidth: maxWidth, maxHeight: maxHeight, frameID: 1, lastAck: 0, scale: 1}
	if len(parts) == 3 {
		cast.deviceW = atoi(parts[0])
		cast.deviceH = atoi(parts[1])
		cast.scale = float64(atoi(parts[2]))
		if cast.scale == 0 {
			cast.scale = 1
		}
	}
	castCtx, cancel := context.WithCancel(ctx)
	cast.cancelFunc = cancel
	s.screencast = cast
	go s.screencastLoop(castCtx, id, cast)
	return cdpResult(id, map[string]any{}), nil
}

func (s *cdpPageSession) pageStopScreencast(id int) map[string]any {
	if s.screencast != nil && s.screencast.cancelFunc != nil {
		s.screencast.cancelFunc()
	}
	s.screencast = nil
	return cdpResult(id, map[string]any{})
}

func (s *cdpPageSession) pageScreencastFrameAck(id int, params map[string]any) map[string]any {
	if s.screencast != nil {
		ack, _ := numericInt(params["sessionId"])
		s.screencast.lastAck = ack
	}
	return cdpResult(id, map[string]any{})
}

func (s *cdpPageSession) screencastLoop(ctx context.Context, id int, cast *cdpScreencast) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if cast.frameID > 1 && cast.lastAck < cast.frameID-1 {
			continue
		}
		offsetTop, scrollX, scrollY := s.screencastOffsets(ctx, id)
		response, err := s.sendInnerWait(ctx, map[string]any{
			"id":     id,
			"method": "Page.snapshotRect",
			"params": map[string]any{
				"x":                0,
				"y":                0,
				"width":            cast.deviceW,
				"height":           cast.deviceH,
				"coordinateSystem": "Viewport",
			},
		})
		if err != nil {
			continue
		}
		result, _ := response["result"].(map[string]any)
		dataURL := stringValue(result["dataURL"])
		data := dataURL
		if _, encoded, ok := strings.Cut(dataURL, "base64,"); ok {
			data = encoded
		}
		frameID := cast.frameID
		cast.frameID++
		_ = s.writeJSON(map[string]any{
			"method": "Page.screencastFrame",
			"params": map[string]any{
				"data":      data,
				"sessionId": frameID,
				"metadata": map[string]any{
					"pageScaleFactor": cast.scale,
					"offsetTop":       offsetTop,
					"deviceWidth":     scaledDimension(cast.deviceW, cast.scale, cast.maxWidth),
					"deviceHeight":    scaledDimension(cast.deviceH, cast.scale, cast.maxHeight),
					"scrollOffsetX":   scrollX,
					"scrollOffsetY":   scrollY,
					"timestamp":       float64(time.Now().UnixMilli()) / 1000,
				},
			},
		})
	}
}

func (s *cdpPageSession) screencastOffsets(ctx context.Context, id int) (int, int, int) {
	remote, err := s.evaluate(ctx, id, "window.document.body.offsetTop + ',' + window.pageXOffset + ',' + window.pageYOffset", true)
	if err != nil {
		return 0, 0, 0
	}
	parts := strings.Split(stringValue(remote["value"]), ",")
	if len(parts) != 3 {
		return 0, 0, 0
	}
	return atoi(parts[0]), atoi(parts[1]), atoi(parts[2])
}

func (s *cdpPageSession) inputEmulateTouchFromMouseEvent(ctx context.Context, id int, params map[string]any) (map[string]any, error) {
	eventType := stringValue(params["type"])
	scale := 1.0
	if s.screencast != nil && s.screencast.scale > 0 {
		scale = s.screencast.scale
	}
	x := int(float64(intValue(params["x"])) / scale)
	y := int(float64(intValue(params["y"])) / scale)
	switch eventType {
	case "mouseWheel":
		deltaX := int(float64(intValue(params["deltaX"])) / scale)
		deltaY := int(float64(intValue(params["deltaY"])) / scale)
		_, err := s.evaluate(ctx, id, fmt.Sprintf("window.scrollBy(%d, %d)", -deltaX, -deltaY), true)
		return cdpResult(id, map[string]any{}), err
	case "mouseReleased":
		return cdpResult(id, map[string]any{}), nil
	default:
		modifiers, _ := numericInt(params["modifiers"])
		button := stringValue(params["button"])
		domType := map[string]string{"mousePressed": "click", "mouseMoved": "mousemove"}[eventType]
		if domType == "" {
			domType = "click"
		}
		eventParams := map[string]any{
			"screenX":    x,
			"screenY":    y,
			"clientX":    x,
			"clientY":    y,
			"altKey":     modifiers&1 != 0,
			"ctrlKey":    modifiers&2 != 0,
			"metaKey":    modifiers&4 != 0,
			"shiftKey":   modifiers&8 != 0,
			"button":     button,
			"bubbles":    true,
			"cancelable": false,
		}
		eventJSON := mustJSON(eventParams)
		script := fmt.Sprintf(`(function(type){ const element = document.elementFromPoint(%d, %d); if (!element) return false; const e = new MouseEvent(type, %s); element.dispatchEvent(e); if (element.focus) element.focus(); return true; })(%s)`, x, y, eventJSON, quoteJS(domType))
		if _, err := s.evaluate(ctx, id, script, true); err != nil {
			return nil, err
		}
		if domType == "click" {
			_, _ = s.evaluate(ctx, id, fmt.Sprintf(`(function(){ const element = document.elementFromPoint(%d, %d); if (!element) return false; element.dispatchEvent(new MouseEvent("mouseup", %s)); return true; })()`, x, y, eventJSON), true)
		}
		return cdpResult(id, map[string]any{}), nil
	}
}

func (s *cdpPageSession) inputDispatchKeyEvent(ctx context.Context, id int, params map[string]any) (map[string]any, error) {
	keyType := stringValue(params["type"])
	key := stringValue(params["key"])
	text := stringValue(params["text"])
	var manipulation string
	switch {
	case keyType == "keyUp" && key == "Backspace":
		manipulation = "document.activeElement.value = document.activeElement.value.slice(0, -1);"
	case keyType == "char" && key == "Enter":
		manipulation = `var tagName = document.activeElement.tagName.toLowerCase(); if (tagName === "textarea" || document.activeElement.isContentEditable) { document.activeElement.value = document.activeElement.value + "\n"; } else { const result = document.evaluate("./ancestor-or-self::form", document.activeElement, null, XPathResult.FIRST_ORDERED_NODE_TYPE, null); if (result.singleNodeValue) { const e = result.singleNodeValue.ownerDocument.createEvent("Event"); e.initEvent("submit", true, true); if (result.singleNodeValue.dispatchEvent(e)) { result.singleNodeValue.submit(); } } }`
	case keyType == "char":
		manipulation = "document.activeElement.value = document.activeElement.value + " + quoteJS(text) + ";"
	default:
		return cdpResult(id, map[string]any{}), nil
	}
	script := "function isEditable(element) { if (!element || element.disabled || element.readOnly) return false; var tagName = element.tagName.toLowerCase(); if (tagName === 'textarea' || element.isContentEditable) return true; if (tagName !== 'input') return false; switch (element.type) { case 'color': case 'date': case 'datetime-local': case 'email': case 'file': case 'month': case 'number': case 'password': case 'range': case 'search': case 'tel': case 'text': case 'time': case 'url': case 'week': return true; } return false; } if (isEditable(document.activeElement)) { " + manipulation + " }"
	_, err := s.evaluate(ctx, id, script, true)
	return cdpResult(id, map[string]any{}), err
}

func (s *cdpPageSession) normalizeCDPEvent(message map[string]any) (map[string]any, bool) {
	method, _ := message["method"].(string)
	if method == "Runtime.executionContextCreated" {
		params, _ := message["params"].(map[string]any)
		contextMap, _ := params["context"].(map[string]any)
		if stringValue(contextMap["type"]) == "normal" {
			s.defaultExecutionID = contextMap["id"]
		}
	}
	if method == "Target.targetCreated" {
		params, _ := message["params"].(map[string]any)
		targetInfo, _ := params["targetInfo"].(map[string]any)
		s.targetID = stringValue(targetInfo["targetId"])
		if s.page.URL != "" {
			targetInfo["url"] = s.page.URL
		}
		if s.page.Title != "" {
			targetInfo["title"] = s.page.Title
		}
	}
	if method == "Target.targetDestroyed" {
		if s.frame != nil {
			_ = s.writeJSON(map[string]any{"method": "Page.frameNavigated", "params": map[string]any{"frame": s.frame}})
			_ = s.writeJSON(map[string]any{"method": "Page.loadEventFired", "params": map[string]any{"timestamp": float64(time.Now().UnixMilli()) / 1000}})
		}
		return map[string]any{"method": "DOM.documentUpdated"}, false
	}
	return normalizeCDPEvent(message)
}

func localCDPResponse(message map[string]any, targetID string, sessionID string, page Page) (bool, map[string]any, []map[string]any) {
	id, _ := numericInt(message["id"])
	method, _ := message["method"].(string)
	result := map[string]any{}
	var extra []map[string]any
	switch method {
	case "Target.setAutoAttach":
		extra = append(extra, map[string]any{
			"method": "Target.attachedToTarget",
			"params": map[string]any{
				"sessionId": sessionID,
				"targetInfo": map[string]any{
					"targetId": targetID,
					"type":     "page",
					"title":    page.Title,
					"url":      page.URL,
					"attached": true,
				},
				"waitingForDebugger": true,
			},
		})
	case "Target.setDiscoverTargets",
		"Target.setRemoteLocations",
		"CSS.trackComputedStyleUpdates",
		"DOM.enable",
		"DOMDebugger.setBreakOnCSPViolation",
		"Debugger.setAsyncCallStackDepth",
		"Emulation.setTouchEmulationEnabled",
		"Emulation.setFocusEmulationEnabled",
		"Emulation.setEmulatedVisionDeficiency",
		"Emulation.setEmitTouchEventsForMouse",
		"HeapProfiler.enable",
		"Input.dispatchKeyEvent",
		"Input.emulateTouchFromMouseEvent",
		"Log.startViolationsReport",
		"Network.clearAcceptedEncodingsOverride",
		"Network.setAttachDebugStack",
		"Overlay.enable",
		"Overlay.hideHighlight",
		"Overlay.setPausedInDebuggerMessage",
		"Overlay.setShowContainerQueryOverlays",
		"Overlay.setShowFlexOverlays",
		"Overlay.setShowGridOverlays",
		"Overlay.setShowIsolatedElements",
		"Overlay.setShowScrollSnapOverlays",
		"Overlay.setShowViewportSizeOnResize",
		"Page.screencastFrameAck",
		"Page.startScreencast",
		"Page.stopScreencast",
		"Profiler.enable",
		"Runtime.runIfWaitingForDebugger":
		if method == "Debugger.setAsyncCallStackDepth" {
			result["result"] = true
		}
	case "CSS.takeComputedStyleUpdates":
		result["nodeIds"] = []int{}
	case "Network.loadNetworkResource":
		result["resource"] = map[string]any{"success": true}
	default:
		return false, nil, nil
	}
	return true, map[string]any{"id": id, "result": result}, extra
}

func translateCDPCommand(message map[string]any) map[string]any {
	method, _ := message["method"].(string)
	params, _ := message["params"].(map[string]any)
	switch method {
	case "Audits.enable":
		message["method"] = "Audit.setup"
	case "DOM.getBoxModel", "Overlay.highlightNode":
		message["method"] = "DOM.highlightNode"
		if params != nil && method == "DOM.getBoxModel" {
			params["highlightConfig"] = map[string]any{
				"showInfo":     true,
				"contentColor": map[string]any{"r": 111, "g": 168, "b": 220, "a": 0.66},
				"paddingColor": map[string]any{"r": 147, "g": 196, "b": 125, "a": 0.55},
				"borderColor":  map[string]any{"r": 255, "g": 229, "b": 153, "a": 0.66},
				"marginColor":  map[string]any{"r": 246, "g": 178, "b": 107, "a": 0.66},
			}
		}
	case "Log.clear":
		message["method"] = "Console.clearMessages"
	case "Log.disable":
		message["method"] = "Console.disable"
	case "Log.enable":
		message["method"] = "Console.enable"
	case "Emulation.setEmulatedMedia":
		message["method"] = "Page.setEmulatedMedia"
	case "Emulation.setAutoDarkModeOverride":
		message["method"] = "Page.setForcedAppearance"
		if params != nil {
			enabled, _ := params["enabled"].(bool)
			message["params"] = map[string]any{"appearance": map[bool]string{true: "Dark", false: "Light"}[enabled]}
		}
	case "Network.setCacheDisabled":
		message["method"] = "Network.setResourceCachingDisabled"
		if params != nil {
			message["params"] = map[string]any{"disabled": params["cacheDisabled"]}
		}
	case "ServiceWorker.enable":
		message["method"] = "Worker.enable"
	case "CSS.addRule":
		if params != nil {
			if rule, _ := params["ruleText"].(string); rule != "" {
				params["selector"] = strings.Split(rule, "{")[0]
			}
		}
	case "Debugger.setBreakpointByUrl":
		if params != nil {
			if condition, ok := params["condition"]; ok {
				options, _ := params["options"].(map[string]any)
				if options == nil {
					options = map[string]any{}
				}
				options["condition"] = condition
				params["options"] = options
				delete(params, "condition")
			}
		}
	}
	return message
}

func normalizeCDPEvent(message map[string]any) (map[string]any, bool) {
	method, _ := message["method"].(string)
	params, _ := message["params"].(map[string]any)
	switch method {
	case "Target.targetCreated":
		message["method"] = "Target.targetInfoChanged"
		if targetInfo, ok := params["targetInfo"].(map[string]any); ok {
			if provisional, ok := targetInfo["isProvisional"]; ok {
				targetInfo["attached"] = provisional
				delete(targetInfo, "isProvisional")
			}
		}
	case "Target.didCommitProvisionalTarget", "Page.defaultAppearanceDidChange":
		return message, true
	case "Debugger.globalObjectCleared":
		return map[string]any{"method": "DOM.documentUpdated"}, false
	case "Debugger.paused":
		if reason := stringValue(params["reason"]); reason != "" {
			params["reason"] = debuggerPausedReason(reason)
		}
		if data, _ := params["data"].(map[string]any); data != nil {
			if breakpointID := stringValue(data["breakpointId"]); breakpointID != "" {
				params["hitBreakpoints"] = []string{breakpointID}
			}
		}
	case "Debugger.scriptFailedToParse":
		source := stringValue(params["scriptSource"])
		contextID := params["executionContextId"]
		if contextID == nil {
			contextID = 0
		}
		params["endColumn"] = 0
		params["endLine"] = params["errorLine"]
		params["executionContextId"] = contextID
		params["startColumn"] = 0
		params["startLine"] = params["startLine"]
		sourceHash := fmt.Sprintf("%x", sha1.Sum([]byte(source)))
		params["scriptId"] = sourceHash
		params["hash"] = sourceHash
		delete(params, "errorLine")
		delete(params, "scriptSource")
	case "Runtime.executionContextCreated":
		if contextMap, ok := params["context"].(map[string]any); ok {
			params["context"] = map[string]any{
				"id":       contextMap["id"],
				"origin":   "default",
				"name":     "",
				"uniqueId": contextMap["frameId"],
			}
		}
	case "Console.messageAdded":
		entry := map[string]any{"source": "javascript", "level": "info", "timestamp": float64(time.Now().UnixMilli()) / 1000}
		if consoleMessage, ok := params["message"].(map[string]any); ok {
			entry["source"] = logSource(stringValue(consoleMessage["source"]))
			entry["level"] = logLevel(stringValue(consoleMessage["level"]))
			entry["text"] = consoleMessage["text"]
			if url := stringValue(consoleMessage["url"]); url != "" {
				entry["url"] = url
			}
			if line, ok := numericInt(consoleMessage["line"]); ok {
				entry["lineNumber"] = line
			}
			if requestID := stringValue(consoleMessage["networkRequestId"]); requestID != "" {
				entry["networkRequestId"] = requestID
			}
		}
		return map[string]any{"method": "Log.entryAdded", "params": map[string]any{"entry": entry}}, false
	case "Network.responseReceived":
		if response, ok := params["response"].(map[string]any); ok {
			resourceType := stringValue(params["type"])
			if !validNetworkResourceType(resourceType) {
				resourceType = "Other"
			}
			normalized := map[string]any{
				"loaderId":  params["loaderId"],
				"requestId": params["requestId"],
				"timestamp": params["timestamp"],
				"type":      resourceType,
				"response": map[string]any{
					"url":               response["url"],
					"status":            response["status"],
					"statusText":        response["statusText"],
					"headers":           response["headers"],
					"mimeType":          response["mimeType"],
					"connectionReused":  false,
					"encodedDataLength": 0,
					"securityState":     "unknown",
				},
			}
			if frameID := params["frameId"]; frameID != nil {
				normalized["frameId"] = frameID
			}
			message["params"] = normalized
		}
	case "Network.loadingFinished":
		metrics, _ := params["metrics"].(map[string]any)
		headerSize, _ := numericInt(metrics["responseHeaderBytesReceived"])
		bodySize, _ := numericInt(metrics["responseBodyBytesReceived"])
		message["params"] = map[string]any{
			"encodedDataLength": headerSize + bodySize,
			"requestId":         params["requestId"],
			"timestamp":         params["timestamp"],
		}
	}
	return message, false
}

func logSource(source string) string {
	switch source {
	case "xml", "javascript", "network", "storage", "appcache", "rendering", "security", "deprecation", "worker", "violation", "intervention", "recommendation", "other":
		return source
	case "console-api":
		return "javascript"
	case "css":
		return "rendering"
	case "content-blocker", "media", "mediasource", "webrtc", "itp-debug", "ad-click-attribution":
		return "other"
	default:
		return "other"
	}
}

func logLevel(level string) string {
	switch level {
	case "log", "info":
		return "info"
	case "warning":
		return "warning"
	case "error":
		return "error"
	case "debug":
		return "verbose"
	default:
		return "info"
	}
}

func validNetworkResourceType(resourceType string) bool {
	switch resourceType {
	case "Document", "Stylesheet", "Image", "Media", "Font", "Script", "TextTrack", "XHR", "Fetch", "EventSource", "WebSocket", "Manifest", "SignedExchange", "Ping", "CSPViolationReport", "Preflight", "Other":
		return true
	default:
		return false
	}
}

func cdpResult(id int, result map[string]any) map[string]any {
	if result == nil {
		result = map[string]any{}
	}
	return map[string]any{"id": id, "result": result}
}

func quoteJS(value string) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return `""`
	}
	return string(encoded)
}

func atoi(value string) int {
	var out int
	_, _ = fmt.Sscanf(strings.TrimSpace(value), "%d", &out)
	return out
}

func scaledDimension(value int, scale float64, maxValue int) int {
	if value <= 0 {
		return 0
	}
	scaled := int(float64(value) * scale)
	if maxValue > 0 && scaled > maxValue {
		return maxValue
	}
	if scaled <= 0 {
		return value
	}
	return scaled
}

func debuggerPausedReason(reason string) string {
	switch reason {
	case "XHR":
		return "XHR"
	case "DOM":
		return "DOM"
	case "Listener":
		return "EventListener"
	case "exception":
		return "exception"
	case "assert":
		return "assert"
	case "CSPViolation":
		return "CSPViolation"
	case "DebuggerStatement":
		return "debugCommand"
	case "Breakpoint", "PauseOnNextStatement":
		return "instrumentation"
	default:
		return "other"
	}
}

func cdpError(id int, err error) map[string]any {
	return map[string]any{
		"id": id,
		"error": map[string]any{
			"code":    -32000,
			"message": err.Error(),
		},
	}
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}

func mustJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(encoded)
}

func nextWIRID() int {
	return int(time.Now().UnixNano() & 0x7fffffff)
}
