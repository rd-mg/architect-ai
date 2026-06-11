// internal/reasoning/gate/validator.go
// Validates that agent responses contain the mandatory Adaptive Reasoning Gate header.
// Used during golden testing, E2E test runs, and orchestrator dispatch validation.
package gate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// GateHeader represents a parsed Adaptive Reasoning Gate header.
// Postures are NOT declared in the header — they are derived by Score().
type GateHeader struct {
	Mode              int
	D1, D2, D3, D4    int
	Rationale         string
	Raw               string
}

// headerPattern matches production comma-separated format:
// [MODE N | D1=X, D2=X, D3=X, D4=X] {Rationale}
//
// The old space-separated D5 format is intentionally NOT matched.
// Postures are no longer declared in-band; they are derived by Score().
var headerPattern = regexp.MustCompile(
	`^\[MODE ([123]) \| D1=([0-3]), D2=([0-3]), D3=([0-3]), D4=([0-3])\](?: (.+))?$`,
)

// ParseHeader parses the gate header from the first line of a response.
// Accepts comma-separated D1-D4 format: [MODE N | D1=X, D2=X, D3=X, D4=X] {Rationale}
// Returns an error if the header is missing, malformed, or uses the old space-separated format.
func ParseHeader(firstLine string) (*GateHeader, error) {
	m := headerPattern.FindStringSubmatch(strings.TrimSpace(firstLine))
	if m == nil {
		return nil, fmt.Errorf("invalid or missing gate header: %q\nExpected format: [MODE N | D1=X, D2=X, D3=X, D4=X] {Rationale}", firstLine)
	}

	parseInt := func(s string) int {
		n, err := strconv.Atoi(s)
		if err != nil {
			return 0
		}
		return n
	}

	h := &GateHeader{
		Mode: parseInt(m[1]),
		D1:   parseInt(m[2]),
		D2:   parseInt(m[3]),
		D3:   parseInt(m[4]),
		D4:   parseInt(m[5]),
		Raw:  firstLine,
	}

	// Capture optional rationale (capture group 6)
	if len(m) > 6 {
		h.Rationale = strings.TrimSpace(m[6])
	}

	return h, nil
}

// ValidateDecision checks that the Mode and derived postures are consistent with D-scores.
// It only validates dimensions + mode consistency; posture validation is done by the
// orchestrator via Score() comparison.
//
// Invariants enforced:
//   - D3 >= 2 → Mode must be 3
//   - Mode 3 → D3 >= 2 OR D4 >= 3
//   - Mode >= 2 → Rationale must be non-empty
//   - D3 >= 2 AND D4 >= 3 → posturePriority must include "+++Forensic"
//   - All D-scores must be 0–3 (enforced by headerPattern regex, documented here)
func ValidateDecision(h *GateHeader) []string {
	var issues []string

	// Invariant: D3 >= 2 must be Mode 3
	if h.D3 >= 2 && h.Mode < 3 {
		issues = append(issues, fmt.Sprintf("D3=%d requires Mode 3, got Mode %d", h.D3, h.Mode))
	}

	// Invariant: Mode 3 requires D3 >= 2 or D4 >= 3
	if h.Mode == 3 && h.D3 < 2 && h.D4 < 3 {
		issues = append(issues, fmt.Sprintf(
			"Mode 3 requires D3 >= 2 or D4 >= 3, got D3=%d, D4=%d", h.D3, h.D4))
	}

	// Invariant: Mode >= 2 should have non-empty rationale
	if h.Mode >= 2 && strings.TrimSpace(h.Rationale) == "" {
		issues = append(issues, fmt.Sprintf(
			"Mode %d requires a non-empty rationale after the bracket — found none", h.Mode))
	}

	// Invariant: joint critical state (D3>=2, D4>=3) must produce +++Forensic posture.
	// We validate by calling Score() and checking the returned postures.
	if h.D3 >= 2 && h.D4 >= 3 {
		_, postures := Score([5]int{h.D1, h.D2, h.D3, h.D4, 0})
		hasForensic := false
		for _, p := range postures {
			if p == "+++Forensic" {
				hasForensic = true
				break
			}
		}
		if !hasForensic {
			issues = append(issues, fmt.Sprintf(
				"D3=%d AND D4=%d (joint critical) requires +++Forensic posture, got %v",
				h.D3, h.D4, postures))
		}
	}

	return issues
}

// ExtractFirstLine returns the first non-empty line of a response.
func ExtractFirstLine(response string) string {
	for _, line := range strings.Split(response, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}
