package mcinstall

import (
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
)

// RemoveProxy unsets the global HTTP proxy config again by deleting the global config profile
// installed by go-ios using the identifier
// I hardcoded 'Go-iOS.CD15976B-E205-4213-9B8E-FDAA5FAB1C22'
func RemoveProxy(device ios.DeviceEntry) error {
	profileService, err := New(device)
	if err != nil {
		return err
	}
	defer profileService.Close()
	return profileService.RemoveProfile("Go-iOS.CD15976B-E205-4213-9B8E-FDAA5FAB1C22")
}

// SetHttpProxy generates the config profile "Go-iOS.CD15976B-E205-4213-9B8E-FDAA5FAB1C22" that will set a global
// http proxy on supervised devices.
func SetHttpProxy(device ios.DeviceEntry, host string, port string, user string, pass string, p12file []byte, p12password string) error {
	profileBytes, err := setUpProfile(host, port, user, pass)
	if err != nil {
		return err
	}
	return InstallProfileSilent(device, p12file, p12password, profileBytes)
}

func setUpProfile(host string, port string, user string, pass string) ([]byte, error) {
	if host == "" || port == "" {
		return []byte{}, fmt.Errorf("host and port must not be empty")
	}
	if user != "" {
		profile := fmt.Sprintf(profileTemplateAuth, host, port, user, pass)
		return []byte(profile), nil
	}
	profile := fmt.Sprintf(profileTemplate, host, port)
	return []byte(profile), nil
}

// InstallProfileSilent install a configuration profile silently.
func InstallProfileSilent(device ios.DeviceEntry, p12file []byte, p12password string, profileBytes []byte) error {
	profileService, err := New(device)
	if err != nil {
		return err
	}
	defer profileService.Close()
	return profileService.AddProfileSupervised(profileBytes, p12file, p12password)
}

const profileTemplate = `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadDescription</key>
			<string>Global HTTP Proxy</string>
			<key>PayloadDisplayName</key>
			<string>Global HTTP Proxy</string>
			<key>PayloadIdentifier</key>
			<string>com.apple.proxy.http.global.20A1B29D-7945-4C7C-9A49-649D3751F85D</string>
			<key>PayloadType</key>
			<string>com.apple.proxy.http.global</string>
			<key>PayloadUUID</key>
			<string>20A1B29D-7945-4C7C-9A49-649D3751F85D</string>
			<key>PayloadVersion</key>
			<integer>1</integer>
			<key>ProxyCaptiveLoginAllowed</key>
			<true/>
			<key>ProxyServer</key>
			<string>%s</string>
			<key>ProxyServerPort</key>
			<integer>%s</integer>
			<key>ProxyType</key>
			<string>Manual</string>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Untitled</string>
	<key>PayloadIdentifier</key>
	<string>Go-iOS.CD15976B-E205-4213-9B8E-FDAA5FAB1C22</string>
	<key>PayloadRemovalDisallowed</key>
	<false/>
	<key>PayloadType</key>
	<string>Configuration</string>
	<key>PayloadUUID</key>
	<string>20A1B29D-7945-4C7C-9A49-649D3751F85D</string>
	<key>PayloadVersion</key>
	<integer>1</integer>
</dict>
</plist>
`

const profileTemplateAuth = `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadDescription</key>
			<string>Global HTTP Proxy</string>
			<key>PayloadDisplayName</key>
			<string>Global HTTP Proxy</string>
			<key>PayloadIdentifier</key>
			<string>com.apple.proxy.http.global.20A1B29D-7945-4C7C-9A49-649D3751F85D</string>
			<key>PayloadType</key>
			<string>com.apple.proxy.http.global</string>
			<key>PayloadUUID</key>
			<string>20A1B29D-7945-4C7C-9A49-649D3751F85D</string>
			<key>PayloadVersion</key>
			<integer>1</integer>
			<key>ProxyCaptiveLoginAllowed</key>
			<true/>
			<key>ProxyServer</key>
			<string>%s</string>
			<key>ProxyServerPort</key>
			<integer>%s</integer>
			<key>ProxyType</key>
			<string>Manual</string>
			<key>ProxyUsername</key>
			<string>%s</string>
			<key>ProxyPassword</key>
			<string>%s</string>
		</dict>
	</array>
	<key>PayloadDisplayName</key>
	<string>Untitled</string>
	<key>PayloadIdentifier</key>
	<string>Go-iOS.CD15976B-E205-4213-9B8E-FDAA5FAB1C22</string>
	<key>PayloadRemovalDisallowed</key>
	<false/>
	<key>PayloadType</key>
	<string>Configuration</string>
	<key>PayloadUUID</key>
	<string>20A1B29D-7945-4C7C-9A49-649D3751F85D</string>
	<key>PayloadVersion</key>
	<integer>1</integer>
</dict>
</plist>
`
