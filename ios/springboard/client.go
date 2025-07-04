package springboard

import (
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
)

type Client struct {
	connection ios.DeviceConnectionInterface
	plistCodec ios.PlistCodecReadWriter
}

func NewClient(d ios.DeviceEntry) (*Client, error) {
	conn, err := ios.ConnectToService(d, "com.apple.springboardservices")
	if err != nil {
		return nil, fmt.Errorf("could not connect to 'com.apple.configurator.xpc.DeviceService': %w", err)
	}
	return &Client{
		connection: conn,
		plistCodec: ios.NewPlistCodecReadWriter(conn.Reader(), conn.Writer()),
	}, nil
}

func (c *Client) Close() error {
	return c.connection.Close()
}

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

type Screen []Icon

type Icon interface {
	DisplayName() string
}

type Folder struct {
	Name     string `plist:"displayName"`
	Icons    [][]Icon
	ListType string `plist:"listType"`
}

func (f Folder) DisplayName() string {
	return f.Name
}

type AppIcon struct {
	Name              string `plist:"displayName"`
	DisplayIdentifier string `plist:"displayIdentifier"`
	BundleId          string `plist:"bundleIdentifier"`
	BundleVersion     string `plist:"bundleVersion"`
}

func (a AppIcon) DisplayName() string {
	return a.Name
}

type WebClip struct {
	Name              string `plist:"displayName"`
	DisplayIdentifier string `plist:"displayIdentifier"`
	URL               string `plist:"webClipURL"`
}

func (w WebClip) DisplayName() string {
	return w.Name
}

type Custom struct {
	IconType string `plist:"iconType"`
}

func (c Custom) DisplayName() string {
	// custom icons don't have a display name
	return ""
}
