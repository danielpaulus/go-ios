//go:build e2e

package harness

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/webinspector"
	"github.com/gorilla/websocket"
)

// RunWebInspectorBrowserControl exercises Safari WebInspector, the local CDP
// bridge, browser input, DOM helpers, and screencast against one real device.
func RunWebInspectorBrowserControl(t *testing.T, udid string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	client, app, page := openWebInspectorFixture(t, ctx, udid)

	value, err := client.Evaluate(ctx, app, page, "document.title")
	if err != nil {
		t.Fatalf("webinspector evaluate document.title: %v", err)
	}
	if value != "go-ios-e2e" {
		t.Fatalf("document.title = %#v, want go-ios-e2e", value)
	}

	port := freeTCPPort(t)
	server := webinspector.NewCDPServer(client, "127.0.0.1", port)
	serverCtx, stopServer := context.WithCancel(ctx)
	defer stopServer()
	errCh := make(chan error, 1)
	go func() { errCh <- server.Serve(serverCtx) }()
	waitForCDPServer(t, server.Addr())
	assertCDPListsFixturePage(t, server.Addr(), page.Key)

	ws := connectCDPPage(t, server.Addr(), page.Key)
	defer ws.Close()

	cdpCommand(t, ws, 1, "Runtime.evaluate", map[string]any{
		"expression":    `document.body.innerHTML = '<input id="agent" value=""><button id="tap" style="position:absolute;left:0;top:0;width:120px;height:48px;color:rgb(1, 2, 3)">tap</button>'; window.clicked = 0; document.getElementById("tap").addEventListener("click", () => { window.clicked += 1; }); document.getElementById("agent").focus();`,
		"returnByValue": true,
	})
	cdpCommand(t, ws, 2, "Input.dispatchKeyEvent", map[string]any{
		"type": "char",
		"key":  "a",
		"text": "a",
	})
	valueResponse := cdpCommand(t, ws, 3, "Runtime.evaluate", map[string]any{
		"expression":    `document.getElementById("agent").value`,
		"returnByValue": true,
	})
	if got := cdpEvaluateValue(valueResponse); got != "a" {
		t.Fatalf("input value after CDP key event = %#v, want a", got)
	}

	cdpCommand(t, ws, 4, "Input.emulateTouchFromMouseEvent", map[string]any{
		"type":      "mousePressed",
		"x":         10,
		"y":         10,
		"button":    "left",
		"modifiers": 0,
	})
	clickedResponse := cdpCommand(t, ws, 5, "Runtime.evaluate", map[string]any{
		"expression":    `window.clicked`,
		"returnByValue": true,
	})
	if got := cdpEvaluateValue(clickedResponse); got != float64(1) {
		t.Fatalf("window.clicked after CDP mouse event = %#v, want 1", got)
	}

	nodeResponse := cdpCommand(t, ws, 6, "DOM.getNodeForLocation", map[string]any{"x": 10, "y": 10})
	nodeResult, _ := nodeResponse["result"].(map[string]any)
	if nodeID, ok := numeric(nodeResult["nodeId"]); !ok || nodeID == 0 {
		t.Fatalf("DOM.getNodeForLocation returned no nodeId: %#v", nodeResponse)
	}

	bodyResponse := cdpCommand(t, ws, 7, "Runtime.evaluate", map[string]any{"expression": `document.body`})
	bodyID := cdpEvaluateObjectID(bodyResponse)
	if bodyID == "" {
		t.Fatalf("Runtime.evaluate document.body returned no objectId: %#v", bodyResponse)
	}
	bodyNodeResponse := cdpCommand(t, ws, 8, "DOM.requestNode", map[string]any{"objectId": bodyID})
	bodyNodeResult, _ := bodyNodeResponse["result"].(map[string]any)
	bodyNodeID, ok := numeric(bodyNodeResult["nodeId"])
	if !ok || bodyNodeID == 0 {
		t.Fatalf("DOM.requestNode returned no body nodeId: %#v", bodyNodeResponse)
	}
	styleResponse := cdpCommand(t, ws, 9, "DOM.getNodesForSubtreeByStyle", map[string]any{
		"nodeId": bodyNodeID,
		"computedStyles": []map[string]any{
			{"name": "color", "value": "rgb(1, 2, 3)"},
		},
	})
	styleResult, _ := styleResponse["result"].(map[string]any)
	styleNodeIDs, _ := styleResult["nodeIds"].([]any)
	if len(styleNodeIDs) == 0 {
		t.Fatalf("DOM.getNodesForSubtreeByStyle returned no matching nodes: %#v", styleResponse)
	}

	objectResponse := cdpCommand(t, ws, 10, "Runtime.evaluate", map[string]any{"expression": `document.getElementById("agent")`})
	objectID := cdpEvaluateObjectID(objectResponse)
	if objectID == "" {
		t.Fatalf("Runtime.evaluate returned no objectId: %#v", objectResponse)
	}
	listenersResponse := cdpCommand(t, ws, 11, "DOMDebugger.getEventListeners", map[string]any{"objectId": objectID})
	listenersResult, _ := listenersResponse["result"].(map[string]any)
	listeners, ok := listenersResult["listeners"].([]any)
	if !ok {
		t.Fatalf("DOMDebugger.getEventListeners returned no listeners array: %#v", listenersResponse)
	}
	if !hasEventListener(listeners, "click") {
		t.Fatalf("DOMDebugger.getEventListeners did not report the click listener: %#v", listenersResponse)
	}

	resourceTree := cdpCommand(t, ws, 12, "Page.getResourceTree", map[string]any{})
	resourceResult, _ := resourceTree["result"].(map[string]any)
	if _, ok := resourceResult["frameTree"].(map[string]any); !ok {
		t.Fatalf("Page.getResourceTree returned no frameTree: %#v", resourceTree)
	}

	cdpCommand(t, ws, 13, "Page.startScreencast", map[string]any{
		"format":    "jpeg",
		"quality":   80,
		"maxWidth":  400,
		"maxHeight": 400,
	})
	frame := cdpEvent(t, ws, "Page.screencastFrame")
	frameParams, _ := frame["params"].(map[string]any)
	if data, _ := frameParams["data"].(string); data == "" {
		t.Fatalf("Page.screencastFrame has empty data: %#v", frame)
	} else if format, width, height, ok := decodeScreencastImage(data); !ok {
		t.Fatalf("Page.screencastFrame data is not a decodable PNG/JPEG payload")
	} else if width == 0 || height == 0 {
		t.Fatalf("Page.screencastFrame decoded as %s with invalid dimensions %dx%d", format, width, height)
	}
	sessionID, _ := numeric(frameParams["sessionId"])
	cdpCommand(t, ws, 14, "Page.screencastFrameAck", map[string]any{"sessionId": sessionID})
	cdpCommand(t, ws, 15, "Page.stopScreencast", map[string]any{})

	_ = ws.Close()
	stopServer()
	select {
	case err := <-errCh:
		if err != nil && !strings.Contains(err.Error(), "context canceled") {
			t.Fatalf("cdp server: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("cdp server did not stop")
	}
}

func openWebInspectorFixture(t *testing.T, ctx context.Context, udid string) (*webinspector.Client, webinspector.Application, webinspector.Page) {
	t.Helper()

	device, err := ios.GetDevice(udid)
	if err != nil {
		t.Fatalf("get device: %v", err)
	}
	client, err := webinspector.New(device)
	if err != nil {
		t.Fatalf("connect webinspector service: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	if err := client.Connect(ctx); err != nil {
		skipWebInspectorSetting(t, err)
		t.Fatalf("start webinspector: %v", err)
	}

	app, err := client.OpenApp(ctx, webinspector.SafariBundleID)
	if err != nil {
		t.Fatalf("open Safari: %v", err)
	}
	session, err := client.AutomationSession(ctx, app)
	if err != nil {
		skipWebInspectorSetting(t, err)
		t.Fatalf("start remote automation: %v", err)
	}
	t.Cleanup(func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = session.Stop(stopCtx)
	})
	if err := session.Start(ctx); err != nil {
		t.Fatalf("start browsing context: %v", err)
	}
	url := "data:text/html,%3Ctitle%3Ego-ios-e2e%3C/title%3E%3Cinput%20id%3Dagent%20value%3D%22%22%3E"
	if err := session.Navigate(ctx, url); err != nil {
		t.Fatalf("navigate fixture page: %v", err)
	}

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		pages, err := client.ListPages(ctx, 500*time.Millisecond)
		if err != nil {
			t.Fatalf("list webinspector pages: %v", err)
		}
		for _, candidate := range pages {
			if candidate.Application.BundleID != webinspector.SafariBundleID {
				continue
			}
			if candidate.Page.Type != webinspector.WIRTypeWeb && candidate.Page.Type != webinspector.WIRTypeWebPage {
				continue
			}
			if strings.Contains(candidate.Page.URL, "go-ios-e2e") || candidate.Page.Title == "go-ios-e2e" {
				return client, candidate.Application, candidate.Page
			}
		}
		time.Sleep(250 * time.Millisecond)
	}
	t.Fatalf("fixture page did not appear in WebInspector listing")
	return nil, webinspector.Application{}, webinspector.Page{}
}

func skipWebInspectorSetting(t *testing.T, err error) {
	t.Helper()
	message := strings.ToLower(err.Error())
	if strings.Contains(message, "web inspector is not enabled") ||
		strings.Contains(message, "remote automation is not enabled") {
		t.Skip(err)
	}
}

func freeTCPPort(t *testing.T) int {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on free port: %v", err)
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port
}

func waitForCDPServer(t *testing.T, addr string) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get("http://" + addr + "/json/version")
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("cdp server %s did not become ready", addr)
}

func assertCDPListsFixturePage(t *testing.T, addr string, pageKey string) {
	t.Helper()
	resp, err := http.Get("http://" + addr + "/json/list")
	if err != nil {
		t.Fatalf("get CDP target list: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("CDP target list status = %d", resp.StatusCode)
	}
	var targets []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&targets); err != nil {
		t.Fatalf("decode CDP target list: %v", err)
	}
	for _, target := range targets {
		if target["id"] == pageKey && target["type"] == "page" && target["webSocketDebuggerUrl"] != "" {
			return
		}
	}
	t.Fatalf("CDP target list did not include page %s: %#v", pageKey, targets)
}

func connectCDPPage(t *testing.T, addr string, pageKey string) *websocket.Conn {
	t.Helper()
	ws, _, err := websocket.DefaultDialer.Dial("ws://"+addr+"/devtools/page/"+pageKey, nil)
	if err != nil {
		t.Fatalf("connect CDP websocket: %v", err)
	}
	return ws
}

func cdpCommand(t *testing.T, ws *websocket.Conn, id int, method string, params map[string]any) map[string]any {
	t.Helper()
	if err := ws.WriteJSON(map[string]any{"id": id, "method": method, "params": params}); err != nil {
		t.Fatalf("write CDP command %s: %v", method, err)
	}
	for {
		message := readCDPMessage(t, ws)
		if gotID, ok := numeric(message["id"]); ok && gotID == id {
			if errValue, ok := message["error"]; ok {
				t.Fatalf("CDP command %s failed: %#v", method, errValue)
			}
			return message
		}
	}
}

func cdpEvent(t *testing.T, ws *websocket.Conn, method string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		message := readCDPMessage(t, ws)
		if message["method"] == method {
			return message
		}
	}
	t.Fatalf("timed out waiting for CDP event %s", method)
	return nil
}

func readCDPMessage(t *testing.T, ws *websocket.Conn) map[string]any {
	t.Helper()
	_ = ws.SetReadDeadline(time.Now().Add(10 * time.Second))
	var message map[string]any
	if err := ws.ReadJSON(&message); err != nil {
		t.Fatalf("read CDP message: %v", err)
	}
	return message
}

func cdpEvaluateValue(response map[string]any) any {
	result, _ := response["result"].(map[string]any)
	remote, _ := result["result"].(map[string]any)
	return remote["value"]
}

func cdpEvaluateObjectID(response map[string]any) string {
	result, _ := response["result"].(map[string]any)
	remote, _ := result["result"].(map[string]any)
	objectID, _ := remote["objectId"].(string)
	return objectID
}

func hasEventListener(listeners []any, listenerType string) bool {
	for _, rawListener := range listeners {
		listener, _ := rawListener.(map[string]any)
		if listener["type"] == listenerType {
			return true
		}
	}
	return false
}

func decodeScreencastImage(data string) (string, int, int, bool) {
	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil || len(decoded) < 8 {
		return "", 0, 0, false
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(decoded))
	if err != nil {
		return "", 0, 0, false
	}
	if format != "jpeg" && format != "png" {
		return format, config.Width, config.Height, false
	}
	return format, config.Width, config.Height, true
}

func numeric(value any) (int, bool) {
	switch v := value.(type) {
	case int:
		return v, true
	case float64:
		return int(v), true
	default:
		return 0, false
	}
}
