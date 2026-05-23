package gate

import (
	"strings"
	"testing"
)

func TestScore_Mode1_LowRisk(t *testing.T) {
	// D1-D5 scores
	dims := [5]int{0, 1, 0, 0, 0}
	mode, postures := Score(dims)

	if mode != 1 {
		t.Errorf("expected Mode 1, got %d", mode)
	}
	if len(postures) == 0 || postures[0] != "+++Pragmatic" {
		t.Errorf("expected +++Pragmatic, got %v", postures)
	}
}

func TestScore_Mode2_HighComplexity(t *testing.T) {
	dims := [5]int{3, 2, 0, 1, 0}
	mode, postures := Score(dims)

	if mode != 2 {
		t.Errorf("expected Mode 2, got %d", mode)
	}
	if len(postures) == 0 || postures[0] != "+++Critical" {
		t.Errorf("expected +++Critical, got %v", postures)
	}
}

func TestScore_D5Override(t *testing.T) {
	// D5 = 2 (Security) should force minimum Mode 2 and +++Adversarial
	dims := [5]int{0, 0, 0, 0, 2}
	mode, postures := Score(dims)

	if mode < 2 {
		t.Errorf("D5=2 should force minimum Mode 2, got %d", mode)
	}
	hasAdversarial := false
	for _, p := range postures {
		if p == "+++Adversarial" {
			hasAdversarial = true
		}
	}
	if !hasAdversarial {
		t.Errorf("expected +++Adversarial posture for D5>=2, got %v", postures)
	}
}

func TestScore_D3Override_Diagnostic(t *testing.T) {
	// D3 >= 2 forces Mode 3
	dims := [5]int{0, 0, 2, 0, 0}
	mode, postures := Score(dims)

	if mode != 3 {
		t.Errorf("D3>=2 should force Mode 3, got %d", mode)
	}
	if len(postures) == 0 || postures[0] != "+++Forensic" {
		t.Errorf("Mode 3 default should be +++Forensic, got %v", postures)
	}
}

func TestParseHeader_Valid(t *testing.T) {
	headerStr := "[MODE 2 | D1=2 D2=1 D3=0 D4=1 D5=0 | POSTURE: +++Critical +++Systemic]"
	h, err := ParseHeader(headerStr)
	if err != nil {
		t.Fatalf("unexpected error parsing valid header: %v", err)
	}
	if h.Mode != 2 {
		t.Errorf("expected Mode 2, got %d", h.Mode)
	}
	if h.D1 != 2 || h.D2 != 1 || h.D3 != 0 || h.D4 != 1 || h.D5 != 0 {
		t.Errorf("unexpected dimension scores in parsed header: %+v", h)
	}
	if h.Posture1 != "+++Critical" || h.Posture2 != "+++Systemic" {
		t.Errorf("unexpected postures in parsed header: Posture1=%s, Posture2=%s", h.Posture1, h.Posture2)
	}
}

func TestParseHeader_Invalid(t *testing.T) {
	invalidHeaders := []string{
		"This is not a header",
		"[MODE 4 | D1=2 D2=1 D3=0 D4=1 D5=0 | POSTURE: +++Critical]", // Invalid Mode
		"[MODE 2 | D1=4 D2=1 D3=0 D4=1 D5=0 | POSTURE: +++Critical]", // Invalid D1
		"[MODE 2 | D1=2 D2=1 D3=0 D4=1 D5=0 | POSTURE: Critical]",   // Missing +++
	}

	for _, hStr := range invalidHeaders {
		_, err := ParseHeader(hStr)
		if err == nil {
			t.Errorf("expected error parsing invalid header: %q", hStr)
		}
	}
}

func TestValidateDecision_D5RequiresAdversarial(t *testing.T) {
	// D5=2 but missing +++Adversarial
	h := &GateHeader{
		Mode:     2,
		D1:       0,
		D2:       0,
		D3:       0,
		D4:       0,
		D5:       2,
		Posture1: "+++Critical",
		Posture2: "+++Systemic",
	}

	issues := ValidateDecision(h)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue, "D5 >= 2 requires +++Adversarial") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected validation issue for missing +++Adversarial when D5=2, got %v", issues)
	}
}

func TestValidateDecision_D3ForcesMode3(t *testing.T) {
	// D3=2 but Mode=2
	h := &GateHeader{
		Mode:     2,
		D1:       0,
		D2:       0,
		D3:       2,
		D4:       0,
		D5:       0,
		Posture1: "+++Critical",
		Posture2: "+++Systemic",
	}

	issues := ValidateDecision(h)
	found := false
	for _, issue := range issues {
		if strings.Contains(issue, "D3 >= 2 forces Mode 3") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected validation issue for D3=2 forcing Mode 3, got %v", issues)
	}
}
