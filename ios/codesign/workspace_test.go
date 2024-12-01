package codesign_test

import (
	"os"
	"testing"

	"github.com/danielpaulus/go-ios/ios/codesign"
	log "github.com/sirupsen/logrus"
)

func TestWorkspaceInit(t *testing.T) {
	_, _, cleanUp := makeWorkspaceWithoutProfiles()
	defer cleanUp()

}

func makeWorkspaceWithoutProfiles() (codesign.SigningWorkspace, string, func()) {
	dir, err := os.MkdirTemp("", "codesign-test")
	if err != nil {
		log.Fatal(err)
	}

	workspace := codesign.NewSigningWorkspace(dir)

	cleanUp := func() {
		defer os.RemoveAll(dir)
	}
	return workspace, dir, cleanUp
}
