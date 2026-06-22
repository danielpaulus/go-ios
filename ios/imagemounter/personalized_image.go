package imagemounter

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"howett.net/plist"
)

type buildManifest struct {
	ProductBuildVersion string `plist:"ProductBuildVersion"`
	BuildIdentities     []buildIdentity
}

func loadBuildManifest(p string) (buildManifest, error) {
	f, err := os.Open(p)
	if err != nil {
		return buildManifest{}, fmt.Errorf("loadBuildManifest: faild to open manifest file: %w", err)
	}
	defer f.Close()
	dec := plist.NewDecoder(f)
	var m buildManifest
	err = dec.Decode(&m)
	if err != nil {
		return buildManifest{}, fmt.Errorf("loadBuildManifest: could not decode manifest file: %w", err)
	}
	return m, nil
}

func (m buildManifest) findIdentity(identifiers personalizationIdentifiers) (buildIdentity, error) {
	for _, i := range m.BuildIdentities {
		if i.ApBoardID() == identifiers.BoardId && i.ApChipID() == identifiers.ChipID {
			return i, nil
		}
	}
	return buildIdentity{}, fmt.Errorf("findIdentity: failed to find identity for ApBoardId 0x%x and ApChipId 0x%x", identifiers.BoardId, identifiers.ChipID)
}

type restoreRequestRule struct {
	Actions    map[string]interface{} `plist:"Actions"`
	Conditions map[string]interface{} `plist:"Conditions"`
}

type manifestEntry struct {
	Digest  []byte
	Trusted bool `plist:"Trusted"`
	EPRO    bool `plist:"EPRO"`
	ESEC    bool `plist:"ESEC"`
	Name    string
	Info    struct {
		Path string
		// RestoreRequestRules dynamically determine fields such as EPRO/ESEC that must be
		// sent to Apple's TSS. iOS 17+ (notably iOS 26) manifests omit static EPRO/ESEC and
		// rely on these rules; without applying them TSS rejects the request with status 94.
		RestoreRequestRules []restoreRequestRule `plist:"RestoreRequestRules"`
	}
}

type buildIdentity struct {
	BoardID  string `plist:"ApBoardID"`
	ChipID   string `plist:"ApChipID"`
	Manifest map[string]manifestEntry
}

func (b buildIdentity) ApBoardID() int {
	return hexToInt(b.BoardID)
}

func (b buildIdentity) ApChipID() int {
	return hexToInt(b.ChipID)
}

func (b buildIdentity) dmgPath() string {
	if entry, ok := b.Manifest["PersonalizedDMG"]; ok {
		return entry.Info.Path
	}
	if entry, ok := b.Manifest["PersonalizedDmg"]; ok {
		return entry.Info.Path
	}
	return ""
}

func (b buildIdentity) trustCachePath() string {
	if entry, ok := b.Manifest["LoadableTrustCache"]; ok {
		return entry.Info.Path
	}
	return ""
}

type personalizationIdentifiers struct {
	BoardId               int
	ChipID                int
	SecurityDomain        int
	AdditionalIdentifiers map[string]interface{}
}

func hexToInt(s string) int {
	i, err := strconv.ParseInt(strings.ReplaceAll(strings.ToLower(s), "0x", ""), 16, 32)
	if err != nil {
		return 0
	}
	return int(i)
}
