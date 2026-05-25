package orchestrator

import (
	"testing"
)

func TestValidateSubAgentResponse_HeaderMatches(t *testing.T) {
	response := "[MODE 2 | D1=1, D2=2, D3=0, D4=1] Complex task\nSome content"
	expectedDims := Dims48{1, 2, 0, 1, 0}
	result := ValidateSubAgentResponse(response, expectedDims, 2, 0)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if !result.Matched {
		t.Error("expected Matched=true for matching modes")
	}
	if result.DeclaredMode != 2 {
		t.Errorf("expected DeclaredMode=2, got %d", result.DeclaredMode)
	}
	if result.ForceMode3 {
		t.Error("expected ForceMode3=false for first attempt with matching header")
	}
}

func TestValidateSubAgentResponse_MissingHeader(t *testing.T) {
	response := "This is a normal response without any gate header\nSome content"
	expectedDims := Dims48{0, 0, 0, 0, 0}
	result := ValidateSubAgentResponse(response, expectedDims, 1, 0)

	if result.Err == nil {
		t.Error("expected error for missing header")
	}
	if result.Matched {
		t.Error("expected Matched=false for missing header")
	}
	if result.ForceMode3 {
		t.Error("expected ForceMode3=false for first attempt")
	}
}

func TestValidateSubAgentResponse_HeaderMismatch(t *testing.T) {
	response := "[MODE 3 | D1=2, D2=0, D3=2, D4=0] High pressure\nContent"
	expectedDims := Dims48{0, 0, 0, 0, 0}
	result := ValidateSubAgentResponse(response, expectedDims, 1, 0)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if result.Matched {
		t.Error("expected Matched=false for mode mismatch")
	}
	if result.DeclaredMode != 3 {
		t.Errorf("expected DeclaredMode=3, got %d", result.DeclaredMode)
	}
	if result.ExpectedMode != 1 {
		t.Errorf("expected ExpectedMode=1, got %d", result.ExpectedMode)
	}
	if result.ForceMode3 {
		t.Error("expected ForceMode3=false for first attempt mismatch")
	}
}

func TestValidateSubAgentResponse_EmptyResponse(t *testing.T) {
	result := ValidateSubAgentResponse("", Dims48{0, 0, 0, 0, 0}, 1, 0)

	if result.Err == nil {
		t.Error("expected error for empty response")
	}
	if result.DeclaredMode != 0 {
		t.Errorf("expected DeclaredMode=0 for empty response, got %d", result.DeclaredMode)
	}
}

func TestValidateSubAgentResponse_CircuitBreakerEnforcement(t *testing.T) {
	// attempt >= 2 and mismatch -> ForceMode3
	response := "[MODE 1 | D1=0, D2=0, D3=0, D4=0]"
	result := ValidateSubAgentResponse(response, Dims48{0, 0, 0, 0, 0}, 2, 2)

	if result.Matched {
		t.Error("expected Matched=false (declared Mode 1 != expected Mode 2)")
	}
	// Mismatch + attempt >= 2 -> ForceMode3
	if !result.ForceMode3 {
		t.Error("expected ForceMode3=true for mismatch with attempt>=2")
	}
	if result.DeclaredMode != 1 {
		t.Errorf("expected DeclaredMode=1, got %d", result.DeclaredMode)
	}
}

func TestValidateSubAgentResponse_CircuitBreakerWithMissingHeader(t *testing.T) {
	result := ValidateSubAgentResponse("no header here", Dims48{0, 0, 0, 0, 0}, 1, 2)

	if result.Err == nil {
		t.Error("expected error for missing header")
	}
	if !result.ForceMode3 {
		t.Error("expected ForceMode3=true for missing header with attempt>=2")
	}
}

func TestValidateSubAgentResponse_DeclaresDims(t *testing.T) {
	response := "[MODE 2 | D1=2, D2=1, D3=0, D4=3] Large context task"
	result := ValidateSubAgentResponse(response, Dims48{2, 1, 0, 3, 0}, 2, 0)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}
	if !result.Matched {
		t.Error("expected Matched=true")
	}
	expected := Dims48{2, 1, 0, 3, 0}
	if result.DeclaredDims != expected {
		t.Errorf("expected DeclaredDims=%v, got %v", expected, result.DeclaredDims)
	}
	if result.ExpectedDims != expected {
		t.Errorf("expected ExpectedDims=%v, got %v", expected, result.ExpectedDims)
	}
}
