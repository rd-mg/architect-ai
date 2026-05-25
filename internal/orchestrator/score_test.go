package orchestrator

import (
	"testing"
)

func TestCognitiveScorer_SimpleTask(t *testing.T) {
	ctx := TaskContext{
		PhaseName:         "sdd-apply",
		ChangeID:          "test-change",
		FileCount:         1,
		CrossPackage:      false,
		HasSpecs:          true,
		AttemptCount:      0,
		ContextEstimateKB: 5,
	}
	dims, mode, postures := CognitiveScorer(ctx)
	if mode != 1 {
		t.Errorf("expected Mode 1, got Mode %d", mode)
	}
	if len(postures) != 1 || postures[0] != "+++Pragmatic" {
		t.Errorf("expected [+++Pragmatic], got %v", postures)
	}
	expectedDims := Dims48{0, 0, 0, 0, 0}
	if dims != expectedDims {
		t.Errorf("expected dims %v, got %v", expectedDims, dims)
	}
}

func TestCognitiveScorer_CrossPackageTask(t *testing.T) {
	ctx := TaskContext{
		PhaseName:         "sdd-design",
		FileCount:         3,
		CrossPackage:      true,
		HasSpecs:          false, // no specs -> D2=1 (design phase)
		AttemptCount:      0,
		ContextEstimateKB: 30,
	}
	dims, mode, postures := CognitiveScorer(ctx)
	// D1=2 (cross-package), D2=1 (no specs + design) -> D1+D2=3 -> Mode 2
	if mode != 2 {
		t.Errorf("expected Mode 2 (D1+D2=3), got Mode %d (dims=%v)", mode, dims)
	}
	if len(postures) < 1 || postures[0] != "+++Critical" {
		t.Errorf("expected Critical posture first, got %v", postures)
	}
	// D1 >= 2 due to cross-package + 3 files -> should include Systemic
	if dims[0] >= 2 {
		found := false
		for _, p := range postures {
			if p == "+++Systemic" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected Systemic posture for D1>=2, got %v", postures)
		}
	}
}

func TestCognitiveScorer_HighErrorPressure(t *testing.T) {
	ctx := TaskContext{
		PhaseName:         "sdd-apply",
		FileCount:         2,
		CrossPackage:      false,
		HasSpecs:          true,
		AttemptCount:      2,
		ContextEstimateKB: 20,
	}
	dims, mode, postures := CognitiveScorer(ctx)
	if mode != 3 {
		t.Errorf("expected Mode 3 (D3=2 -> force Mode 3), got Mode %d", mode)
	}
	if dims[2] != 3 {
		t.Errorf("expected D3=3 (circuit breaker override), got D3=%d", dims[2])
	}
	if len(postures) == 0 {
		t.Fatal("expected postures, got none")
	}
	foundForensic := false
	for _, p := range postures {
		if p == "+++Forensic" {
			foundForensic = true
			break
		}
	}
	if !foundForensic {
		t.Errorf("expected Forensic posture for error pressure, got %v", postures)
	}
}

func TestCognitiveScorer_ArchitecturalChange(t *testing.T) {
	ctx := TaskContext{
		PhaseName:         "sdd-design",
		FileCount:         8,
		CrossPackage:      true,
		HasSpecs:          false,
		AttemptCount:      0,
		ContextEstimateKB: 80,
	}
	dims, mode, _ := CognitiveScorer(ctx)
	if dims[0] != 3 {
		t.Errorf("expected D1=3 for architectural change, got D1=%d", dims[0])
	}
	// D1+D2 >=3 + D3=0 -> Mode 2 (D2=1 from no specs + design phase)
	if mode != 2 {
		t.Errorf("expected Mode 2, got Mode %d", mode)
	}
}

func TestCognitiveScorer_CircuitBreakerOverride(t *testing.T) {
	ctx := TaskContext{
		PhaseName:         "sdd-apply",
		FileCount:         1,
		CrossPackage:      false,
		HasSpecs:          true,
		AttemptCount:      2, // triggers circuit breaker
		ContextEstimateKB: 5,
	}
	dims, mode, _ := CognitiveScorer(ctx)
	if dims[2] != 3 {
		t.Errorf("circuit breaker: expected D3=3, got D3=%d", dims[2])
	}
	if mode != 3 {
		t.Errorf("circuit breaker: expected Mode 3, got Mode %d", mode)
	}
}

func TestCognitiveScorer_ZeroAttempt(t *testing.T) {
	ctx := TaskContext{
		PhaseName:         "sdd-init",
		FileCount:         0,
		CrossPackage:      false,
		HasSpecs:          false,
		AttemptCount:      0,
		ContextEstimateKB: 0,
	}
	dims, mode, postures := CognitiveScorer(ctx)
	if dims[2] != 0 {
		t.Errorf("expected D3=0 for first attempt, got D3=%d", dims[2])
	}
	// D2=3 (terra incognita from empty phase), D1=0 -> D1+D2=3 -> Mode 2
	if mode != 2 {
		t.Errorf("expected Mode 2 for terra incognita, got Mode %d", mode)
	}
	_ = postures
}

func TestCognitiveScorer_SmallContext(t *testing.T) {
	ctx := TaskContext{
		PhaseName:         "sdd-apply",
		FileCount:         1,
		CrossPackage:      false,
		HasSpecs:          true,
		AttemptCount:      0,
		ContextEstimateKB: 3,
	}
	dims, _, _ := CognitiveScorer(ctx)
	if dims[3] != 0 {
		t.Errorf("expected D4=0 for <10KB, got D4=%d", dims[3])
	}
}

func TestCognitiveScorer_LargeContext(t *testing.T) {
	ctx := TaskContext{
		PhaseName:         "sdd-verify",
		FileCount:         2,
		CrossPackage:      false,
		HasSpecs:          true,
		AttemptCount:      0,
		ContextEstimateKB: 120,
	}
	dims, mode, postures := CognitiveScorer(ctx)
	if dims[3] != 3 {
		t.Errorf("expected D4=3 for >100KB, got D4=%d", dims[3])
	}
	// D4=3 forces Mode 3
	if mode != 3 {
		t.Errorf("expected Mode 3 for D4=3, got Mode %d", mode)
	}
	if len(postures) != 1 || postures[0] != "+++Pragmatic" {
		t.Errorf("expected [+++Pragmatic] for context saturation, got %v", postures)
	}
}

func TestCognitiveScorer_TerraIncognitaPhase(t *testing.T) {
	ctx := TaskContext{
		PhaseName:         "",
		FileCount:         1,
		CrossPackage:      false,
		HasSpecs:          false,
		AttemptCount:      0,
		ContextEstimateKB: 5,
	}
	_, mode, _ := CognitiveScorer(ctx)
	if mode != 2 {
		t.Errorf("expected Mode 2 for terra incognita, got Mode %d", mode)
	}
}

func TestCognitiveScorer_EconomicPosture(t *testing.T) {
	ctx := TaskContext{
		PhaseName:        "sdd-propose",
		FileCount:        3,
		CrossPackage:     true,
		HasSpecs:         false,
		AttemptCount:     0,
		ContextEstimateKB: 30,
		IsCostSensitive:  true,
	}
	_, mode, postures := CognitiveScorer(ctx)
	if mode < 2 {
		t.Fatalf("expected mode >= 2 for cross-package propose, got Mode %d", mode)
	}
	// Should have +++Economic as one of the postures (replaces or augments second)
	found := false
	for _, p := range postures {
		if p == "+++Economic" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected +++Economic posture for cost-sensitive task, got %v", postures)
	}
	if len(postures) > 2 {
		t.Errorf("expected max 2 postures, got %d: %v", len(postures), postures)
	}
}

func TestCognitiveScorer_EmpiricalPosture(t *testing.T) {
	// Use a task that starts at Mode 2 to test empirical override
	ctx := TaskContext{
		PhaseName:         "sdd-design",
		FileCount:         3,
		CrossPackage:      true,
		HasSpecs:          false,
		AttemptCount:      0,
		ContextEstimateKB: 30,
		IsMeasurementTask: true,
	}
	_, mode, postures := CognitiveScorer(ctx)
	if mode < 2 {
		t.Fatalf("expected mode >= 2, got Mode %d", mode)
	}
	found := false
	for _, p := range postures {
		if p == "+++Empirical" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected +++Empirical posture for measurement task, got %v", postures)
	}
	if len(postures) > 2 {
		t.Errorf("expected max 2 postures, got %d: %v", len(postures), postures)
	}
}

func TestCognitiveScorer_CostTrumpsMeasurement(t *testing.T) {
	ctx := TaskContext{
		PhaseName:         "sdd-design",
		FileCount:         3,
		CrossPackage:      true,
		HasSpecs:          false,
		AttemptCount:      0,
		ContextEstimateKB: 30,
		IsCostSensitive:   true,
		IsMeasurementTask: true,
	}
	_, mode, postures := CognitiveScorer(ctx)
	if mode < 2 {
		t.Fatalf("expected mode >= 2, got Mode %d", mode)
	}
	// When both, cost should win (+++Economic, not +++Empirical)
	hasEconomic := false
	hasEmpirical := false
	for _, p := range postures {
		if p == "+++Economic" {
			hasEconomic = true
		}
		if p == "+++Empirical" {
			hasEmpirical = true
		}
	}
	if !hasEconomic {
		t.Errorf("expected +++Economic (cost trumps measurement), got %v", postures)
	}
	if hasEmpirical {
		t.Errorf("did NOT expect +++Empirical (cost trumps measurement), got %v", postures)
	}
}

func TestCognitiveScorer_SecuritySensitive(t *testing.T) {
	ctx := TaskContext{
		PhaseName:           "sdd-apply",
		FileCount:           1,
		CrossPackage:        false,
		HasSpecs:            true,
		AttemptCount:        0,
		ContextEstimateKB:   5,
		IsSecuritySensitive: true,
	}
	dims, mode, postures := CognitiveScorer(ctx)
	// D5 should be 1 (security-aware), but D5 alone at level 1 doesn't force higher mode
	// D1=0, D2=0, D5=1 -> Mode 1 still
	if mode != 1 {
		t.Errorf("expected Mode 1 (D5=1 only), got Mode %d (dims=%v)", mode, dims)
	}
	if dims[4] != 1 {
		t.Errorf("expected D5=1 for security-sensitive, got D5=%d", dims[4])
	}
	if len(postures) != 1 || postures[0] != "+++Pragmatic" {
		t.Errorf("expected [+++Pragmatic], got %v", postures)
	}
}

func TestCognitiveScorer_SecuritySensitiveWithFailure(t *testing.T) {
	ctx := TaskContext{
		PhaseName:           "sdd-apply",
		FileCount:           2,
		CrossPackage:        false,
		HasSpecs:            true,
		AttemptCount:        1,
		ContextEstimateKB:   20,
		IsSecuritySensitive: true,
	}
	dims, mode, postures := CognitiveScorer(ctx)
	// D5=2 (security concern, attempt=1), D3=1 -> Mode 2 (D3>=1)
	if mode != 2 {
		t.Errorf("expected Mode 2, got Mode %d (dims=%v)", mode, dims)
	}
	if dims[4] != 2 {
		t.Errorf("expected D5=2 (security concern with 1 attempt), got D5=%d", dims[4])
	}
	// D5=2 overrides posture: +++Adversarial + +++Critical/Systemic
	if len(postures) != 2 {
		t.Fatalf("expected 2 postures, got %d: %v", len(postures), postures)
	}
	if postures[0] != "+++Adversarial" {
		t.Errorf("expected +++Adversarial first from D5=2, got %s", postures[0])
	}
}

func TestCognitiveScorer_SecurityIncident(t *testing.T) {
	ctx := TaskContext{
		PhaseName:           "sdd-apply",
		FileCount:           1,
		CrossPackage:        false,
		HasSpecs:            true,
		AttemptCount:        2,
		ContextEstimateKB:   5,
		IsSecuritySensitive: true,
	}
	dims, mode, postures := CognitiveScorer(ctx)
	// D5=3 (security incident), D3=3 (circuit breaker) -> Mode 3
	if mode != 3 {
		t.Errorf("expected Mode 3, got Mode %d (dims=%v)", mode, dims)
	}
	if dims[4] != 3 {
		t.Errorf("expected D5=3 (security incident with 2 attempts), got D5=%d", dims[4])
	}
	if len(postures) != 2 {
		t.Fatalf("expected 2 postures, got %d: %v", len(postures), postures)
	}
	if postures[0] != "+++Adversarial" {
		t.Errorf("expected +++Adversarial first from D5=3, got %s", postures[0])
	}
	if postures[1] != "+++Forensic" {
		t.Errorf("expected +++Forensic second from D5=3, got %s", postures[1])
	}
}
