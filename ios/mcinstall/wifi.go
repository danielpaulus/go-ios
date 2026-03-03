package mcinstall

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/mobileactivation"
	log "github.com/sirupsen/logrus"
)

func PrepareWifi(device ios.DeviceEntry, ssid string, psw string, encType string) error {

	isActivated, err := mobileactivation.IsActivated(device)
	if err != nil {
		return err
	}
	if !isActivated {
		return fmt.Errorf("please activate the device first")
	}
	log.Infof("device is activated:%v", isActivated)

	conn, err := New(device)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Info("send flush request")
	re, err := check(conn.sendAndReceive(request("Flush")))
	if err != nil {
		return err
	}
	log.Debugf("flush: %v", re)

	err = conn.EscalateUnsupervised()
	if err != nil {
		// the device always throws a CertificateRejected error here, but it works just fine
		log.Debug(err)
	}

	safeSSID := sanitizeIdentifier(ssid)
	profileId := fmt.Sprintf("com.apple.wifi.managed.%s", safeSSID)

	profile := fmt.Sprintf(wifiProfileTemplate,
		encType,
		ssid,
		psw,
		profileId+".payload",
		profileId,
	)

	err = conn.AddProfile([]byte(profile))
	if err != nil {
		return err
	}

	return nil
}

func RemoveWifi(device ios.DeviceEntry, ssid string) error {

	isActivated, err := mobileactivation.IsActivated(device)
	if err != nil {
		return err
	}
	if !isActivated {
		return fmt.Errorf("please activate the device first")
	}
	log.Infof("device is activated:%v", isActivated)

	conn, err := New(device)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Info("send flush request")
	re, err := check(conn.sendAndReceive(request("Flush")))
	if err != nil {
		return err
	}
	log.Debugf("flush: %v", re)

	err = conn.EscalateUnsupervised()
	if err != nil {
		// the device always throws a CertificateRejected error here, but it works just fine
		log.Debug(err)
	}

	safeSSID := sanitizeIdentifier(ssid)
	profileId := fmt.Sprintf("com.apple.wifi.managed.%s", safeSSID)

	err = conn.RemoveProfile(profileId)
	if err != nil {
		return err
	}

	return nil
}

// Convert a string into a PayloadIdentifier Apple compatible string
func sanitizeIdentifier(input string) string {
	reg := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	sanitized := reg.ReplaceAllString(input, "-")
	return strings.Trim(sanitized, "-")
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
            <string>A1B2C3D4-E5F6-4A5B-8C9D-0E1F2A3B4C5D</string>
            <key>PayloadVersion</key>
            <integer>1</integer>
        </dict>
    </array>
    <key>PayloadIdentifier</key>
    <string>%s</string>
    <key>PayloadType</key>
    <string>Configuration</string>
    <key>PayloadUUID</key>
    <string>7E1B12F3-FCF1-4120-83A0-26E950519103</string>
    <key>PayloadVersion</key>
    <integer>1</integer>
</dict>
</plist>
`
