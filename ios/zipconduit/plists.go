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
