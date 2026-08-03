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

type manifestEntry struct {
	Digest  []byte
	Trusted bool  `plist:"Trusted"`
	EPRO    *bool `plist:"EPRO"`
	ESEC    *bool `plist:"ESEC"`
	Name    string
	Info    struct {
		Path                string
		RestoreRequestRules []restoreRequestRule `plist:"RestoreRequestRules"`
	}
}

// restoreRequestRule gates the EPRO/ESEC flags of a manifest component on the
// device's personalization parameters (production mode, security mode, …).
// Modern developer-disk-image build manifests omit literal EPRO/ESEC values and
// express them through these rules instead.
type restoreRequestRule struct {
	Actions    map[string]interface{} `plist:"Actions"`
	Conditions map[string]interface{} `plist:"Conditions"`
}

// restoreRequestRules returns the rules that determine the EPRO/ESEC flags for
// the TSS personalization request. Apple's libauthinstall (and pymobiledevice3)
// take these from the LoadableTrustCache component and apply them to every
// trusted component of a developer disk image.
func (b buildIdentity) restoreRequestRules() []restoreRequestRule {
	if entry, ok := b.Manifest["LoadableTrustCache"]; ok {
		return entry.Info.RestoreRequestRules
	}
	return nil
}

// applyRestoreRequestRules mutates a TSS manifest entry, applying the actions of
// every rule whose conditions are satisfied by params. This is how EPRO/ESEC end
// up on the request: e.g. a device in production + secure mode gets EPRO=true and
// ESEC=true even though the build manifest carries no literal values. Mirrors the
// rule evaluation in Apple's libauthinstall / pymobiledevice3.
func applyRestoreRequestRules(entry, params map[string]interface{}, rules []restoreRequestRule) {
	for _, rule := range rules {
		if !restoreRuleConditionsMet(rule.Conditions, params) {
			continue
		}
		for key, value := range rule.Actions {
			// 255 is a "leave unchanged" sentinel used by some components.
			if v, ok := value.(uint64); ok && v == 255 {
				continue
			}
			entry[key] = value
		}
	}
}

func restoreRuleConditionsMet(conditions, params map[string]interface{}) bool {
	for key, want := range conditions {
		var got interface{}
		switch key {
		case "ApRawProductionMode", "ApCurrentProductionMode":
			got = params["ApProductionMode"]
		case "ApRawSecurityMode":
			got = params["ApSecurityMode"]
		case "ApRequiresImage4":
			got = params["ApSupportsImg4"]
		case "ApDemotionPolicyOverride":
			got = params["DemotionPolicy"]
		case "ApInRomDFU":
			got = params["ApInRomDFU"]
		default:
			// Unknown condition: treat the rule as unmatched, matching the
			// reference implementations rather than guessing.
			return false
		}
		if got == nil || want != got {
			return false
		}
	}
	return true
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
