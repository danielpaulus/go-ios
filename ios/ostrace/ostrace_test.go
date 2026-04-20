package ostrace

import (
	"encoding/binary"
	"testing"
	"time"
)

func TestParseEntry(t *testing.T) {
	// Build a synthetic binary entry matching the libimobiledevice ostrace_packet_header_t struct
	data := make([]byte, 300)

	// pid at offset 9 (uint32 LE)
	binary.LittleEndian.PutUint32(data[9:13], 1234)

	// procpathLen at offset 37 (uint16 LE) — "/usr/lib/dyld\x00" = 14 bytes
	binary.LittleEndian.PutUint16(data[37:39], 14)

	// timeSec at offset 55 (uint64 LE) — 2024-01-15 10:30:00 UTC = 1705312200
	binary.LittleEndian.PutUint64(data[55:63], 1705312200)

	// timeUsec at offset 63 (uint32 LE)
	binary.LittleEndian.PutUint32(data[63:67], 500000)

	// level at offset 68
	data[68] = byte(LogLevelError)

	// threadID at offset 83 (uint32 LE)
	binary.LittleEndian.PutUint32(data[83:87], 4567)

	// imagepathLen at offset 107 (uint16 LE) — "MyApp\x00" = 6 bytes
	binary.LittleEndian.PutUint16(data[107:109], 6)

	// messageLen at offset 109 (uint32 LE) — "hello world\x00" = 12 bytes
	binary.LittleEndian.PutUint32(data[109:113], 12)

	// senderImageOffset at offset 113 (uint32 LE)
	binary.LittleEndian.PutUint32(data[113:117], 0x1234)

	// subsystemLen at offset 117 (uint16 LE) — "com.example.app\x00" = 16 bytes
	binary.LittleEndian.PutUint16(data[117:119], 16)

	// categoryLen at offset 121 (uint16 LE) — "networking\x00" = 11 bytes
	binary.LittleEndian.PutUint16(data[121:123], 11)

	// Variable-length fields starting at offset 129
	offset := 129

	// procpath (14 bytes including null)
	copy(data[offset:], "/usr/lib/dyld\x00")
	offset += 14

	// imagepath (6 bytes including null)
	copy(data[offset:], "MyApp\x00")
	offset += 6

	// message (12 bytes including null)
	copy(data[offset:], "hello world\x00")
	offset += 12

	// subsystem (16 bytes including null)
	copy(data[offset:], "com.example.app\x00")
	offset += 16

	// category (11 bytes including null)
	copy(data[offset:], "networking\x00")
	offset += 11

	data = data[:offset]

	entry, err := parseEntry(data)
	if err != nil {
		t.Fatalf("parseEntry failed: %v", err)
	}

	if entry.PID != 1234 {
		t.Errorf("PID = %d, want 1234", entry.PID)
	}

	expectedTime := time.Unix(1705312200, 500000*1000)
	if !entry.Timestamp.Equal(expectedTime) {
		t.Errorf("Timestamp = %v, want %v", entry.Timestamp, expectedTime)
	}

	if entry.Level != LogLevelError {
		t.Errorf("Level = %d, want %d (Error)", entry.Level, LogLevelError)
	}

	if entry.LevelName != "Error" {
		t.Errorf("LevelName = %q, want %q", entry.LevelName, "Error")
	}

	if entry.ThreadID != 4567 {
		t.Errorf("ThreadID = %d, want 4567", entry.ThreadID)
	}

	if entry.ImageName != "MyApp" {
		t.Errorf("ImageName = %q, want %q", entry.ImageName, "MyApp")
	}

	if entry.ImageOffset != 0x1234 {
		t.Errorf("ImageOffset = 0x%x, want 0x1234", entry.ImageOffset)
	}

	if entry.Filename != "/usr/lib/dyld" {
		t.Errorf("Filename = %q, want %q", entry.Filename, "/usr/lib/dyld")
	}

	if entry.Message != "hello world" {
		t.Errorf("Message = %q, want %q", entry.Message, "hello world")
	}

	if entry.Label == nil {
		t.Fatal("Label is nil, expected non-nil")
	}

	if entry.Label.Subsystem != "com.example.app" {
		t.Errorf("Label.Subsystem = %q, want %q", entry.Label.Subsystem, "com.example.app")
	}

	if entry.Label.Category != "networking" {
		t.Errorf("Label.Category = %q, want %q", entry.Label.Category, "networking")
	}
}

func TestParseEntryNoLabel(t *testing.T) {
	data := make([]byte, 200)

	binary.LittleEndian.PutUint32(data[9:13], 42)
	binary.LittleEndian.PutUint16(data[37:39], 5) // "main\x00"
	binary.LittleEndian.PutUint64(data[55:63], 1000000)
	binary.LittleEndian.PutUint32(data[63:67], 0)
	data[68] = byte(LogLevelInfo)
	binary.LittleEndian.PutUint16(data[107:109], 4) // "foo\x00"
	binary.LittleEndian.PutUint32(data[109:113], 4) // "bar\x00"
	// subsystemLen = 0, categoryLen = 0 → no label
	binary.LittleEndian.PutUint16(data[117:119], 0)
	binary.LittleEndian.PutUint16(data[121:123], 0)

	offset := 129
	copy(data[offset:], "main\x00")
	offset += 5
	copy(data[offset:], "foo\x00")
	offset += 4
	copy(data[offset:], "bar\x00")
	offset += 4
	data = data[:offset]

	entry, err := parseEntry(data)
	if err != nil {
		t.Fatalf("parseEntry failed: %v", err)
	}

	if entry.PID != 42 {
		t.Errorf("PID = %d, want 42", entry.PID)
	}

	if entry.Label != nil {
		t.Errorf("Label = %+v, want nil", entry.Label)
	}
}

func TestParseEntryTooShort(t *testing.T) {
	data := make([]byte, 50)
	_, err := parseEntry(data)
	if err == nil {
		t.Error("expected error for short data, got nil")
	}
}

func TestLogLevelString(t *testing.T) {
	tests := []struct {
		level LogLevel
		want  string
	}{
		{LogLevelDefault, "Default"},
		{LogLevelInfo, "Info"},
		{LogLevelDebug, "Debug"},
		{LogLevelUserAction, "UserAction"},
		{LogLevelError, "Error"},
		{LogLevelFault, "Fault"},
		{LogLevel(0xFF), "Unknown(255)"},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("LogLevel(%d).String() = %q, want %q", tt.level, got, tt.want)
		}
	}
}

func TestCstring(t *testing.T) {
	tests := []struct {
		input []byte
		want  string
	}{
		{[]byte("hello\x00"), "hello"},
		{[]byte("hello"), "hello"},
		{[]byte{}, ""},
		{[]byte{0}, ""},
	}

	for _, tt := range tests {
		if got := cstring(tt.input); got != tt.want {
			t.Errorf("cstring(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseLevelFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantMF   uint16
		wantSF   uint32
		wantLvls []LogLevel
		wantErr  bool
	}{
		{
			name:     "empty → default filter",
			input:    "",
			wantMF:   MessageFilterLogMessage,
			wantSF:   StreamFlagsAll,
			wantLvls: nil,
		},
		{
			name:     "error,fault → no severity gates needed",
			input:    "error,fault",
			wantMF:   MessageFilterLogMessage,
			wantSF:   0,
			wantLvls: []LogLevel{LogLevelError, LogLevelFault},
		},
		{
			name:     "default → no severity gates",
			input:    "default",
			wantMF:   MessageFilterLogMessage,
			wantSF:   0,
			wantLvls: []LogLevel{LogLevelDefault},
		},
		{
			name:     "info → Debug stream bit (Info piggybacks on it)",
			input:    "info",
			wantMF:   MessageFilterLogMessage,
			wantSF:   StreamFlagsDebug,
			wantLvls: []LogLevel{LogLevelInfo},
		},
		{
			name:     "debug → Debug stream bit",
			input:    "debug",
			wantMF:   MessageFilterLogMessage,
			wantSF:   StreamFlagsDebug,
			wantLvls: []LogLevel{LogLevelDebug},
		},
		{
			name:     "all five → Debug bit enables Info+Debug",
			input:    "default,info,debug,error,fault",
			wantMF:   MessageFilterLogMessage,
			wantSF:   StreamFlagsDebug,
			wantLvls: []LogLevel{LogLevelDefault, LogLevelInfo, LogLevelDebug, LogLevelError, LogLevelFault},
		},
		{
			name:     "case-insensitive and deduped",
			input:    "Error, FAULT, error",
			wantMF:   MessageFilterLogMessage,
			wantSF:   0,
			wantLvls: []LogLevel{LogLevelError, LogLevelFault},
		},
		{
			name:    "bogus name",
			input:   "bogus",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		got, err := ParseLevelFilter(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("%s: ParseLevelFilter(%q) expected error, got nil", tt.name, tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: ParseLevelFilter(%q) unexpected error: %v", tt.name, tt.input, err)
			continue
		}
		if got.MessageFilter != tt.wantMF {
			t.Errorf("%s: MessageFilter = 0x%04x, want 0x%04x", tt.name, got.MessageFilter, tt.wantMF)
		}
		if got.StreamFlags != tt.wantSF {
			t.Errorf("%s: StreamFlags = 0x%04x, want 0x%04x", tt.name, got.StreamFlags, tt.wantSF)
		}
		if len(got.ClientLevels) != len(tt.wantLvls) {
			t.Errorf("%s: ClientLevels = %v, want %v", tt.name, got.ClientLevels, tt.wantLvls)
			continue
		}
		for i := range got.ClientLevels {
			if got.ClientLevels[i] != tt.wantLvls[i] {
				t.Errorf("%s: ClientLevels[%d] = %v, want %v", tt.name, i, got.ClientLevels[i], tt.wantLvls[i])
			}
		}
	}
}

func TestClientFilter(t *testing.T) {
	entry := LogEntry{
		Level:   LogLevelError,
		Message: "connection timeout on endpoint /api/health",
		Label:   &LogLabel{Subsystem: "com.kwolf.testapp", Category: "networking"},
	}
	entryNoLabel := LogEntry{
		Level:   LogLevelDebug,
		Message: "something happened",
	}

	tests := []struct {
		name   string
		filter ClientFilter
		entry  LogEntry
		want   bool
	}{
		{"empty filter matches all", ClientFilter{}, entry, true},
		{"subsystem match", ClientFilter{Subsystem: "com.kwolf"}, entry, true},
		{"subsystem no match", ClientFilter{Subsystem: "com.apple"}, entry, false},
		{"subsystem filter on entry without label", ClientFilter{Subsystem: "com.kwolf"}, entryNoLabel, false},
		{"match hit", ClientFilter{Match: "timeout"}, entry, true},
		{"match miss", ClientFilter{Match: "success"}, entry, false},
		{"exclude hit", ClientFilter{Exclude: "timeout"}, entry, false},
		{"exclude miss", ClientFilter{Exclude: "success"}, entry, true},
		{"level hit", ClientFilter{Levels: []LogLevel{LogLevelError, LogLevelFault}}, entry, true},
		{"level miss", ClientFilter{Levels: []LogLevel{LogLevelDebug}}, entry, false},
		{"combined filters pass", ClientFilter{Subsystem: "com.kwolf", Match: "timeout"}, entry, true},
		{"combined filters fail on match", ClientFilter{Subsystem: "com.kwolf", Match: "success"}, entry, false},
		{"combined level + subsystem pass", ClientFilter{Levels: []LogLevel{LogLevelError}, Subsystem: "com.kwolf"}, entry, true},
		{"combined level + subsystem fail on level", ClientFilter{Levels: []LogLevel{LogLevelDebug}, Subsystem: "com.kwolf"}, entry, false},
	}

	for _, tt := range tests {
		if got := tt.filter.Matches(tt.entry); got != tt.want {
			t.Errorf("%s: Matches() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
