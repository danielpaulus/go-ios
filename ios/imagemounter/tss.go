package imagemounter

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"howett.net/plist"
)

// tssClient is used to talk to https://gs.apple.com/TSS for getting the personalized developer disk image signatures
type tssClient struct {
	h *http.Client
}

func newTssClient() tssClient {
	c := &http.Client{
		Timeout:   1 * time.Minute,
		Transport: http.DefaultTransport,
	}

	return tssClient{
		h: c,
	}
}

func (t tssClient) getSignature(identity buildIdentity, identifiers personalizationIdentifiers, nonce []byte, ecid uint64) ([]byte, error) {
	params := map[string]interface{}{
		"@ApImg4Ticket":     true,
		"@BBTicket":         true,
		"@HostPlatformInfo": "mac",
		"@VersionInfo":      "libauthinstall-1104.0.9",
		"@UUID":             uuid.New().String(),
		"ApBoardID":         identifiers.BoardId,
		"ApChipID":          identifiers.ChipID,
		"ApECID":            ecid,
		"ApNonce":           nonce,
		"ApProductionMode":  true,
		"ApSecurityDomain":  identifiers.SecurityDomain,
		"ApSecurityMode":    true,
		"SepNonce":          make([]byte, 20),
		"UID_MODE":          false,
	}

	// Personalization parameters describing this device to the manifest's
	// RestoreRequestRules. Developer disk images are always personalized for a
	// production, secure, img4-based device.
	tssParameters := map[string]interface{}{
		"ApProductionMode": true,
		"ApSecurityMode":   true,
		"ApSupportsImg4":   true,
	}
	rules := identity.restoreRequestRules()

	for key, entry := range identity.Manifest {
		if !entry.Trusted {
			continue
		}
		entryParams := map[string]interface{}{
			"Digest":  entry.Digest,
			"Trusted": true,
		}
		// Keep any literal EPRO/ESEC the manifest provides, then let its
		// RestoreRequestRules override them for this device. Modern DDI manifests
		// omit the literals entirely; without applying the rules the request goes
		// out with EPRO/ESEC=false and Apple TSS rejects it with status 94
		// ("This device isn't eligible for the requested build.").
		if entry.EPRO != nil {
			entryParams["EPRO"] = *entry.EPRO
		}
		if entry.ESEC != nil {
			entryParams["ESEC"] = *entry.ESEC
		}
		applyRestoreRequestRules(entryParams, tssParameters, rules)

		if key == "PersonalizedDMG" || key == "PersonalizedDmg" {
			if entry.Name != "" {
				entryParams["Name"] = entry.Name
			} else {
				entryParams["Name"] = "DeveloperDiskImage"
			}
		}
		params[key] = entryParams
	}

	for k, v := range identifiers.AdditionalIdentifiers {
		params[k] = v
	}

	buf := bytes.NewBuffer(nil)
	enc := plist.NewEncoderForFormat(buf, plist.XMLFormat)
	err := enc.Encode(params)
	if err != nil {
		return nil, fmt.Errorf("getSignature: failed to encode request body: %w", err)
	}

	h := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
			Proxy: t.h.Transport.(*http.Transport).Proxy,
		},
		Timeout: 1 * time.Minute,
	}
	req, err := http.NewRequest("POST", "https://gs.apple.com/TSS/controller?action=2", buf)
	if err != nil {
		return nil, err
	}
	res, err := h.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getSignature: failed to send request: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusOK {
		resp, err := parseResponse(res.Body)
		if err != nil {
			return nil, fmt.Errorf("getSignature: failed to parse response: %w", err)
		}
		if resp.status != 0 {
			return nil, fmt.Errorf("unexpected status in response %d: %q", resp.status, resp.message)
		}
		var ticket map[string]interface{}
		_, err = plist.Unmarshal([]byte(resp.requestString), &ticket)
		if err != nil {
			return nil, fmt.Errorf("getSignature: failed to decode plist data: %w", err)
		}
		if ticket, ok := ticket["ApImg4Ticket"].([]byte); ok {
			return ticket, nil
		} else {
			return nil, fmt.Errorf("getSignature: could not get 'ApImg4Ticket' value from response")
		}
	}
	return nil, fmt.Errorf("getSignature: unexpected response status %d", res.StatusCode)
}

type response struct {
	status        int
	message       string
	requestString string
}

func parseResponse(r io.Reader) (response, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return response{}, fmt.Errorf("parseResponse: could not read content. %w", err)
	}
	s := string(b)
	end := func(s string) int {
		idx := strings.Index(s, "&")
		if idx < 0 {
			return len(s)
		} else {
			return idx
		}
	}

	var res response

	statusIdx := strings.Index(s, "STATUS=")
	if statusIdx >= 0 {
		statusStart := statusIdx + len("STATUS=")
		status := s[statusStart:]
		statusEnd := end(status)
		status = status[:statusEnd]
		stat, err := strconv.ParseInt(status, 10, 64)
		if err != nil {
			return response{}, fmt.Errorf("parseResponse: could not parse status '%s'. %w", status, err)
		}
		res.status = int(stat)
	}
	messageIdx := strings.Index(s, "MESSAGE=")
	if messageIdx >= 0 {
		messageStart := messageIdx + len("MESSAGE=")
		message := s[messageStart:]
		messageEnd := end(message)
		message = message[:messageEnd]
		res.message = message
	}

	requestStringIdx := strings.Index(s, "REQUEST_STRING=")
	if requestStringIdx >= 0 {
		if requestStringIdx <= messageIdx || requestStringIdx <= statusIdx {
			return response{}, fmt.Errorf("REQUEST_STRING value must come last")
		}
		requestStringStart := requestStringIdx + len("REQUEST_STRING=")
		requestString := s[requestStringStart:]
		requestStringEnd := end(requestString)
		requestString = requestString[:requestStringEnd]
		res.requestString = requestString
	}

	return res, nil
}
