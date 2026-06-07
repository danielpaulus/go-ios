package webinspector

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type AutomationSession struct {
	client    *Client
	app       Application
	page      Page
	sessionID string
	nextID    int
	topHandle string
	mu        sync.Mutex
}

func (s *AutomationSession) ID() string {
	return s.sessionID
}

func (s *AutomationSession) Start(ctx context.Context) error {
	response, err := s.send(ctx, "createBrowsingContext", nil)
	if err != nil {
		return err
	}
	result, _ := response["result"].(map[string]any)
	handle, _ := result["handle"].(string)
	if handle == "" {
		return fmt.Errorf("automation start: missing browsing context handle: %#v", response)
	}
	s.topHandle = handle
	return nil
}

func (s *AutomationSession) Stop(ctx context.Context) error {
	handles, err := s.WindowHandles(ctx)
	if err != nil {
		return err
	}
	for _, handle := range handles {
		_, _ = s.send(ctx, "closeBrowsingContext", map[string]any{"handle": handle})
	}
	s.topHandle = ""
	return nil
}

func (s *AutomationSession) Navigate(ctx context.Context, url string) error {
	if s.topHandle == "" {
		if err := s.Start(ctx); err != nil {
			return err
		}
	}
	_, err := s.send(ctx, "navigateBrowsingContext", map[string]any{
		"handle":          s.topHandle,
		"pageLoadTimeout": 3000000,
		"url":             url,
	})
	return err
}

func (s *AutomationSession) ExecuteScript(ctx context.Context, script string, args ...any) (any, error) {
	if s.topHandle == "" {
		if err := s.Start(ctx); err != nil {
			return nil, err
		}
	}
	encodedArgs := make([]string, 0, len(args))
	for _, arg := range args {
		encoded, err := json.Marshal(arg)
		if err != nil {
			return nil, err
		}
		encodedArgs = append(encodedArgs, string(encoded))
	}
	response, err := s.send(ctx, "evaluateJavaScriptFunction", map[string]any{
		"browsingContextHandle": s.topHandle,
		"function":              "function(){\n" + script + "\n}",
		"arguments":             encodedArgs,
	})
	if err != nil {
		return nil, err
	}
	result, _ := response["result"].(map[string]any)
	raw, _ := result["result"].(string)
	if raw == "" {
		return nil, nil
	}
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return raw, nil
	}
	return value, nil
}

func (s *AutomationSession) CurrentURL(ctx context.Context) (string, error) {
	value, err := s.ExecuteScript(ctx, "return window.location.href")
	if err != nil {
		return "", err
	}
	url, _ := value.(string)
	return url, nil
}

func (s *AutomationSession) Title(ctx context.Context) (string, error) {
	value, err := s.ExecuteScript(ctx, "return document.title")
	if err != nil {
		return "", err
	}
	title, _ := value.(string)
	return title, nil
}

func (s *AutomationSession) WindowHandles(ctx context.Context) ([]string, error) {
	response, err := s.send(ctx, "getBrowsingContexts", nil)
	if err != nil {
		return nil, err
	}
	result, _ := response["result"].(map[string]any)
	contexts, _ := result["contexts"].([]any)
	handles := make([]string, 0, len(contexts))
	for _, rawContext := range contexts {
		contextMap, _ := rawContext.(map[string]any)
		handle, _ := contextMap["handle"].(string)
		if handle != "" {
			handles = append(handles, handle)
		}
	}
	return handles, nil
}

func (s *AutomationSession) send(ctx context.Context, method string, params map[string]any) (map[string]any, error) {
	if params == nil {
		params = map[string]any{}
	}
	s.mu.Lock()
	id := s.nextID
	s.nextID++
	s.mu.Unlock()
	return s.client.SendCommand(ctx, s.sessionID, s.app, s.page, id, "Automation."+method, params)
}
