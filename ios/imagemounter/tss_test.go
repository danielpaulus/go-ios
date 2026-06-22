package imagemounter

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  response
	}{
		{
			name:  "success without request message",
			input: "STATUS=0&MESSAGE=SUCCESS",
			want: response{
				status:  0,
				message: "SUCCESS",
			},
		},
		{
			name:  "response with multiword status",
			input: "STATUS=69&MESSAGE=This device isn't eligible for the requested build.",
			want: response{
				status:  69,
				message: "This device isn't eligible for the requested build.",
			},
		},
		{
			name:  "response with request string",
			input: "STATUS=0&MESSAGE=SUCCESS&REQUEST_STRING=<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>",
			want: response{
				status:        0,
				message:       "SUCCESS",
				requestString: "<?xml version=\"1.0\" encoding=\"UTF-8\" standalone=\"yes\"?>",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			r, err := parseResponse(reader)
			require.NoError(t, err)
			assert.Equal(t, tt.want, r)
		})
	}
}

// TODO: It looks like `REQUEST_STRING` always comes last, but if that's not the case we are not sure what to do
// as it could also contain the '&' separator character
func TestParseResponseRequiresRequestStringLast(t *testing.T) {
	_, err := parseResponse(strings.NewReader("REQUEST_STRING=abc&STATUS=0"))
	assert.Error(t, err)
}

// iOS 17+ (e.g. iOS 26) BuildManifests omit static EPRO/ESEC and instead provide
// RestoreRequestRules that must be evaluated against the request parameters. These rules
// mirror the LoadableTrustCache.Info.RestoreRequestRules found on a real iOS 26 device.
func ios26RestoreRequestRules() []restoreRequestRule {
	return []restoreRequestRule{
		{
			Actions:    map[string]interface{}{"EPRO": false},
			Conditions: map[string]interface{}{"ApCurrentProductionMode": false, "ApRequiresImage4": true},
		},
		{
			Actions:    map[string]interface{}{"EPRO": true},
			Conditions: map[string]interface{}{"ApCurrentProductionMode": true, "ApRequiresImage4": true},
		},
		{
			Actions:    map[string]interface{}{"EPRO": false},
			Conditions: map[string]interface{}{"ApCurrentProductionMode": true, "ApDemotionPolicyOverride": "Demote", "ApInRomDFU": true, "ApRequiresImage4": true},
		},
		{
			Actions:    map[string]interface{}{"ESEC": false},
			Conditions: map[string]interface{}{"ApRawSecurityMode": false, "ApRequiresImage4": true},
		},
		{
			Actions:    map[string]interface{}{"ESEC": true},
			Conditions: map[string]interface{}{"ApRawSecurityMode": true, "ApRequiresImage4": true},
		},
	}
}

func TestApplyRestoreRequestRulesProductionDevice(t *testing.T) {
	// Production + secure + Img4 device: EPRO and ESEC must both resolve to true,
	// otherwise Apple's TSS rejects the personalization request with status 94.
	params := map[string]interface{}{
		"ApProductionMode": true,
		"ApSecurityMode":   true,
		"ApSupportsImg4":   true,
	}
	entry := map[string]interface{}{"Digest": []byte{0x01}, "Trusted": true}

	applyRestoreRequestRules(entry, params, ios26RestoreRequestRules())

	assert.Equal(t, true, entry["EPRO"])
	assert.Equal(t, true, entry["ESEC"])
	// The DFU-only EPRO=false rule must not fire (its ApInRomDFU/ApDemotionPolicyOverride
	// conditions reference parameters that are absent).
}

func TestApplyRestoreRequestRulesNonProductionDevice(t *testing.T) {
	params := map[string]interface{}{
		"ApProductionMode": false,
		"ApSecurityMode":   false,
		"ApSupportsImg4":   true,
	}
	entry := map[string]interface{}{"Digest": []byte{0x01}, "Trusted": true}

	applyRestoreRequestRules(entry, params, ios26RestoreRequestRules())

	assert.Equal(t, false, entry["EPRO"])
	assert.Equal(t, false, entry["ESEC"])
}

func TestApplyRestoreRequestRulesSkipsUnknownCondition(t *testing.T) {
	params := map[string]interface{}{"ApProductionMode": true}
	entry := map[string]interface{}{}
	rules := []restoreRequestRule{
		{
			Actions:    map[string]interface{}{"EPRO": true},
			Conditions: map[string]interface{}{"SomeUnmappedCondition": true},
		},
	}

	applyRestoreRequestRules(entry, params, rules)

	_, ok := entry["EPRO"]
	assert.False(t, ok, "rule with an unmapped condition key must be skipped")
}
