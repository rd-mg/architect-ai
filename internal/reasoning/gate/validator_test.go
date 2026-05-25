// internal/reasoning/gate/validator_test.go
package gate

import "testing"

func TestParseHeader_Valid(t *testing.T) {
	line := "[MODE 2 | D1=2 D2=1 D3=0 D4=1 D5=0 | POSTURE: +++Critical +++Systemic]"
	h, err := ParseHeader(line)
	if err != nil {
		t.Fatalf("ParseHeader: %v", err)
	}
	if h.Mode != 2 { t.Errorf("Mode: want 2, got %d", h.Mode) }
	if h.D1 != 2   { t.Errorf("D1: want 2, got %d", h.D1) }
	if h.D5 != 0   { t.Errorf("D5: want 0, got %d", h.D5) }
	if h.Posture1 != "+++Critical" { t.Errorf("Posture1: %s", h.Posture1) }
	if h.Posture2 != "+++Systemic" { t.Errorf("Posture2: %s", h.Posture2) }
}

func TestParseHeader_SinglePosture(t *testing.T) {
	line := "[MODE 1 | D1=0 D2=0 D3=0 D4=0 D5=0 | POSTURE: +++Pragmatic]"
	h, err := ParseHeader(line)
	if err != nil { t.Fatalf("ParseHeader: %v", err) }
	if h.Posture1 != "+++Pragmatic" { t.Errorf("Posture1: %s", h.Posture1) }
	if h.Posture2 != "" { t.Errorf("Posture2 should be empty: %s", h.Posture2) }
}

func TestParseHeader_Invalid(t *testing.T) {
	cases := []string{
		"This is a normal response without gate header",
		"[MODE 4 | D1=0 D2=0 D3=0 D4=0 D5=0 | POSTURE: +++Pragmatic]", // MODE 4 invalid
		"",
	}
	for _, c := range cases {
		_, err := ParseHeader(c)
		if err == nil {
			t.Errorf("expected error for invalid header: %q", c)
		}
	}
}

func TestValidateDecision_D5RequiresAdversarial(t *testing.T) {
	h := &GateHeader{Mode: 2, D5: 2, Posture1: "+++Critical", Posture2: "+++Systemic"}
	issues := ValidateDecision(h)
	if len(issues) == 0 {
		t.Error("D5=2 without +++Adversarial should produce issues")
	}
}

func TestValidateDecision_D5AdversarialCorrect(t *testing.T) {
	h := &GateHeader{Mode: 2, D5: 2, Posture1: "+++Adversarial", Posture2: "+++Critical"}
	issues := ValidateDecision(h)
	if len(issues) > 0 {
		t.Errorf("valid D5=2 setup should have no issues, got: %v", issues)
    }
}

func TestValidateDecision_D3ForcesMode3(t *testing.T) {
	h := &GateHeader{Mode: 2, D3: 2, Posture1: "+++Forensic", Posture2: "+++Pragmatic"}
	issues := ValidateDecision(h)
	if len(issues) == 0 {
		t.Error("D3=2 with Mode 2 should produce issue (requires Mode 3)")
	}
}

func TestValidateDecision_DuplicatePostures(t *testing.T) {
	h := &GateHeader{Mode: 2, Posture1: "+++Critical", Posture2: "+++Critical"}
	issues := ValidateDecision(h)
	if len(issues) == 0 {
		t.Error("duplicate postures should produce issue")
	}
}

func TestExtractFirstLine(t *testing.T) {
	response := "\n\n[MODE 1 | D1=0 D2=0 D3=0 D4=0 D5=0 | POSTURE: +++Pragmatic]\n\nSome content"
	got := ExtractFirstLine(response)
	if got != "[MODE 1 | D1=0 D2=0 D3=0 D4=0 D5=0 | POSTURE: +++Pragmatic]" {
		t.Errorf("unexpected first line: %q", got)
	}
}