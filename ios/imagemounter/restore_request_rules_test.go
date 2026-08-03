package imagemounter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func boolPtr(b bool) *bool { return &b }

// ddiTrustCacheRules mirrors the RestoreRequestRules that modern developer-disk-image
// build manifests attach to the LoadableTrustCache component. EPRO/ESEC are not literal
// values in the manifest; they are derived from these rules and the device's parameters.
func ddiTrustCacheRules() []restoreRequestRule {
	return []restoreRequestRule{
		{Actions: map[string]interface{}{"EPRO": false}, Conditions: map[string]interface{}{"ApCurrentProductionMode": false, "ApRequiresImage4": true}},
		{Actions: map[string]interface{}{"EPRO": true}, Conditions: map[string]interface{}{"ApCurrentProductionMode": true, "ApRequiresImage4": true}},
		{Actions: map[string]interface{}{"ESEC": false}, Conditions: map[string]interface{}{"ApRawSecurityMode": false, "ApRequiresImage4": true}},
		{Actions: map[string]interface{}{"ESEC": true}, Conditions: map[string]interface{}{"ApRawSecurityMode": true, "ApRequiresImage4": true}},
	}
}

// A production, secure device must end up with EPRO=true and ESEC=true, otherwise
// Apple TSS rejects the request with status 94 ("This device isn't eligible for the
// requested build.").
func TestApplyRestoreRequestRulesProductionSecure(t *testing.T) {
	params := map[string]interface{}{
		"ApProductionMode": true,
		"ApSecurityMode":   true,
		"ApSupportsImg4":   true,
	}
	entry := map[string]interface{}{"Digest": []byte{1, 2, 3}, "Trusted": true}

	applyRestoreRequestRules(entry, params, ddiTrustCacheRules())

	assert.Equal(t, true, entry["EPRO"])
	assert.Equal(t, true, entry["ESEC"])
}

// A non-production, non-secure device (e.g. a demoted / development unit) must get
// EPRO=false and ESEC=false from the same rules.
func TestApplyRestoreRequestRulesDevelopment(t *testing.T) {
	params := map[string]interface{}{
		"ApProductionMode": false,
		"ApSecurityMode":   false,
		"ApSupportsImg4":   true,
	}
	entry := map[string]interface{}{}

	applyRestoreRequestRules(entry, params, ddiTrustCacheRules())

	assert.Equal(t, false, entry["EPRO"])
	assert.Equal(t, false, entry["ESEC"])
}

// A rule referencing a condition we do not model (or whose parameter is absent) must
// not match, leaving the entry untouched rather than applying its actions.
func TestApplyRestoreRequestRulesUnmatchedConditions(t *testing.T) {
	params := map[string]interface{}{"ApProductionMode": true, "ApSupportsImg4": true}
	entry := map[string]interface{}{}

	rules := []restoreRequestRule{
		// ApInRomDFU is not set in params -> rule must not apply.
		{Actions: map[string]interface{}{"EPRO": false}, Conditions: map[string]interface{}{"ApInRomDFU": true}},
		// Unknown condition -> rule must not apply.
		{Actions: map[string]interface{}{"ESEC": true}, Conditions: map[string]interface{}{"SomeUnknownCondition": true}},
	}

	applyRestoreRequestRules(entry, params, rules)

	assert.NotContains(t, entry, "EPRO")
	assert.NotContains(t, entry, "ESEC")
}

// The 255 sentinel means "leave unchanged" and must not overwrite an existing value.
func TestApplyRestoreRequestRulesSentinel(t *testing.T) {
	params := map[string]interface{}{"ApProductionMode": true, "ApSupportsImg4": true}
	entry := map[string]interface{}{"EPRO": true}

	rules := []restoreRequestRule{
		{Actions: map[string]interface{}{"EPRO": uint64(255)}, Conditions: map[string]interface{}{"ApCurrentProductionMode": true, "ApRequiresImage4": true}},
	}

	applyRestoreRequestRules(entry, params, rules)

	assert.Equal(t, true, entry["EPRO"])
}
