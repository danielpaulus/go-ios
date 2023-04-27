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

// get locales and check if the specified locale works
// set locale, set lang
// set timezone
// TimeIntervalSince1970
func Prepare(device ios.DeviceEntry) error {
	locale := "de_DE"
	lang := "de"
	//tz := "Europe/Berlin"
	//time := ""

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

	setCloudConfig := map[string]interface{}{
		"CloudConfiguration": map[string]interface{}{
			"AllowPairing": 1,
			"SkipSetup":    skipAllSetup,
		},
		"RequestType": "SetCloudConfiguration",
	}
	setResp, err := check(conn.sendAndReceive(setCloudConfig))
	if err != nil {
		return fmt.Errorf("failed setting cloud config, resp: %v err: %v", setResp, err)
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
	return setupSkipSetup(device)
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
