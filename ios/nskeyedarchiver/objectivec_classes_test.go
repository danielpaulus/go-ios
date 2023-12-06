package nskeyedarchiver

import (
	"encoding/hex"
	"github.com/stretchr/testify/require"
	"log"
	"testing"
)

func TestXctCaps(t *testing.T) {
	caps := XCTCapabilities{}
	bin, err := ArchiveBin(caps)
	require.NoError(t, err)
	log.Printf("%s", hex.EncodeToString(bin))
}
