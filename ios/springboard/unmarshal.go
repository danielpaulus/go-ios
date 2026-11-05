package springboard

import (
	"slices"
)

// internalUnmarshaler is a helper struct that knows how to unmarshal a plist that represents an Icon.
type internalUnmarshaler struct {
	icon Icon
}

// UnmarshalPlist will try to unmarshall the different types of Icons we have and populate the icon field
// with the detected type of Icon
func (t *internalUnmarshaler) UnmarshalPlist(unmarshal func(interface{}) error) error {
	var app AppIcon
	err := unmarshal(&app)
	if err != nil {
		return err
	}
	if len(app.BundleId) > 0 {
		t.icon = app
		return nil
	}

	var webclip WebClip
	err = unmarshal(&webclip)
	if err != nil {
		return err
	}
	if len(webclip.URL) > 0 {
		t.icon = webclip
		return nil
	}

	var custom Custom
	err = unmarshal(&custom)
	if err != nil {
		return err
	}
	if custom.IconType == "custom" {
		t.icon = custom
		return nil
	}

	var list Folder
	err = unmarshal(&list)
	if err != nil {
		return err
	}
	if list.ListType == "folder" {
		var f folderUnmarshaler
		err := unmarshal(&f)
		if err != nil {
			return err
		}
		list.Icons = make([][]Icon, len(f.Icons))
		for i, icons := range f.Icons {
			list.Icons[i] = make([]Icon, len(icons))
			for j, icon := range icons {
				list.Icons[i][j] = icon.icon
			}
		}
		t.icon = list
		return nil
	}

	// iOS reports built-in apps that were offloaded or taken off the home screen with only
	// a displayIdentifier and displayName while the app stays installed. Such entries carry
	// no actionable data for icon automation but must not break the whole layout decode.
	if len(app.DisplayIdentifier) > 0 {
		t.icon = PlaceholderIcon{Name: app.Name, DisplayIdentifier: app.DisplayIdentifier}
		return nil
	}

	// The plist decoder cannot tell us where an entry sits, so an entry matching no known
	// shape is stored here and ListIcons fails on it with the entry's page and position.
	var keys map[string]any
	if err := unmarshal(&keys); err != nil {
		return err
	}
	names := make([]string, 0, len(keys))
	for k := range keys {
		names = append(names, k)
	}
	slices.Sort(names)
	t.icon = unknownIcon{keys: names}
	return nil
}

// unknownIcon marks a dict that matched no known icon shape. It never reaches consumers:
// ListIcons rejects it with an error naming its keys and position.
type unknownIcon struct {
	keys []string
}

func (u unknownIcon) DisplayName() string {
	return ""
}

type folderUnmarshaler struct {
	Icons [][]internalUnmarshaler `plist:"iconLists"`
}
