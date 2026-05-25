// internal/reasoning/gate/validator.go
// Validates that agent responses contain the mandatory Gate header.
// Used during golden testing and E2E test runs.
package gate

import (
	"fmt"
	"regexp"
	"strings"
)

// GateHeader represents a parsed Adaptive Reasoning Gate header
type GateHeader struct {
	Mode    int
	D1, D2, D3, D4, D5 int
	Posture1 string
	Posture2 string // empty if only one posture
	Raw      string
}

// headerPattern matches: [MODE N | D1=X D2=X D3=X D4=X D5=X | POSTURE: +++P1 [+++P2]]
var headerPattern = regexp.MustCompile(
	`^\[MODE ([123]) \| D1=([0-3]) D2=([0-3]) D3=([0-3]) D4=([0-3]) D5=([0-3]) \| POSTURE: (\+\+\+\w+)(?: (\+\+\+\w+))?\]`)

// ParseHeader parses the gate header from the first line of a response
func ParseHeader(firstLine string) (*GateHeader, error) {
	m := headerPattern.FindStringSubmatch(strings.TrimSpace(firstLine))
	if m == nil {
		return nil, fmt.Errorf("invalid or missing gate header: %q\nExpected format: [MODE N | D1=X D2=X D3=X D4=X D5=X | POSTURE: +++P]", firstLine)
	}

	parseInt := func(s string) int { var n int; fmt.Sscan(s, &n); return n }

	return &GateHeader{
		Mode:     parseInt(m[1]),
		D1:       parseInt(m[2]),
		D2:       parseInt(m[3]),
		D3:       parseInt(m[4]),
		D4:       parseInt(m[5]),
		D5:       parseInt(m[6]),
		Posture1: m[7],
		Posture2: m[8],
		Raw:      firstLine,
	}, nil
}

// ValidateDecision checks that the Mode and Postures are consistent with D-scores
func ValidateDecision(h *GateHeader) []string {
	var issues []string

	// D5 >= 2 must have +++Adversarial
	if h.D5 >= 2 {
		hasAdversarial := h.Posture1 == "+++Adversarial" || h.Posture2 == "+++Adversarial"
		if !hasAdversarial {
			issues = append(issues, fmt.Sprintf("D5=%d requires +++Adversarial posture", h.D5))
		}
		if h.Mode < 2 {
			issues = append(issues, fmt.Sprintf("D5=%d requires Mode >= 2, got Mode %d", h.D5, h.Mode))
		}
	}

	// D3 >= 2 must be Mode 3
	if h.D3 >= 2 && h.Mode < 3 {
		issues = append(issues, fmt.Sprintf("D3=%d requires Mode 3, got Mode %d", h.D3, h.Mode))
	}

	// Two postures must not be the same
	if h.Posture2 != "" && h.Posture1 == h.Posture2 {
		issues = append(issues, fmt.Sprintf("duplicate postures: %s and %s", h.Posture1, h.Posture2))
	}

	// Divergent/Lateral/Diamond only valid in ideator context (warning, not error)
	creativePostures := map[string]bool{"+++Divergent": true, "+++Lateral": true, "+++Diamond": true}
	if creativePostures[h.Posture1] && h.Mode == 3 {
		issues = append(issues, "creative postures (Divergent/Lateral/Diamond) are unusual for Mode 3")
	}

	return issues
}

// ExtractFirstLine returns the first non-empty line of a response
func ExtractFirstLine(response string) string {
	for _, line := range strings.Split(response, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}
