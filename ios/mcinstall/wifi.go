package mcinstall

import (
	"crypto/sha1"
	"fmt"
	"regexp"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/golog"
	"github.com/danielpaulus/go-ios/ios/mobileactivation"
)

func PrepareWifi(device ios.DeviceEntry, ssid string, psw string, encType string) error {
	if ssid == "" || psw == "" {
		return fmt.Errorf("PrepareWifi: both ssid and password must be specified")
	}
	isActivated, err := mobileactivation.IsActivated(device)
	if err != nil {
		return fmt.Errorf("PrepareWifi: %w", err)
	}
	if !isActivated {
		return fmt.Errorf("PrepareWifi: device is not activated, activate it first")
	}
	golog.Info("device is activated", "module", logModule, "udid", device.Properties.SerialNumber, "activated", isActivated)

	conn, err := New(device)
	if err != nil {
		return fmt.Errorf("PrepareWifi: %w", err)
	}
	defer conn.Close()
	golog.Info("send flush request", "module", logModule, "udid", device.Properties.SerialNumber)
	re, err := check(conn.sendAndReceive(request("Flush")))
	if err != nil {
		return fmt.Errorf("PrepareWifi: flush failed: %w", err)
	}
	golog.Debug("flush response", "module", logModule, "udid", device.Properties.SerialNumber, "response", re)

	supervised := isSupervised(conn)

	err = conn.EscalateUnsupervised()
	if err != nil {
		// the device always throws a CertificateRejected error here, but it works just fine
		golog.Debug("ignoring expected CertificateRejected error", "module", logModule, "udid", device.Properties.SerialNumber, "error", err)
	}

	safeSSID := sanitizeIdentifier(ssid)
	profileId := fmt.Sprintf("com.apple.wifi.managed.%s", safeSSID)

	// SSID/password are user-controlled, so XML-escape them before they go into
	// the plist; the two PayloadUUIDs are derived from the SSID so each network
	// gets a distinct, stable profile (re-installing the same SSID is idempotent).
	profile := fmt.Sprintf(wifiProfileTemplate,
		xmlEscape(encType),
		xmlEscape(ssid),
		xmlEscape(psw),
		profileId+".payload",
		profileUUID(ssid+".payload"),
		profileId,
		profileUUID(ssid),
	)

	if err := conn.AddProfile([]byte(profile)); err != nil {
		return fmt.Errorf("PrepareWifi: %w", err)
	}

	if !supervised {
		golog.Warn("device is not supervised: the Wi-Fi profile was installed but must be approved manually on the device (Settings > General > VPN & Device Management) before it takes effect. Supervise the device for a silent install.",
			"module", logModule, "udid", device.Properties.SerialNumber, "ssid", ssid)
	}
	return nil
}

func RemoveWifi(device ios.DeviceEntry, ssid string) error {
	if ssid == "" {
		return fmt.Errorf("RemoveWifi: ssid must be specified")
	}
	isActivated, err := mobileactivation.IsActivated(device)
	if err != nil {
		return fmt.Errorf("RemoveWifi: %w", err)
	}
	if !isActivated {
		return fmt.Errorf("RemoveWifi: device is not activated, activate it first")
	}
	golog.Info("device is activated", "module", logModule, "udid", device.Properties.SerialNumber, "activated", isActivated)

	conn, err := New(device)
	if err != nil {
		return fmt.Errorf("RemoveWifi: %w", err)
	}
	defer conn.Close()
	golog.Info("send flush request", "module", logModule, "udid", device.Properties.SerialNumber)
	re, err := check(conn.sendAndReceive(request("Flush")))
	if err != nil {
		return fmt.Errorf("RemoveWifi: flush failed: %w", err)
	}
	golog.Debug("flush response", "module", logModule, "udid", device.Properties.SerialNumber, "response", re)

	supervised := isSupervised(conn)

	err = conn.EscalateUnsupervised()
	if err != nil {
		// the device always throws a CertificateRejected error here, but it works just fine
		golog.Debug("ignoring expected CertificateRejected error", "module", logModule, "udid", device.Properties.SerialNumber, "error", err)
	}

	safeSSID := sanitizeIdentifier(ssid)
	profileId := fmt.Sprintf("com.apple.wifi.managed.%s", safeSSID)

	if err := conn.RemoveProfile(profileId); err != nil {
		return fmt.Errorf("RemoveWifi: %w", err)
	}

	if !supervised {
		golog.Warn("device is not supervised: profile removal may require manual confirmation on the device (Settings > General > VPN & Device Management).",
			"module", logModule, "udid", device.Properties.SerialNumber, "ssid", ssid)
	}
	return nil
}

// isSupervised reports whether the device is supervised (managed). A non-supervised
// device can still receive profiles, but they must be approved manually on-device.
// On any error reading the cloud configuration it conservatively returns false.
func isSupervised(conn *Connection) bool {
	cfg, err := conn.GetCloudConfiguration()
	if err != nil {
		return false
	}
	cloudCfg, ok := cfg["CloudConfiguration"].(map[string]interface{})
	if !ok {
		return false
	}
	s, _ := cloudCfg["IsSupervised"].(bool)
	return s
}

// sanitizeIdentifier converts a string into an Apple-compatible PayloadIdentifier fragment.
func sanitizeIdentifier(input string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sanitized := reg.ReplaceAllString(input, "-")
	return strings.Trim(sanitized, "-")
}

// xmlEscape escapes the five XML predefined entities so user-provided values
// (SSID, password) can't break the plist or inject markup.
func xmlEscape(s string) string {
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		"\"", "&quot;",
		"'", "&apos;",
	).Replace(s)
}

// profileUUID derives a stable UUID-formatted string from seed so each SSID gets
// its own PayloadUUID (Apple expects these to be unique across profiles).
func profileUUID(seed string) string {
	h := sha1.Sum([]byte("go-ios.wifi." + seed))
	return strings.ToUpper(fmt.Sprintf("%x-%x-%x-%x-%x", h[0:4], h[4:6], h[6:8], h[8:10], h[10:16]))
}

const wifiProfileTemplate = `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>PayloadContent</key>
    <array>
        <dict>
            <key>AutoJoin</key>
            <true/>
            <key>EncryptionType</key>
            <string>%s</string>
            <key>HIDDEN_NETWORK</key>
            <false/>
            <key>SSID_STR</key>
            <string>%s</string>
            <key>Password</key>
            <string>%s</string>
            <key>PayloadDisplayName</key>
            <string>Wi-Fi</string>
            <key>PayloadIdentifier</key>
            <string>%s</string>
            <key>PayloadType</key>
            <string>com.apple.wifi.managed</string>
            <key>PayloadUUID</key>
            <string>%s</string>
            <key>PayloadVersion</key>
            <integer>1</integer>
        </dict>
    </array>
    <key>PayloadIdentifier</key>
    <string>%s</string>
    <key>PayloadType</key>
    <string>Configuration</string>
    <key>PayloadUUID</key>
    <string>%s</string>
    <key>PayloadVersion</key>
    <integer>1</integer>
</dict>
</plist>
`
