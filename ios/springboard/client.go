package springboard

import (
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
)

// Client is a connection to the `com.apple.springboardservices` service on the device
type Client struct {
	connection ios.DeviceConnectionInterface
	plistCodec ios.PlistCodecReadWriter
}

func NewClient(d ios.DeviceEntry) (*Client, error) {
	conn, err := ios.ConnectToService(d, "com.apple.springboardservices")
	if err != nil {
		return nil, fmt.Errorf("could not connect to 'com.apple.springboardservices': %w", err)
	}
	return &Client{
		connection: conn,
		plistCodec: ios.NewPlistCodecReadWriter(conn.Reader(), conn.Writer()),
	}, nil
}

func (c *Client) Close() error {
	return c.connection.Close()
}

// GetHomeScreenWallpaperPNG returns the home screen wallpaper as PNG bytes.
// This does not require supervision. iOS does not currently expose the lock
// screen wallpaper through this service.
func (c *Client) GetHomeScreenWallpaperPNG() ([]byte, error) {
	err := c.plistCodec.Write(map[string]any{
		"command": "getHomeScreenWallpaperPNGData",
	})
	if err != nil {
		return nil, fmt.Errorf("could not write plist: %w", err)
	}
	var response struct {
		PNGData []byte `plist:"pngData"`
	}
	if err := c.plistCodec.Read(&response); err != nil {
		return nil, fmt.Errorf("could not read plist: %w", err)
	}
	if len(response.PNGData) == 0 {
		return nil, fmt.Errorf("device returned empty pngData")
	}
	return response.PNGData, nil
}

// GetIconLayout returns the home screen icon layout in its raw plist shape so
// it can be saved as JSON, edited, and pushed back via SetIconLayout. Format
// version "2" matches what cfgutil and pymobiledevice3 use; pass an empty
// string for the legacy v1 layout.
func (c *Client) GetIconLayout(formatVersion string) (any, error) {
	cmd := map[string]any{"command": "getIconState"}
	if formatVersion != "" {
		cmd["formatVersion"] = formatVersion
	}
	if err := c.plistCodec.Write(cmd); err != nil {
		return nil, fmt.Errorf("could not write plist: %w", err)
	}
	var response any
	if err := c.plistCodec.Read(&response); err != nil {
		return nil, fmt.Errorf("could not read plist: %w", err)
	}
	return response, nil
}

// SetIconLayout pushes a previously-fetched icon layout back to the device.
// The state should be the value returned by GetIconLayout.
func (c *Client) SetIconLayout(state any) error {
	if state == nil {
		state = map[string]any{}
	}
	if err := c.plistCodec.Write(map[string]any{
		"command":   "setIconState",
		"iconState": state,
	}); err != nil {
		return fmt.Errorf("could not write plist: %w", err)
	}
	return nil
}

// ListIcons provides the homescreen layout of the device
func (c *Client) ListIcons() ([]Screen, error) {
	err := c.plistCodec.Write(map[string]any{
		"command":       "getIconState",
		"formatVersion": "2",
	})
	if err != nil {
		return nil, fmt.Errorf("could not write plist: %w", err)
	}
	var response [][]internalUnmarshaler
	err = c.plistCodec.Read(&response)
	if err != nil {
		return nil, fmt.Errorf("could not read plist: %w", err)
	}

	screens := make([]Screen, len(response))
	for i, s := range response {
		screen := make([]Icon, len(s))
		for j, t := range s {
			err := checkIcon(t.icon, fmt.Sprintf("page %d position %d", i+1, j+1))
			if err != nil {
				return nil, err
			}
			screen[j] = t.icon
		}
		screens[i] = screen
	}

	return screens, nil
}

// checkIcon rejects icons that matched no known shape. Positions are 1-based, matching how
// the pages appear on the device.
func checkIcon(icon Icon, location string) error {
	switch v := icon.(type) {
	case unknownIcon:
		return fmt.Errorf("could not unmarshal icon at %s with keys %v", location, v.keys)
	case Folder:
		for i, page := range v.Icons {
			for j, child := range page {
				err := checkIcon(child, fmt.Sprintf("folder page %d position %d inside folder at %s", i+1, j+1, location))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// Screen is a list of Icons displayed on one page of the home screen
// the first entry is always the bar on the bottom of the home screen (also if it is empty)
type Screen []Icon

type Icon interface {
	DisplayName() string
}

// Folder is a collection of items in a folder. Folders can also have multiple pages
type Folder struct {
	Name     string `plist:"displayName"`
	Icons    [][]Icon
	ListType string `plist:"listType"`
}

func (f Folder) DisplayName() string {
	return f.Name
}

// AppIcon represent a native app
type AppIcon struct {
	Name              string `plist:"displayName"`
	DisplayIdentifier string `plist:"displayIdentifier"`
	BundleId          string `plist:"bundleIdentifier"`
	BundleVersion     string `plist:"bundleVersion"`
}

func (a AppIcon) DisplayName() string {
	return a.Name
}

// WebClip represent a Safari bookmark, or a progressive-web-app
type WebClip struct {
	Name              string `plist:"displayName"`
	DisplayIdentifier string `plist:"displayIdentifier"`
	URL               string `plist:"webClipURL"`
}

func (w WebClip) DisplayName() string {
	return w.Name
}

// PlaceholderIcon is the home screen entry for a built-in app whose download was removed
// (offloaded) or that was taken off the home screen while staying installed. iOS reports
// it with only a displayIdentifier and displayName, so it matches no other icon type.
type PlaceholderIcon struct {
	Name              string `plist:"displayName"`
	DisplayIdentifier string `plist:"displayIdentifier"`
}

func (p PlaceholderIcon) DisplayName() string {
	return p.Name
}

// Custom may be a widget or a paginated widget on the homescreen. We don't provide any information for this type
// of icon on the home screen
type Custom struct {
	IconType string `plist:"iconType"`
}

func (c Custom) DisplayName() string {
	// custom icons don't have a display name
	return ""
}
