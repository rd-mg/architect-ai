package app

import (
	"bytes"
	"context"
	"testing"
)

func TestSkillsPatternsCmd_NoArgs(t *testing.T) {
	// When engram is not installed, the command should return a clear error,
	// not panic or produce garbled output.
	var buf bytes.Buffer
	err := runSkillsPatternsCmd(context.Background(), nil, &buf)
	// Either succeeds (engram installed) or fails with a clear message.
	// We just check it does not panic.
	_ = err
}

func TestSkillsPatternsCmd_ClearWithoutSkill_ReturnsError(t *testing.T) {
	var buf bytes.Buffer
	err := runSkillsPatternsCmd(context.Background(), []string{"--clear"}, &buf)
	if err == nil {
		t.Error("expected error when --clear used without --skill")
	}
}

func TestSkillsPatternsCmd_ClearWithSkill_PrintsInstructions(t *testing.T) {
	var buf bytes.Buffer
	err := runSkillsPatternsCmd(context.Background(), []string{"--clear", "--skill", "sdd-apply"}, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !contains(out, "sdd-apply") {
		t.Errorf("expected skill name in output, got: %s", out)
	}
	if !contains(out, "engram") {
		t.Errorf("expected engram instructions in output, got: %s", out)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
