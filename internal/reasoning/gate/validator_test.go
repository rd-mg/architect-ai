// internal/reasoning/gate/validator_test.go
package gate

import "testing"

func TestParseHeader_Valid(t *testing.T) {
	line := "[MODE 2 | D1=1, D2=2, D3=0, D4=1] Complex task"
	h, err := ParseHeader(line)
	if err != nil {
		t.Fatalf("ParseHeader: %v", err)
	}
	if h.Mode != 2 {
		t.Errorf("Mode: want 2, got %d", h.Mode)
	}
	if h.D1 != 1 {
		t.Errorf("D1: want 1, got %d", h.D1)
	}
	if h.D2 != 2 {
		t.Errorf("D2: want 2, got %d", h.D2)
	}
	if h.D3 != 0 {
		t.Errorf("D3: want 0, got %d", h.D3)
	}
	if h.D4 != 1 {
		t.Errorf("D4: want 1, got %d", h.D4)
	}
	if h.Rationale != "Complex task" {
		t.Errorf("Rationale: want %q, got %q", "Complex task", h.Rationale)
	}
}

func TestParseHeader_Mode1(t *testing.T) {
	line := "[MODE 1 | D1=0, D2=0, D3=0, D4=0]"
	h, err := ParseHeader(line)
	if err != nil {
		t.Fatalf("ParseHeader: %v", err)
	}
	if h.Mode != 1 {
		t.Errorf("Mode: want 1, got %d", h.Mode)
	}
	if h.D1 != 0 || h.D2 != 0 || h.D3 != 0 || h.D4 != 0 {
		t.Errorf("Dims: want [0,0,0,0], got [%d,%d,%d,%d]", h.D1, h.D2, h.D3, h.D4)
	}
	if h.Rationale != "" {
		t.Errorf("Rationale should be empty, got %q", h.Rationale)
	}
}

func TestParseHeader_WithRationale(t *testing.T) {
	line := "[MODE 3 | D1=2, D2=0, D3=2, D4=0] Repeated failure in cross-module refactor"
	h, err := ParseHeader(line)
	if err != nil {
		t.Fatalf("ParseHeader: %v", err)
	}
	if h.Mode != 3 {
		t.Errorf("Mode: want 3, got %d", h.Mode)
	}
	if h.D1 != 2 || h.D2 != 0 || h.D3 != 2 || h.D4 != 0 {
		t.Errorf("Dims: want [2,0,2,0], got [%d,%d,%d,%d]", h.D1, h.D2, h.D3, h.D4)
	}
	if h.Rationale != "Repeated failure in cross-module refactor" {
		t.Errorf("Rationale: want %q, got %q", "Repeated failure in cross-module refactor", h.Rationale)
	}
}

func TestParseHeader_MissingRationale(t *testing.T) {
	line := "[MODE 2 | D1=1, D2=0, D3=0, D4=2]"
	h, err := ParseHeader(line)
	if err != nil {
		t.Fatalf("ParseHeader: %v", err)
	}
	if h.Mode != 2 {
		t.Errorf("Mode: want 2, got %d", h.Mode)
	}
	if h.Rationale != "" {
		t.Errorf("Rationale should be empty, got %q", h.Rationale)
	}
}

func TestParseHeader_RejectsSpaceSeparated(t *testing.T) {
	// Old space-separated format with D5 and POSTURE must NOT match
	cases := []string{
		"[MODE 1 | D1=0 D2=0 D3=0 D4=0 D5=0 | POSTURE: +++Pragmatic]",
		"[MODE 2 | D1=2 D2=1 D3=0 D4=1 D5=0]",
		"[MODE 1 | D1=0 D2=0 D3=0 D4=0]",
	}
	for _, c := range cases {
		_, err := ParseHeader(c)
		if err == nil {
			t.Errorf("expected error for space-separated header: %q", c)
		}
	}
}

func TestParseHeader_RejectsInvalidMode(t *testing.T) {
	line := "[MODE 4 | D1=0, D2=0, D3=0, D4=0]"
	_, err := ParseHeader(line)
	if err == nil {
		t.Error("expected error for MODE 4")
	}
}

func TestParseHeader_RejectsModeZero(t *testing.T) {
	line := "[MODE 0 | D1=0, D2=0, D3=0, D4=0]"
	_, err := ParseHeader(line)
	if err == nil {
		t.Error("expected error for MODE 0")
	}
}

func TestParseHeader_EmptyInput(t *testing.T) {
	cases := []string{"", "   ", "\n\n"}
	for _, c := range cases {
		_, err := ParseHeader(c)
		if err == nil {
			t.Errorf("expected error for empty/whitespace input: %q", c)
		}
	}
}

func TestParseHeader_NormalResponse(t *testing.T) {
	line := "This is a normal response without any gate header"
	_, err := ParseHeader(line)
	if err == nil {
		t.Error("expected error for non-header response")
	}
}

func TestParseHeader_AllMaxDimensions(t *testing.T) {
	line := "[MODE 3 | D1=3, D2=3, D3=3, D4=3] All dimensions at maximum"
	h, err := ParseHeader(line)
	if err != nil {
		t.Fatalf("ParseHeader: %v", err)
	}
	if h.Mode != 3 || h.D1 != 3 || h.D2 != 3 || h.D3 != 3 || h.D4 != 3 {
		t.Errorf("want Mode=3, D1=3, D2=3, D3=3, D4=3; got Mode=%d, [%d,%d,%d,%d]", h.Mode, h.D1, h.D2, h.D3, h.D4)
	}
}

func TestParseHeader_TrailingD5Ignored(t *testing.T) {
	// D5 in comma-separated list makes the header invalid per v3 format.
	// Per spec: "SHALL either ignore D5 or error; MUST NOT silently accept D5 as valid field."
	// We choose to error — clean rejection is safer than silent interpretation.
	line := "[MODE 2 | D1=1, D2=2, D3=0, D4=1, D5=0]"
	_, err := ParseHeader(line)
	if err == nil {
		t.Errorf("expected error for header with D5 field")
	}
}

func TestValidateDecision_D3ForcesMode3(t *testing.T) {
	h := &GateHeader{Mode: 2, D3: 2}
	issues := ValidateDecision(h)
	if len(issues) == 0 {
		t.Error("D3=2 with Mode 2 should produce issue (requires Mode 3)")
	}
}

func TestValidateDecision_D3ZeroMode3(t *testing.T) {
	h := &GateHeader{Mode: 3, D3: 0, D4: 0}
	issues := ValidateDecision(h)
	if len(issues) == 0 {
		t.Error("Mode 3 with D3=0, D4=0 should produce issue")
	}
}

func TestValidateDecision_ValidMode3(t *testing.T) {
	h := &GateHeader{Mode: 3, D3: 2, D4: 1}
	issues := ValidateDecision(h)
	if len(issues) > 0 {
		t.Errorf("valid Mode 3 with D3=2 should have no issues, got: %v", issues)
	}
}

func TestValidateDecision_ValidMode2(t *testing.T) {
	h := &GateHeader{Mode: 2, D3: 0, D4: 0}
	issues := ValidateDecision(h)
	if len(issues) > 0 {
		t.Errorf("valid Mode 2 should have no issues, got: %v", issues)
	}
}

func TestValidateDecision_ValidMode1(t *testing.T) {
	h := &GateHeader{Mode: 1, D3: 0, D4: 0}
	issues := ValidateDecision(h)
	if len(issues) > 0 {
		t.Errorf("valid Mode 1 should have no issues, got: %v", issues)
	}
}

func TestExtractFirstLine(t *testing.T) {
	response := "\n\n[MODE 1 | D1=0, D2=0, D3=0, D4=0]\n\nSome content"
	got := ExtractFirstLine(response)
	if got != "[MODE 1 | D1=0, D2=0, D3=0, D4=0]" {
		t.Errorf("unexpected first line: %q", got)
	}
}

func TestExtractFirstLine_EmptyResponse(t *testing.T) {
	got := ExtractFirstLine("")
	if got != "" {
		t.Errorf("expected empty for empty response, got %q", got)
	}
}

func TestExtractFirstLine_WhitespaceOnly(t *testing.T) {
	got := ExtractFirstLine("  \n  \n  ")
	if got != "" {
		t.Errorf("expected empty for whitespace-only, got %q", got)
	}
}
