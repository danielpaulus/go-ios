package imagemounter

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"howett.net/plist"
)

type buildManifest struct {
	BuildIdentities []buildIdentity
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

type buildIdentity struct {
	BoardID  string `plist:"ApBoardID"`
	ChipID   string `plist:"ApChipID"`
	Manifest struct {
		LoadableTrustCache struct {
			Digest []byte
			Info   struct {
				Path string
			}
		}
		PersonalizedDmg struct {
			Digest []byte
			Info   struct {
				Path string
			}
		} `plist:"PersonalizedDMG"`
	}
}

func (b buildIdentity) ApBoardID() int {
	return hexToInt(b.BoardID)
}

func (b buildIdentity) ApChipID() int {
	return hexToInt(b.ChipID)
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
