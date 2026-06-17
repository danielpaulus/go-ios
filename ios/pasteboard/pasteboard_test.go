package pasteboard

import "testing"

func TestTextItem(t *testing.T) {
	item := TextItem("héllo 🌍")
	if len(item.Types) != 3 {
		t.Fatalf("expected 3 UTIs, got %d", len(item.Types))
	}
	for _, uti := range textUTIs {
		got, ok := item.Data[uti]
		if !ok {
			t.Fatalf("missing data for UTI %q", uti)
		}
		if string(got) != "héllo 🌍" {
			t.Fatalf("UTI %q: expected UTF-8 bytes of input, got %q", uti, got)
		}
	}
}

func TestDataItem(t *testing.T) {
	raw := []byte{0x01, 0x02, 0x03}
	item := DataItem(UTIURL, raw)
	if len(item.Types) != 1 || item.Types[0] != UTIURL {
		t.Fatalf("unexpected types: %v", item.Types)
	}
	if string(item.Data[UTIURL]) != string(raw) {
		t.Fatalf("unexpected data: %v", item.Data[UTIURL])
	}
}

func TestSnapshotTextFromPullReply(t *testing.T) {
	reply := map[string]interface{}{
		"command": "PULL_REPLY",
		"pasteboard": map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{
					"types": []interface{}{UTIUTF8PlainText},
					"data": map[string]interface{}{
						UTIUTF8PlainText: map[string]interface{}{"data": []byte("clipboard text")},
					},
				},
			},
		},
	}
	text, ok := snapshotText(reply)
	if !ok {
		t.Fatal("expected text to be extracted")
	}
	if text != "clipboard text" {
		t.Fatalf("unexpected text: %q", text)
	}
}

func TestSnapshotTextPrefersUTF8Order(t *testing.T) {
	reply := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"data": map[string]interface{}{
					UTIText:          map[string]interface{}{"data": []byte("plain")},
					UTIUTF8PlainText: map[string]interface{}{"data": []byte("utf8")},
				},
			},
		},
	}
	text, ok := snapshotText(reply)
	if !ok {
		t.Fatal("expected text to be extracted")
	}
	if text != "utf8" {
		t.Fatalf("expected UTF-8 UTI to win, got %q", text)
	}
}

func TestSnapshotTextPromisedItemHasNoInlineData(t *testing.T) {
	reply := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"data": map[string]interface{}{
					UTIUTF8PlainText: map[string]interface{}{
						"isPromised":  true,
						"isAvailable": false,
						"size":        int64(42),
					},
				},
			},
		},
	}
	if _, ok := snapshotText(reply); ok {
		t.Fatal("expected no text for a promised-only item")
	}
}

func TestSnapshotTextEmpty(t *testing.T) {
	if _, ok := snapshotText(map[string]interface{}{}); ok {
		t.Fatal("expected no text from empty reply")
	}
}
