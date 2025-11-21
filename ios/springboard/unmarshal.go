package springboard

import "fmt"

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

	return fmt.Errorf("could not unmarshal plist")
}

type folderUnmarshaler struct {
	Icons [][]internalUnmarshaler `plist:"iconLists"`
}
