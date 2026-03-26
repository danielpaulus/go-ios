// Package debuggertools provides tools for interacting with iOS processes
// via the debug proxy and GDB RSP — ObjC runtime calls, memory access,
// and view hierarchy capture. No external tools (LLDB, Xcode) required.
package debuggertools

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/danielpaulus/go-ios/ios/debugserver"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
)

// DefaultProperties is the default set of properties to capture.
var DefaultProperties = []string{
	"dbgFormattedDisplayName",
	"frame",
	"bounds",
	"position",
	"accessibilityLabel",
	"accessibilityIdentifier",
	"accessibilityValue",
	"title",
	"text",
}

// ViewHierarchySession holds a prepared debugger session for repeated view hierarchy captures.
// Call SetupViewHierarchy() once to inject frameworks and resolve symbols,
// then Dump() for each capture.
type ViewHierarchySession struct {
	rt      *objCRuntime
	hub     uint64 // [DBGTargetHub sharedHub] singleton
	request uint64 // NSString with the base64 request (reused across dumps)
}

// SetupViewHierarchy prepares a session for repeated view hierarchy captures.
// The GDB connection must already be attached and the process stopped.
// Pass nil or empty properties to use DefaultProperties.
// Call Dump() to capture, and Close() when done.
func SetupViewHierarchy(gdb *debugserver.GDBServer, properties []string) (*ViewHierarchySession, error) {
	rt, err := newObjCRuntime(gdb)
	if err != nil {
		return nil, err
	}

	// Load DebugHierarchyFoundation and its view registration support
	rt.Dlopen("/System/Library/PrivateFrameworks/DebugHierarchyFoundation.framework/DebugHierarchyFoundation")
	rt.Dlopen("/usr/lib/libViewDebuggerSupport.dylib")
	rt.Dlopen("/Developer/Library/PrivateFrameworks/DTDDISupport.framework/libViewDebuggerSupport.dylib")

	// hub = [DBGTargetHub sharedHub]
	hub, err := rt.ClassCall("DBGTargetHub", "sharedHub")
	if err != nil {
		rt.cleanup()
		return nil, fmt.Errorf("[DBGTargetHub sharedHub]: %w", err)
	}

	// Pre-build the NSString request (reused across dumps)
	if len(properties) == 0 {
		properties = DefaultProperties
	}
	b64 := buildRequestBase64(properties)
	request, err := rt.ClassCall("NSString", "stringWithUTF8String:", rt.CString(b64))
	if err != nil {
		rt.cleanup()
		return nil, fmt.Errorf("[NSString stringWithUTF8String:]: %w", err)
	}

	return &ViewHierarchySession{rt: rt, hub: hub, request: request}, nil
}

// Dump captures the current view hierarchy and returns decompressed JSON.
// The process must be stopped before calling Dump (e.g. via interrupt 0x03).
func (s *ViewHierarchySession) Dump() ([]byte, error) {
	log.Info("Capturing view hierarchy...")
	nsData, err := s.rt.Call(s.hub, "performRequestWithRequestInBase64:", s.request)
	if err != nil {
		return nil, fmt.Errorf("[hub performRequestWithRequestInBase64:]: %w", err)
	}

	dataLen, _ := s.rt.Call(nsData, "length")
	bytesPtr, _ := s.rt.Call(nsData, "bytes")
	if dataLen == 0 || bytesPtr == 0 {
		return nil, fmt.Errorf("NSData empty (length=%d ptr=0x%x)", dataLen, bytesPtr)
	}

	log.WithField("bytes", dataLen).Info("Reading hierarchy data")
	rawData, err := s.rt.mem.readMemory(bytesPtr, dataLen)
	if err != nil {
		return nil, fmt.Errorf("read memory: %w", err)
	}

	if gz, err := gzip.NewReader(bytes.NewReader(rawData)); err == nil {
		if decompressed, err := io.ReadAll(gz); err == nil {
			return decompressed, nil
		}
	}
	return rawData, nil
}

// Close cleans up the session (restores registers, frees allocated memory).
// Does NOT detach from the process — that's the caller's responsibility.
func (s *ViewHierarchySession) Close() {
	s.rt.cleanup()
}

// CaptureViewHierarchy is a convenience function that sets up, dumps once, and cleans up.
func CaptureViewHierarchy(gdb *debugserver.GDBServer, properties []string) ([]byte, error) {
	session, err := SetupViewHierarchy(gdb, properties)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	return session.Dump()
}

func buildRequestBase64(properties []string) string {
	propList := make([]interface{}, len(properties))
	for i, p := range properties {
		propList[i] = p
	}

	request := map[string]interface{}{
		"DBGHierarchyRequestName":                 "Initial request",
		"DBGHierarchyRequestInitiatorVersionKey":  4,
		"DBGHierarchyRequestPriority":             0,
		"DBGHierarchyObjectDiscovery":             1,
		"DBGHierarchyRequestIdentifier":           uuid.New().String(),
		"DBGHierarchyRequestTransportCompression": true,
		"DBGHierarchyRequestActions": []interface{}{
			map[string]interface{}{"actionClass": "DebugHierarchyResetAction"},
			map[string]interface{}{
				"actionClass":                   "DebugHierarchyPropertyAction",
				"propertyNames":                 propList,
				"visibility":                    15,
				"options":                       0,
				"optionsComparisonStyle":        0,
				"exactTypesAreExclusive":        false,
				"typesAreExclusive":             false,
				"objectIdentifiersAreExclusive": false,
				"propertyNamesAreExclusive":     false,
			},
		},
	}
	j, _ := json.Marshal(request)
	return base64.StdEncoding.EncodeToString(j)
}
