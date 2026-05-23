package gate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// GateHeader represents a parsed Adaptive Reasoning Gate header
type GateHeader struct {
	Mode                          int
	D1, D2, D3, D4, D5            int
	Posture1                      string
	Posture2                      string // empty if only one posture
	Raw                           string
}

// headerPattern matches: [MODE N | D1=X D2=X D3=X D4=X D5=X | POSTURE: +++P1 [+++P2]]
var headerPattern = regexp.MustCompile(
	`^\[MODE ([1-3]) \| D1=([0-3]) D2=([0-3]) D3=([0-3]) D4=([0-3]) D5=([0-3]) \| POSTURE: (\+\+\+\w+)(?: (\+\+\+\w+))?\]`,
)

// ParseHeader parses the gate header line from the first line of a response
func ParseHeader(firstLine string) (*GateHeader, error) {
	firstLine = strings.TrimSpace(firstLine)
	matches := headerPattern.FindStringSubmatch(firstLine)
	if len(matches) == 0 {
		return nil, fmt.Errorf("invalid header format: %q", firstLine)
	}

	mode, _ := strconv.Atoi(matches[1])
	d1, _ := strconv.Atoi(matches[2])
	d2, _ := strconv.Atoi(matches[3])
	d3, _ := strconv.Atoi(matches[4])
	d4, _ := strconv.Atoi(matches[5])
	d5, _ := strconv.Atoi(matches[6])

	p1 := matches[7]
	p2 := ""
	if len(matches) > 8 && matches[8] != "" {
		p2 = matches[8]
	}

	return &GateHeader{
		Mode:     mode,
		D1:       d1,
		D2:       d2,
		D3:       d3,
		D4:       d4,
		D5:       d5,
		Posture1: p1,
		Posture2: p2,
		Raw:      firstLine,
	}, nil
}

// ValidateDecision checks that the Mode and Postures are consistent with D-scores
func ValidateDecision(h *GateHeader) []string {
	var issues []string

	// D5 security override
	if h.D5 >= 2 {
		if h.Posture1 != "+++Adversarial" && h.Posture2 != "+++Adversarial" {
			issues = append(issues, "D5 >= 2 requires +++Adversarial posture")
		}
		if h.Mode < 2 {
			issues = append(issues, "D5 >= 2 requires minimum Mode 2")
		}
	}

	// D3 overrides
	if h.D3 >= 2 && h.Mode != 3 {
		issues = append(issues, "D3 >= 2 forces Mode 3")
	}

	return issues
}
