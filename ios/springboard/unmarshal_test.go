package springboard

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/danielpaulus/go-ios/ios"
	"howett.net/plist"
)

// iconStateClient returns a Client whose codec answers one getIconState reply built from
// state, so every case runs through the same decode path as a device.
func iconStateClient(t *testing.T, state any) *Client {
	t.Helper()
	payload, err := plist.Marshal(state, plist.XMLFormat)
	if err != nil {
		t.Fatalf("could not marshal test state: %v", err)
	}
	msg := make([]byte, 4, 4+len(payload))
	binary.BigEndian.PutUint32(msg, uint32(len(payload)))
	msg = append(msg, payload...)
	return &Client{plistCodec: ios.NewPlistCodecReadWriter(bytes.NewReader(msg), &bytes.Buffer{})}
}

func appEntry(bundleID, name string) map[string]any {
	return map[string]any{
		"bundleIdentifier":  bundleID,
		"displayIdentifier": bundleID,
		"displayName":       name,
		"bundleVersion":     "1.0",
	}
}

func webclipEntry(name, url string) map[string]any {
	return map[string]any{
		"displayIdentifier": "17269AEB545E4DED9954EAAAA75C91D7",
		"displayName":       name,
		"webClipURL":        url,
		"iconModDate":       "2026-10-05T11:26:55Z",
	}
}

// placeholderEntry is what iOS reports for a built-in app whose download was removed
// (offloaded) or that was taken off the home screen while staying installed.
func placeholderEntry(identifier, name string) map[string]any {
	return map[string]any{"displayIdentifier": identifier, "displayName": name}
}

func folderEntry(name string, pages ...[]any) map[string]any {
	return map[string]any{"listType": "folder", "displayName": name, "iconLists": pages}
}

func TestListIcons(t *testing.T) {
	tests := []struct {
		name     string
		state    []any
		wantErr  string
		exactErr bool
		check    func(t *testing.T, screens []Screen)
	}{
		{
			name:  "app icon",
			state: []any{[]any{appEntry("com.example.app", "App")}},
			check: func(t *testing.T, screens []Screen) {
				app, ok := screens[0][0].(AppIcon)
				if !ok {
					t.Fatalf("expected AppIcon, got %T", screens[0][0])
				}
				if app.BundleId != "com.example.app" || app.Name != "App" {
					t.Errorf("unexpected app contents: %+v", app)
				}
			},
		},
		{
			name:  "webclip",
			state: []any{[]any{webclipEntry("Example", "https://example.com")}},
			check: func(t *testing.T, screens []Screen) {
				clip, ok := screens[0][0].(WebClip)
				if !ok {
					t.Fatalf("expected WebClip, got %T", screens[0][0])
				}
				if clip.URL != "https://example.com" || clip.Name != "Example" {
					t.Errorf("unexpected webclip contents: %+v", clip)
				}
			},
		},
		{
			name:  "custom icon",
			state: []any{[]any{map[string]any{"iconType": "custom"}}},
			check: func(t *testing.T, screens []Screen) {
				if _, ok := screens[0][0].(Custom); !ok {
					t.Fatalf("expected Custom, got %T", screens[0][0])
				}
			},
		},
		{
			name:  "folder with children",
			state: []any{[]any{folderEntry("Utils", []any{appEntry("com.example.app", "App"), webclipEntry("Example", "https://example.com")})}},
			check: func(t *testing.T, screens []Screen) {
				folder, ok := screens[0][0].(Folder)
				if !ok {
					t.Fatalf("expected Folder, got %T", screens[0][0])
				}
				if folder.Name != "Utils" || len(folder.Icons) != 1 || len(folder.Icons[0]) != 2 {
					t.Fatalf("unexpected folder contents: %+v", folder)
				}
				if _, ok := folder.Icons[0][0].(AppIcon); !ok {
					t.Errorf("expected AppIcon child, got %T", folder.Icons[0][0])
				}
				if _, ok := folder.Icons[0][1].(WebClip); !ok {
					t.Errorf("expected WebClip child, got %T", folder.Icons[0][1])
				}
			},
		},
		{
			name: "offloaded built-in keeps the layout decodable",
			state: []any{[]any{
				appEntry("com.apple.mobilephone", "Telefon"),
				placeholderEntry("com.apple.Music", "Musik"),
				webclipEntry("Example", "https://example.com"),
			}},
			check: func(t *testing.T, screens []Screen) {
				placeholder, ok := screens[0][1].(PlaceholderIcon)
				if !ok {
					t.Fatalf("expected PlaceholderIcon, got %T", screens[0][1])
				}
				if placeholder.DisplayIdentifier != "com.apple.Music" || placeholder.Name != "Musik" {
					t.Errorf("unexpected placeholder contents: %+v", placeholder)
				}
				if _, ok := screens[0][2].(WebClip); !ok {
					t.Errorf("expected WebClip after the placeholder, got %T", screens[0][2])
				}
			},
		},
		{
			name: "offloaded built-in inside a folder",
			state: []any{[]any{folderEntry("Utils", []any{
				placeholderEntry("com.apple.shortcuts", "Shortcuts"),
				appEntry("com.example.app", "App"),
			})}},
			check: func(t *testing.T, screens []Screen) {
				folder, ok := screens[0][0].(Folder)
				if !ok {
					t.Fatalf("expected Folder, got %T", screens[0][0])
				}
				if _, ok := folder.Icons[0][0].(PlaceholderIcon); !ok {
					t.Errorf("expected PlaceholderIcon child, got %T", folder.Icons[0][0])
				}
				if _, ok := folder.Icons[0][1].(AppIcon); !ok {
					t.Errorf("expected AppIcon child, got %T", folder.Icons[0][1])
				}
			},
		},
		{
			name: "mixed layout lets consumers enumerate webclips",
			state: []any{
				[]any{appEntry("com.apple.mobilephone", "Telefon"), webclipEntry("Example", "https://example.com")},
				[]any{
					placeholderEntry("com.apple.Music", "Musik"),
					folderEntry("Utils", []any{placeholderEntry("com.apple.shortcuts", "Shortcuts"), webclipEntry("Other", "https://other.org")}),
					map[string]any{"iconType": "custom"},
				},
			},
			check: func(t *testing.T, screens []Screen) {
				clips := 0
				for _, screen := range screens {
					for _, icon := range screen {
						if _, ok := icon.(WebClip); ok {
							clips++
						}
						if folder, ok := icon.(Folder); ok {
							for _, page := range folder.Icons {
								for _, child := range page {
									if _, ok := child.(WebClip); ok {
										clips++
									}
								}
							}
						}
					}
				}
				if clips != 2 {
					t.Errorf("expected 2 webclips across screens and folders, got %d", clips)
				}
			},
		},
		{
			name: "unknown dict fails with keys and position",
			state: []any{
				[]any{appEntry("com.apple.mobilephone", "Telefon")},
				[]any{appEntry("com.example.one", "One"), appEntry("com.example.two", "Two"),
					map[string]any{"betaTitle": "SECRET-BETA-VALUE", "isRemovable": true}},
			},
			wantErr:  "could not unmarshal icon at page 2 position 3 with keys [betaTitle isRemovable]",
			exactErr: true,
		},
		{
			name:     "empty bundleIdentifier is not an app",
			state:    []any{[]any{map[string]any{"bundleIdentifier": "", "displayName": "X"}}},
			wantErr:  "could not unmarshal icon at page 1 position 1 with keys [bundleIdentifier displayName]",
			exactErr: true,
		},
		{
			name:     "empty displayIdentifier is not a placeholder",
			state:    []any{[]any{map[string]any{"displayIdentifier": "", "displayName": "X"}}},
			wantErr:  "could not unmarshal icon at page 1 position 1 with keys [displayIdentifier displayName]",
			exactErr: true,
		},
		{
			name: "unknown dict inside a folder fails with the folder location",
			state: []any{[]any{
				appEntry("com.apple.mobilephone", "Telefon"),
				folderEntry("Utils", []any{appEntry("com.example.app", "App")}, []any{map[string]any{"foo": "SECRET-FOLDER-VALUE"}}),
			}},
			wantErr:  "could not unmarshal icon at folder page 2 position 1 inside folder at page 1 position 2 with keys [foo]",
			exactErr: true,
		},
		{
			name:    "non-dict entry keeps failing with the plist type mismatch",
			state:   []any{[]any{"placeholder"}},
			wantErr: "type mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := iconStateClient(t, tt.state)
			screens, err := client.ListIcons()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got none", tt.wantErr)
				}
				// exact matches pin that dict values never leak into the error
				if tt.exactErr && err.Error() != tt.wantErr {
					t.Fatalf("error %q does not equal %q", err.Error(), tt.wantErr)
				}
				if !tt.exactErr && !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.check != nil {
				tt.check(t, screens)
			}
		})
	}
}
