package verify_test

import (
	"strings"
	"testing"

	"github.com/rd-mg/architect-ai/internal/verify"
)

func TestVerdict_AllPassed(t *testing.T) {
	r := verify.BuildReport([]verify.CheckResult{
		{ID: "a", Status: verify.CheckStatusPassed},
		{ID: "b", Status: verify.CheckStatusPassed},
	})
	if r.Verdict != verify.VerdictReady {
		t.Errorf("want READY, got %s", r.Verdict)
	}
	if !r.Ready {
		t.Error("Ready should be true")
	}
}

func TestVerdict_WarningsOnly(t *testing.T) {
	r := verify.BuildReport([]verify.CheckResult{
		{ID: "a", Status: verify.CheckStatusWarning},
	})
	if r.Verdict != verify.VerdictWarning {
		t.Errorf("want WARNING, got %s", r.Verdict)
	}
	// Backward compat: Ready is still true when only warnings exist
	if !r.Ready {
		t.Error("Ready should be true when only warnings exist")
	}
}

func TestVerdict_WithFailures(t *testing.T) {
	r := verify.BuildReport([]verify.CheckResult{
		{ID: "a", Status: verify.CheckStatusFailed},
		{ID: "b", Status: verify.CheckStatusWarning},
	})
	if r.Verdict != verify.VerdictBlocked {
		t.Errorf("want BLOCKED, got %s", r.Verdict)
	}
	if r.Ready {
		t.Error("Ready should be false when failures exist")
	}
}

func TestVerdict_SkippedOnly_IsReady(t *testing.T) {
	r := verify.BuildReport([]verify.CheckResult{
		{ID: "a", Status: verify.CheckStatusSkipped},
	})
	if r.Verdict != verify.VerdictReady {
		t.Errorf("want READY for skipped-only, got %s", r.Verdict)
	}
}

func TestRender_ShowsFixHintOnFailure(t *testing.T) {
	r := verify.BuildReport([]verify.CheckResult{
		{
			ID:      "engram.binary",
			Status:  verify.CheckStatusFailed,
			Error:   "engram not found",
			FixHint: "architect-ai install --component engram",
		},
	})
	out := verify.RenderReport(r)
	if !strings.Contains(out, "fix:") {
		t.Errorf("expected fix: line, got:\n%s", out)
	}
	if !strings.Contains(out, "BLOCKED") {
		t.Errorf("expected BLOCKED banner, got:\n%s", out)
	}
}

func TestRender_NoFixHintOnPassed(t *testing.T) {
	r := verify.BuildReport([]verify.CheckResult{
		{ID: "a", Status: verify.CheckStatusPassed, FixHint: "some hint"},
	})
	out := verify.RenderReport(r)
	if strings.Contains(out, "fix:") {
		t.Errorf("fix: should NOT appear for passed checks, got:\n%s", out)
	}
}

func TestRender_FixHintOnWarning(t *testing.T) {
	r := verify.BuildReport([]verify.CheckResult{
		{
			ID:      "notebooklm.mcp-config",
			Status:  verify.CheckStatusWarning,
			Error:   "section not found",
			FixHint: "architect-ai install --component notebooklm",
		},
	})
	out := verify.RenderReport(r)
	if !strings.Contains(out, "fix:") {
		t.Errorf("expected fix: line for warning, got:\n%s", out)
	}
	if !strings.Contains(out, "WARNING") {
		t.Errorf("expected WARNING banner, got:\n%s", out)
	}
}
