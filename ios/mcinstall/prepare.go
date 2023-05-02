package mcinstall

import (
	"bytes"
	"fmt"
	"github.com/danielpaulus/go-ios/ios"
	"github.com/danielpaulus/go-ios/ios/afc"
	"github.com/danielpaulus/go-ios/ios/mobileactivation"
	log "github.com/sirupsen/logrus"
)

//removepath iTunes_Control/iTunes/SkipSetup, mkdir iTunes_Control/iTunes
//, fileopen iTunes_Control/iTunes/SkipSetup, write 1, file close

const skipSetupDirPath = "iTunes_Control/iTunes"
const skipSetupFilePath = "iTunes_Control/iTunes/SkipSetup"

var skipAllSetup = []string{"Location", "Restore", "SIMSetup", "Android", "AppleID", "Siri", "ScreenTime",
	"Diagnostics", "SoftwareUpdate", "Passcode", "Biometric", "Payment", "Zoom", "DisplayTone",
	"MessagingActivationUsingPhoneNumber", "HomeButtonSensitivity", "CloudStorage", "ScreenSaver",
	"TapToSetup", "Keyboard", "PreferredLanguage", "SpokenLanguage", "WatchMigration", "OnBoarding",
	"TVProviderSignIn", "TVHomeScreenSync", "Privacy", "TVRoom", "iMessageAndFaceTime", "AppStore",
	"Safety", "TermsOfAddress", "Welcome", "Appearance", "RestoreCompleted", "UpdateCompleted"}

// GetAllSetupSkipOptions returns a list of all possible values you can skip during device preparation
func GetAllSetupSkipOptions() []string {
	return skipAllSetup
}

// get locales and check if the specified locale works
// set locale, set lang
// set timezone
// TimeIntervalSince1970
func Prepare(device ios.DeviceEntry, skip []string, certBytes []byte, orgname string, locale string, lang string) error {
	if locale == "" {
		locale = "en_US"
	}
	if lang == "" {
		lang = "en"
	}

	//tz := "Europe/Berlin"
	//time := ""
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
	log.Infof("flush: %v", re)
	log.Info("get cloud config")
	config, err := check(conn.sendAndReceive(request("GetCloudConfiguration")))
	if err != nil {
		return err
	}
	log.Infof("config: %v", config)
	hello, err := check(conn.sendAndReceive(request("HelloHostIdentifier")))
	if err != nil {
		return err
	}
	log.Infof("hello: %v", hello)

	cloudConfig := map[string]interface{}{
		"AllowPairing": 1,
		"SkipSetup":    skip,
	}

	if supervise {
		cloudConfig["OrganizationName"] = orgname
		cloudConfig["SupervisorHostCertificates"] = [][]byte{certBytes}
		cloudConfig["IsSupervised"] = true
	}

	setCloudConfig := map[string]interface{}{
		"CloudConfiguration": cloudConfig,
		"RequestType":        "SetCloudConfiguration",
	}
	log.Warnf("set: %v", setCloudConfig)
	setResp, err := check(conn.sendAndReceive(setCloudConfig))
	if err != nil {
		return fmt.Errorf("failed setting cloud config, resp: %v err: %v", setResp, err)
	}
	hello, err = check(conn.sendAndReceive(request("HelloHostIdentifier")))
	if err != nil {
		return err
	}
	log.Info("get cloud config")
	config, err = check(conn.sendAndReceive(request("GetCloudConfiguration")))
	if err != nil {
		return err
	}
	log.Infof("config: %v", config)

	hello, err = check(conn.sendAndReceive(request("HelloHostIdentifier")))
	if err != nil {
		return err
	}
	err = conn.EscalateUnsupervised()
	if err != nil {
		log.Warn(err)
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

	log.Infof("%v", setResp)
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

	err = setupSkipSetup(device)
	if err != nil {
		log.Warn(err)
	}
	return err
}

/*
get cloud config when empty just gives: config: map[Status:Acknowledged]
*/
/*
map[CloudConfiguration:map[AllowPairing:true CloudConfigurationUIComplete:true ConfigurationSource:2 ConfigurationWasApplied:true IsMDMUnremovable:false IsSupervised:true OrganizationName:DanielsOrg PostSetupProfileWasInstalled:true SupervisorHostCertificates:
*/

func setupSkipSetup(device ios.DeviceEntry) error {
	afc, err := afc.New(device)
	if err != nil {
		return err
	}
	err = afc.RemovePathAndContents(skipSetupFilePath)
	if err != nil {
		log.Warn("nothing to remove")
	}
	err = afc.MkDir(skipSetupDirPath)
	if err != nil {
		log.Warn("error creating dir")
	}
	afc.WriteToFile(bytes.NewReader([]byte{}), skipSetupFilePath)
	f, _ := afc.ListFiles(skipSetupDirPath, "*")
	log.Info(f)
	return nil
}

/*
{"RequestType":"SetCloudConfiguration","CloudConfiguration":{"SkipSetup":["Location","Restore","SIMSetup","Android","AppleID","Siri","ScreenTime","Diagnostics","SoftwareUpdate","Passcode","Biometric","Payment","Zoom","DisplayTone","MessagingActivationUsingPhoneNumber","HomeButtonSensitivity","CloudStorage","ScreenSaver","TapToSetup","Keyboard","PreferredLanguage","SpokenLanguage","WatchMigration","OnBoarding","TVProviderSignIn","TVHomeScreenSync","Privacy","TVRoom","iMessageAndFaceTime","AppStore","Safety","TermsOfAddress","Welcome","Appearance","RestoreCompleted","UpdateCompleted"],"AllowPairing":true}}
*/

/* actresphead:

{"level":"info","msg":"actresphead:map[Cache-Control:[private, no-cache, no-store, must-revalidate, max-age=0] Connection:[keep-alive] Content-Length:[17515] Content-Type:[application/x-buddyml] Date:[Sun, 16 Apr 2023 11:41:38 GMT] Server:[Apple] Strict-Transport-Security:[max-age=31536000; includeSubdomains] X-Client-Request-Id:[728de7a5-f740-45bd-84e7-266fead5d751] X-Content-Type-Options:[nosniff] X-Frame-Options:[SAMEORIGIN] X-Xss-Protection:[1; mode=block]]","time":"2023-04-16T13:41:38+02:00"}
<xmlui style="setupAssistant"><page name="FMIPLockChallenge">
    <script>
    <![CDATA[
        function enableNext() {
            var username = xmlui.getFieldValue('login');
            var password = xmlui.getFieldValue('password');
            if(username && password) {
                return true;
            }
            if (!username && password) {
                password = password.replace(/-/g, "");
                if(password.length == 26) {
                    return true;
                }
            }
            return false;
        }

        function limitMaxLength(existingText, selectionLocation, selectionLength, newText) {
            var fullString = existingText.substring(0, selectionLocation) + newText + existingText.substring(selectionLocation + selectionLength);
            var maxLength = 1000;
            if (fullString.length > maxLength) {
                fullString = fullString.substring(0, maxLength);
            }
            return fullString;
        }

        function enableButton() {
            var passcode = xmlui.getFieldValue('passcode');
            if (passcode.length > 0) {
                return true;
            } else {
                return false;
            }
        }
    ]]>
    </script>
    <navigationBar title="Activation Lock" hidesBackButton="false" loadingTitle="Activating...">
        <linkBarItem id="next" url="/deviceservices/deviceActivation" position="right" label="Next" enabledFunction="enableNext" httpMethod="POST" />
    </navigationBar>
    <tableView>

    <section>
        <footer>This iPad is linked to an Apple ID. Enter the Apple ID and password that were used to set up this iPad. n●●●●●@yahoo.de</footer>
    </section>

    <section>
        <footer></footer>
    </section>

    <section>
        <editableTextRow id="login" label="Apple ID" keyboardType="email" firstResponder="true" disableAutocapitalization="true" disableAutocorrection="true"
         placeholder="example@icloud.com" changeCharactersFunction="limitMaxLength" value=""/>
        <editableTextRow id="password" label="Password" placeholder="Required" secure="true"/>
    </section>

    <section>
        <footer url="https://static.deviceservices.apple.com/deviceservices/buddy/barney_activation_help_en_us.buddyml">Activation Lock Help</footer>
    </section>

    </tableView>
</page>

<serverInfo activation-info-base64="PD94bWwgdm...
*/

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
