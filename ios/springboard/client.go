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
			screen[j] = t.icon
		}
		screens[i] = screen
	}

	return screens, nil
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

// Custom may be a widget or a paginated widget on the homescreen. We don't provide any information for this type
// of icon on the home screen
type Custom struct {
	IconType string `plist:"iconType"`
}

func (c Custom) DisplayName() string {
	// custom icons don't have a display name
	return ""
}
