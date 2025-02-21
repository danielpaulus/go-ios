package mcinstall

import (
	"bytes"
	"fmt"

	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/afc"
	"github.com/danielpaulus/go-ios/ios/mobileactivation"
	log "github.com/sirupsen/logrus"
)

const (
	skipSetupDirPath  = "iTunes_Control/iTunes"
	skipSetupFilePath = "iTunes_Control/iTunes/SkipSetup"
)

// Those keys can be found in the official Apple documentation
// here - https://developer.apple.com/documentation/devicemanagement/skipkeys
var skipAllSetup = []string{
	"Accessibility",
	"ActionButton",
	"Android",
	"Appearance",
	"AppleID",
	"AppStore",
	"Biometric",
	"CameraButton",
	"DeviceToDeviceMigration",
	"Diagnostics",
	"EnableLockdownMode",
	"FileVault",
	"iCloudDiagnostics",
	"iCloudStorage",
	"iMessageAndFaceTime",
	"Intelligence",
	"Keyboard",
	"Location",
	"MessagingActivationUsingPhoneNumber",
	"Passcode",
	"Payment",
	"Privacy",
	"Restore",
	"RestoreCompleted",
	"Safety",
	"ScreenSaver",
	"ScreenTime",
	"SIMSetup",
	"Siri",
	"SoftwareUpdate",
	"SpokenLanguage",
	"TapToSetup",
	"TermsOfAddress",
	"TOS",
	"TVHomeScreenSync",
	"TVProviderSignIn",
	"TVRoom",
	"UpdateCompleted",
	"WatchMigration",
	"Welcome",
	// Deprecated keys
	// Deprecated in iOS 15
	"DisplayTone",
	// Deprecated in iOS 15, was used only on iPhone 7, 7 Plus, 8, 8 Plus and SE
	"HomeButtonSensitivity",
	// Deprecated in iOS 14
	"OnBoarding",
	// Deprecated in iOS 17
	"Zoom",
}

// GetAllSetupSkipOptions returns a list of all possible values you can skip during device preparation
func GetAllSetupSkipOptions() []string {
	return skipAllSetup
}

// Prepare prepares an activated device and supervises it if desired. skip is the list of setup options to skip, use GetAllSetupSkipOptions()
// to get a list of all available options. certBytes is the DER encoded supervision certificate. If it is nil then the device won't be supervised.
// ios.CreateDERFormattedSupervisionCert() provides an example how to generate these certificates. Orgname can be any string, it will show up as the
// supervision name on the device. Locale and lang can be set. If they are empty strings, then the default will be en_US and en.
func Prepare(device ios.DeviceEntry, skip []string, certBytes []byte, orgname string, locale string, lang string) error {
	if locale == "" {
		locale = "en_US"
	}
	if lang == "" {
		lang = "en"
	}

	supervise := certBytes != nil
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
	log.Info("get cloud config")
	config, err := check(conn.sendAndReceive(request("GetCloudConfiguration")))
	if err != nil {
		return err
	}
	log.Debugf("get first cloudconfig: %v", config)
	hello, err := check(conn.sendAndReceive(request("HelloHostIdentifier")))
	if err != nil {
		return err
	}
	log.Debugf("hello response: %v", hello)

	cloudConfig := map[string]interface{}{
		"AllowPairing": 1,
		"SkipSetup":    skip,
	}

	if supervise {
		log.Info("supervising device")
		cloudConfig["OrganizationName"] = orgname
		cloudConfig["SupervisorHostCertificates"] = [][]byte{certBytes}
		cloudConfig["IsSupervised"] = true
	}

	setCloudConfig := map[string]interface{}{
		"CloudConfiguration": cloudConfig,
		"RequestType":        "SetCloudConfiguration",
	}
	log.Debugf("set cloud config: %v", setCloudConfig)
	setResp, err := check(conn.sendAndReceive(setCloudConfig))
	if err != nil {
		return fmt.Errorf("failed setting cloud config, resp: %v err: %v", setResp, err)
	}
	log.Debugf("set response: %v", setResp)
	hello, err = check(conn.sendAndReceive(request("HelloHostIdentifier")))
	if err != nil {
		return err
	}
	log.Debug("get cloud config")
	config, err = check(conn.sendAndReceive(request("GetCloudConfiguration")))
	if err != nil {
		return err
	}
	log.Debugf("cloud config config: %v", config)

	hello, err = check(conn.sendAndReceive(request("HelloHostIdentifier")))
	if err != nil {
		return err
	}
	err = conn.EscalateUnsupervised()
	if err != nil {
		// the device always throws a CertificateRejected error here, but it works just fine
		log.Debug(err)
	}
	hello, err = check(conn.sendAndReceive(request("HelloHostIdentifier")))
	if err != nil {
		return err
	}
	err = conn.AddProfile([]byte(initialProfile))
	if err != nil {
		return err
	}
	hello, err = check(conn.sendAndReceive(request("HelloHostIdentifier")))
	if err != nil {
		return err
	}

	err = ios.SetLanguage(device, ios.LanguageConfiguration{Language: lang, Locale: locale})
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	err = ios.SetSystemTime(device)
	if err != nil {
		return err
	}

	return setupSkipSetup(device)
}

func setupSkipSetup(device ios.DeviceEntry) error {
	afcConn, err := afc.New(device)
	if err != nil {
		return err
	}
	err = afcConn.RemovePathAndContents(skipSetupFilePath)
	if err != nil {
		log.Debug("skip setup: nothing to remove")
	}
	err = afcConn.MkDir(skipSetupDirPath)
	if err != nil {
		log.Warn("error creating dir")
	}
	err = afcConn.WriteToFile(bytes.NewReader([]byte{}), skipSetupFilePath)
	if err != nil {
		return err
	}
	if log.GetLevel() == log.DebugLevel {
		f, _ := afcConn.ListFiles(skipSetupDirPath, "*")
		log.Debugf("list of files %v", f)
	}
	return nil
}

const initialProfile = `
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>PayloadContent</key>
	<array>
		<dict>
			<key>PayloadDescription</key>
			<string>Configures Restrictions</string>
			<key>PayloadDisplayName</key>
			<string>Restrictions</string>
			<key>PayloadIdentifier</key>
			<string>com.apple.applicationaccess.8D384EDB-E60E-4D14-8E07-98ED164374D3</string>
			<key>PayloadType</key>
			<string>com.apple.applicationaccess</string>
			<key>PayloadUUID</key>
			<string>889918E1-B5AA-4AD6-B819-240081988CAB</string>
			<key>PayloadVersion</key>
			<integer>1</integer>
			<key>allowActivityContinuation</key>
			<true/>
			<key>allowAddingGameCenterFriends</key>
			<true/>
			<key>allowAirPlayIncomingRequests</key>
			<true/>
			<key>allowAirPrint</key>
			<true/>
			<key>allowAirPrintCredentialsStorage</key>
			<true/>
			<key>allowAirPrintiBeaconDiscovery</key>
			<true/>
			<key>allowAppCellularDataModification</key>
			<true/>
			<key>allowAppClips</key>
			<true/>
			<key>allowAppInstallation</key>
			<true/>
			<key>allowAppRemoval</key>
			<true/>
			<key>allowApplePersonalizedAdvertising</key>
			<true/>
			<key>allowAssistant</key>
			<true/>
			<key>allowAssistantWhileLocked</key>
			<true/>
			<key>allowAutoCorrection</key>
			<true/>
			<key>allowAutoUnlock</key>
			<true/>
			<key>allowAutomaticAppDownloads</key>
			<true/>
			<key>allowBluetoothModification</key>
			<true/>
			<key>allowBookstore</key>
			<true/>
			<key>allowBookstoreErotica</key>
			<true/>
			<key>allowCamera</key>
			<true/>
			<key>allowCellularPlanModification</key>
			<true/>
			<key>allowChat</key>
			<true/>
			<key>allowCloudBackup</key>
			<true/>
			<key>allowCloudDocumentSync</key>
			<true/>
			<key>allowCloudPhotoLibrary</key>
			<true/>
			<key>allowContinuousPathKeyboard</key>
			<true/>
			<key>allowDefinitionLookup</key>
			<true/>
			<key>allowDeviceNameModification</key>
			<true/>
			<key>allowDeviceSleep</key>
			<true/>
			<key>allowDictation</key>
			<true/>
			<key>allowESIMModification</key>
			<true/>
			<key>allowEnablingRestrictions</key>
			<true/>
			<key>allowEnterpriseAppTrust</key>
			<true/>
			<key>allowEnterpriseBookBackup</key>
			<true/>
			<key>allowEnterpriseBookMetadataSync</key>
			<true/>
			<key>allowEraseContentAndSettings</key>
			<true/>
			<key>allowExplicitContent</key>
			<true/>
			<key>allowFilesNetworkDriveAccess</key>
			<true/>
			<key>allowFilesUSBDriveAccess</key>
			<true/>
			<key>allowFindMyDevice</key>
			<true/>
			<key>allowFindMyFriends</key>
			<true/>
			<key>allowFingerprintForUnlock</key>
			<true/>
			<key>allowFingerprintModification</key>
			<true/>
			<key>allowGameCenter</key>
			<true/>
			<key>allowGlobalBackgroundFetchWhenRoaming</key>
			<true/>
			<key>allowInAppPurchases</key>
			<true/>
			<key>allowKeyboardShortcuts</key>
			<true/>
			<key>allowManagedAppsCloudSync</key>
			<true/>
			<key>allowMultiplayerGaming</key>
			<true/>
			<key>allowMusicService</key>
			<true/>
			<key>allowNews</key>
			<true/>
			<key>allowNotificationsModification</key>
			<true/>
			<key>allowOpenFromManagedToUnmanaged</key>
			<true/>
			<key>allowOpenFromUnmanagedToManaged</key>
			<true/>
			<key>allowPairedWatch</key>
			<true/>
			<key>allowPassbookWhileLocked</key>
			<true/>
			<key>allowPasscodeModification</key>
			<true/>
			<key>allowPasswordAutoFill</key>
			<true/>
			<key>allowPasswordProximityRequests</key>
			<true/>
			<key>allowPasswordSharing</key>
			<true/>
			<key>allowPersonalHotspotModification</key>
			<true/>
			<key>allowPhotoStream</key>
			<true/>
			<key>allowPredictiveKeyboard</key>
			<true/>
			<key>allowProximitySetupToNewDevice</key>
			<true/>
			<key>allowRadioService</key>
			<true/>
			<key>allowRemoteAppPairing</key>
			<true/>
			<key>allowRemoteScreenObservation</key>
			<true/>
			<key>allowSafari</key>
			<true/>
			<key>allowScreenShot</key>
			<true/>
			<key>allowSharedStream</key>
			<true/>
			<key>allowSpellCheck</key>
			<true/>
			<key>allowSpotlightInternetResults</key>
			<true/>
			<key>allowSystemAppRemoval</key>
			<true/>
			<key>allowUIAppInstallation</key>
			<true/>
			<key>allowUIConfigurationProfileInstallation</key>
			<true/>
			<key>allowUSBRestrictedMode</key>
			<false/>
			<key>allowUnpairedExternalBootToRecovery</key>
			<false/>
			<key>allowUntrustedTLSPrompt</key>
			<true/>
			<key>allowVPNCreation</key>
			<true/>
			<key>allowVideoConferencing</key>
			<true/>
			<key>allowVoiceDialing</key>
			<true/>
			<key>allowWallpaperModification</key>
			<true/>
			<key>allowiTunes</key>
			<true/>
			<key>forceAirDropUnmanaged</key>
			<false/>
			<key>forceAirPrintTrustedTLSRequirement</key>
			<false/>
			<key>forceAssistantProfanityFilter</key>
			<false/>
			<key>forceAuthenticationBeforeAutoFill</key>
			<false/>
			<key>forceAutomaticDateAndTime</key>
			<false/>
			<key>forceClassroomAutomaticallyJoinClasses</key>
			<false/>
			<key>forceClassroomRequestPermissionToLeaveClasses</key>
			<false/>
			<key>forceClassroomUnpromptedAppAndDeviceLock</key>
			<false/>
			<key>forceClassroomUnpromptedScreenObservation</key>
			<false/>
			<key>forceDelayedSoftwareUpdates</key>
			<false/>
			<key>forceEncryptedBackup</key>
			<false/>
			<key>forceITunesStorePasswordEntry</key>
			<false/>
			<key>forceLimitAdTracking</key>
			<false/>
			<key>forceWatchWristDetection</key>
			<false/>
			<key>forceWiFiPowerOn</key>
			<false/>
			<key>forceWiFiWhitelisting</key>
			<false/>
			<key>ratingApps</key>
			<integer>1000</integer>
			<key>ratingMovies</key>
			<integer>1000</integer>
			<key>ratingTVShows</key>
			<integer>1000</integer>
			<key>safariAcceptCookies</key>
			<integer>2</integer>
			<key>safariAllowAutoFill</key>
			<true/>
			<key>safariAllowJavaScript</key>
			<true/>
			<key>safariAllowPopups</key>
			<true/>
			<key>safariForceFraudWarning</key>
			<false/>
		</dict>
	</array>
	<key>PayloadDescription</key>
	<string>USB Accessories while locked allowed</string>
	<key>PayloadDisplayName</key>
	<string>USB Accessories while locked allowed</string>
	<key>PayloadIdentifier</key>
	<string>com.apple.configurator.usbrestrictedmode</string>
	<key>PayloadRemovalDisallowed</key>
	<false/>
	<key>PayloadType</key>
	<string>Configuration</string>
	<key>PayloadUUID</key>
	<string>7E1B12F3-FCF1-4120-83A0-26E950519103</string>
	<key>PayloadVersion</key>
	<integer>1</integer>
</dict>
</plist>
`
