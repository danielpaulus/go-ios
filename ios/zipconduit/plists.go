package zipconduit

import (
	"fmt"
	"path"
)

//metadata is used to write to a plist file that we have to add to what we send
type metadata struct {
	StandardDirectoryPerms int
	StandardFilePerms      int
	RecordCount            int
	TotalUncompressedBytes uint64
	Version                int
}

//initTransfer is the request you have to send initially to start the transfer
type initTransfer struct {
	InstallOptionsDictionary    installoptions
	InstallTransferredDirectory int
	MediaSubdir                 string
	UserInitiatedTransfer       int
}

//installOptions contains some settings, we just use what XCode uses
type installoptions struct {
	DisableDeltaTransfer int
	InstallDeltaTypeKey  string
	IsUserInitiated      int
	PackageType          string
	PreferWifi           int
}

const signingError = "ApplicationVerificationFailed"

func evaluateProgress(progressUpdate map[string]interface{}) (bool, int, string, error) {
	//done, percent, status
	statusIntf, ok := progressUpdate["Status"]
	if ok {
		status := statusIntf.(string)
		if "DataComplete" == status {
			return true, 100, status, nil
		}
		return false, 0, "", fmt.Errorf("invalid progressUpdate, unknown Status field:+%+v", progressUpdate)
	}

	installProgressDictIntf, ok := progressUpdate["InstallProgressDict"]
	if !ok {
		return false, 0, "", fmt.Errorf("invalid progressUpdate, missing InstallProgressDict field:+%+v", progressUpdate)
	}
	installProgressDict := installProgressDictIntf.(map[string]interface{})

	errorMessage, ok := installProgressDict["Error"]
	if ok {
		description, _ := installProgressDict["ErrorDescription"]
		if signingError == errorMessage {
			return false, 0, "", fmt.Errorf("your app is not properly signed for this device, check your codesigning and provisioningprofile. original error: '%s' errorDescription:'%s'", errorMessage, description)
		}
		return false, 0, "", fmt.Errorf("failed installing: '%s' errorDescription:'%s'", errorMessage, description)
	}

	percentIntf, ok := installProgressDict["PercentComplete"]
	if !ok {
		return false, 0, "", fmt.Errorf("invalid installProgressDict, missing PercentComplete field:+%+v", progressUpdate)
	}
	percent := int(percentIntf.(uint64))

	statusIntf, ok = installProgressDict["Status"]
	if !ok {
		return false, 0, "", fmt.Errorf("invalid installProgressDict, missing Status field:+%+v", progressUpdate)
	}
	status := statusIntf.(string)
	return false, percent, status, nil
}

// newInitTransfer returns a initTransfer request with
// the same values XCode uses
func newInitTransfer(fileName string) initTransfer {
	base := path.Base(fileName)
	return initTransfer{
		InstallTransferredDirectory: 1,
		UserInitiatedTransfer:       0,
		MediaSubdir:                 fmt.Sprintf("PublicStaging/%s", base),
		InstallOptionsDictionary: installoptions{
			InstallDeltaTypeKey:  "InstallDeltaTypeSparseIPAFiles",
			DisableDeltaTransfer: 1,
			IsUserInitiated:      1,
			PreferWifi:           1,
			PackageType:          "Customer",
		},
	}
}
