package gate

import "testing"

// TestScore_Mode1 tests that simple tasks with low dimensions produce Mode 1.
func TestScore_Mode1(t *testing.T) {
	dims := [5]int{0, 0, 0, 0, 0}
	mode, postures := Score(dims)
	if mode != 1 {
		t.Errorf("expected Mode 1, got Mode %d", mode)
	}
	if len(postures) != 1 || postures[0] != "+++Pragmatic" {
		t.Errorf("expected [+++Pragmatic], got %v", postures)
	}
}

// TestScore_Mode1_D5Low checks D5=0 or 1 does not force Mode up.
func TestScore_Mode1_D5Low(t *testing.T) {
	dims := [5]int{0, 1, 0, 0, 1}
	mode, postures := Score(dims)
	// D1+D2=1, D3=0, D4=0, D5=1 -> Mode 1
	if mode != 1 {
		t.Errorf("expected Mode 1 (D1+D2=1, D5=1), got Mode %d", mode)
	}
	if len(postures) != 1 || postures[0] != "+++Pragmatic" {
		t.Errorf("expected [+++Pragmatic], got %v", postures)
	}
}

// TestScore_Mode2_Complexity tests D1+D2=3 -> Mode 2.
func TestScore_Mode2_Complexity(t *testing.T) {
	dims := [5]int{2, 1, 0, 0, 0}
	mode, postures := Score(dims)
	if mode != 2 {
		t.Errorf("expected Mode 2 (D1+D2=3), got Mode %d", mode)
	}
	if len(postures) < 1 || postures[0] != "+++Critical" {
		t.Errorf("expected +++Critical first, got %v", postures)
	}
}

// TestScore_Mode2_D1HighTrigger tests D1=2,D2=1 (sum=3) -> Mode 2 with +++Systemic.
func TestScore_Mode2_D1HighTrigger(t *testing.T) {
	dims := [5]int{2, 1, 0, 0, 0}
	mode, postures := Score(dims)
	if mode != 2 {
		t.Errorf("expected Mode 2, got Mode %d", mode)
	}
	if len(postures) < 2 || postures[1] != "+++Systemic" {
		t.Errorf("expected Systemic for D1>=2, got %v", postures)
	}
}

// TestScore_Mode2_D2High tests D2>=2 adds +++Socratic in Mode 2.
// D2 alone requires D1+D2 >= 3, so D1 must be at least 1.
func TestScore_Mode2_D2High(t *testing.T) {
	dims := [5]int{1, 2, 0, 0, 0}
	mode, postures := Score(dims)
	if mode != 2 {
		t.Errorf("expected Mode 2 (D1+D2=3), got Mode %d", mode)
	}
	if len(postures) < 2 || postures[1] != "+++Socratic" {
		t.Errorf("expected Socratic for D2>=2, got %v", postures)
	}
}

// TestScore_Mode2_D3Error tests D3=1 adds +++Forensic in Mode 2.
func TestScore_Mode2_D3Error(t *testing.T) {
	dims := [5]int{0, 0, 1, 0, 0}
	mode, postures := Score(dims)
	if mode != 2 {
		t.Errorf("expected Mode 2 (D3=1), got Mode %d", mode)
	}
	if len(postures) < 2 || postures[1] != "+++Forensic" {
		t.Errorf("expected Forensic for D3=1, got %v", postures)
	}
}

// TestScore_Mode3_ErrorPressure tests D3=2 forces Mode 3.
func TestScore_Mode3_ErrorPressure(t *testing.T) {
	dims := [5]int{0, 0, 2, 0, 0}
	mode, postures := Score(dims)
	if mode != 3 {
		t.Errorf("expected Mode 3 (D3=2), got Mode %d", mode)
	}
	if len(postures) < 1 || postures[0] != "+++Forensic" {
		t.Errorf("expected Forensic for D3>=2, got %v", postures)
	}
}

// TestScore_Mode3_ContextSaturated tests D4=3 forces Mode 3 with +++Pragmatic.
func TestScore_Mode3_ContextSaturated(t *testing.T) {
	dims := [5]int{0, 0, 0, 3, 0}
	mode, postures := Score(dims)
	if mode != 3 {
		t.Errorf("expected Mode 3 (D4=3), got Mode %d", mode)
	}
	if len(postures) != 1 || postures[0] != "+++Pragmatic" {
		t.Errorf("expected [+++Pragmatic] for context saturation, got %v", postures)
	}
}

// TestScore_Mode3_Architectural tests D1=3 in Mode 3.
func TestScore_Mode3_Architectural(t *testing.T) {
	dims := [5]int{3, 0, 2, 0, 0}
	mode, postures := Score(dims)
	if mode != 3 {
		t.Errorf("expected Mode 3, got Mode %d", mode)
	}
	if len(postures) < 2 || postures[0] != "+++Systemic" || postures[1] != "+++Adversarial" {
		t.Errorf("expected [+++Systemic, +++Adversarial] for D1=3, got %v", postures)
	}
}

// TestScore_D5_Equal3 tests D5=3 forces Mode 3 with +++Adversarial + +++Forensic.
func TestScore_D5_Equal3(t *testing.T) {
	dims := [5]int{0, 0, 0, 0, 3}
	mode, postures := Score(dims)
	if mode != 3 {
		t.Errorf("expected Mode 3 (D5=3), got Mode %d", mode)
	}
	if len(postures) != 2 {
		t.Fatalf("expected 2 postures, got %d: %v", len(postures), postures)
	}
	if postures[0] != "+++Adversarial" {
		t.Errorf("expected +++Adversarial first for D5=3, got %s", postures[0])
	}
	if postures[1] != "+++Forensic" {
		t.Errorf("expected +++Forensic second for D5=3, got %s", postures[1])
	}
}

// TestScore_D5_Equal2 tests D5=2 forces Mode 2 with +++Adversarial.
func TestScore_D5_Equal2(t *testing.T) {
	dims := [5]int{0, 0, 0, 0, 2}
	mode, postures := Score(dims)
	if mode != 2 {
		t.Errorf("expected Mode 2 (D5=2), got Mode %d", mode)
	}
	if len(postures) != 2 {
		t.Fatalf("expected 2 postures, got %d: %v", len(postures), postures)
	}
	if postures[0] != "+++Adversarial" {
		t.Errorf("expected +++Adversarial first for D5=2, got %s", postures[0])
	}
	// D1=0 so second should be +++Critical
	if postures[1] != "+++Critical" {
		t.Errorf("expected +++Critical second for D5=2 (D1=0), got %s", postures[1])
	}
}

// TestScore_D5_Equal2_D1High tests D5=2 with D1>=2 gives +++Adversarial + +++Systemic.
func TestScore_D5_Equal2_D1High(t *testing.T) {
	dims := [5]int{2, 0, 0, 0, 2}
	mode, postures := Score(dims)
	if mode != 2 {
		t.Errorf("expected Mode 2 (D5=2), got Mode %d", mode)
	}
	if len(postures) != 2 {
		t.Fatalf("expected 2 postures, got %d: %v", len(postures), postures)
	}
	if postures[0] != "+++Adversarial" {
		t.Errorf("expected +++Adversarial first for D5=2, got %s", postures[0])
	}
	if postures[1] != "+++Systemic" {
		t.Errorf("expected +++Systemic second for D5=2 with D1>=2, got %s", postures[1])
	}
}

// TestScore_D5_Equal1_NoEffect tests D5=1 does not change routing.
func TestScore_D5_Equal1_NoEffect(t *testing.T) {
	dims := [5]int{0, 0, 0, 0, 1}
	mode, postures := Score(dims)
	if mode != 1 {
		t.Errorf("expected Mode 1 (D5=1 only), got Mode %d", mode)
	}
	if len(postures) != 1 || postures[0] != "+++Pragmatic" {
		t.Errorf("expected [+++Pragmatic], got %v", postures)
	}
}

// TestScore_D5_Equal3_overridesMode3 tests D5=3 with high D3 does not change posture.
func TestScore_D5_Equal3_OverridesPosture(t *testing.T) {
	// D5=3 should give +++Adversarial + +++Forensic regardless of other dims
	dims := [5]int{3, 0, 2, 0, 3}
	mode, postures := Score(dims)
	if mode != 3 {
		t.Errorf("expected Mode 3, got Mode %d", mode)
	}
	if len(postures) != 2 {
		t.Fatalf("expected 2 postures, got %d: %v", len(postures), postures)
	}
	if postures[0] != "+++Adversarial" {
		t.Errorf("expected +++Adversarial first, got %s", postures[0])
	}
	if postures[1] != "+++Forensic" {
		t.Errorf("expected +++Forensic second, got %s", postures[1])
	}
}

// TestScore_Mode2_Default tests that Mode 2 without high dims just has +++Critical.
func TestScore_Mode2_Default(t *testing.T) {
	dims := [5]int{1, 2, 0, 0, 0}
	mode, postures := Score(dims)
	if mode != 2 {
		t.Errorf("expected Mode 2, got Mode %d", mode)
	}
	// D2=2 -> should add +++Socratic
	if len(postures) < 2 || postures[1] != "+++Socratic" {
		t.Errorf("expected +++Socratic for D2=2, got %v", postures)
	}
}

// TestScore_Mode2_NoSecondary tests Mode 2 when no secondary dim qualifies.
func TestScore_Mode2_NoSecondary(t *testing.T) {
	dims := [5]int{1, 1, 0, 0, 0}
	_, postures := Score(dims)
	if len(postures) != 1 {
		t.Errorf("expected only 1 posture (+++Critical), got %v", postures)
	}
}

// TestScore_Mode3_Default tests default Mode 3 posture.
func TestScore_Mode3_Default(t *testing.T) {
	dims := [5]int{1, 0, 2, 1, 0}
	mode, postures := Score(dims)
	if mode != 3 {
		t.Errorf("expected Mode 3, got Mode %d", mode)
	}
	// D3>=2, D1=1 (not >=3), D4=1 (not >=3) -> default: +++Forensic, +++Pragmatic
	if len(postures) != 2 || postures[0] != "+++Forensic" || postures[1] != "+++Pragmatic" {
		t.Errorf("expected [+++Forensic, +++Pragmatic], got %v", postures)
	}
}
