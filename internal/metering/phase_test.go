package metering

import (
	"strings"
	"sync"
	"testing"
	"time"
)

func TestRecordPhaseStart_TracksTokens(t *testing.T) {
	s := NewSessionStats("sdd-orchestrator", "test-session-1")
	s.RecordPhaseStart("sdd-explore")
	s.Add(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 5000, CompletionTokens: 1000})
	s.RecordPhaseEnd("sdd-explore")

	s.RecordPhaseStart("sdd-apply")
	s.Add(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 15000, CompletionTokens: 8000})
	s.RecordPhaseEnd("sdd-apply")

	bd := s.PhaseBreakdown()
	if bd["sdd-explore"] != 6000 {
		t.Errorf("sdd-explore: got %d, want 6000", bd["sdd-explore"])
	}
	if bd["sdd-apply"] != 23000 {
		t.Errorf("sdd-apply: got %d, want 23000", bd["sdd-apply"])
	}
	if s.TotalTokens() != 29000 {
		t.Errorf("TotalTokens: got %d, want 29000", s.TotalTokens())
	}
}

func TestRecordPhaseStart_NoPhase_GlobalStillAccumulates(t *testing.T) {
	s := NewSessionStats("general", "test-session-2")
	s.Add(UsageDelta{Model: "claude-haiku-4", PromptTokens: 1000, CompletionTokens: 200})
	if s.TotalTokens() != 1200 {
		t.Errorf("global total: got %d, want 1200", s.TotalTokens())
	}
	bd := s.PhaseBreakdown()
	if len(bd) != 0 {
		t.Errorf("no phases started — breakdown should be empty, got %v", bd)
	}
}

func TestBudgetAlert_FiresWhenExceeded(t *testing.T) {
	alerted := make(chan string, 1)
	s := NewSessionStats("sdd-orchestrator", "test-session-3")
	s.WithBudgetAlert(5000, func(phase string, used, limit int64) {
		alerted <- phase
	})

	s.RecordPhaseStart("sdd-apply")
	s.Add(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 6000})

	select {
	case phase := <-alerted:
		if phase != "sdd-apply" {
			t.Errorf("expected alert for sdd-apply, got %s", phase)
		}
	case <-time.After(200 * time.Millisecond):
		t.Error("budget alert should have fired within 200ms")
	}
}

func TestBudgetAlert_DoesNotFireBelowLimit(t *testing.T) {
	fired := false
	s := NewSessionStats("sdd-orchestrator", "test-session-4")
	s.WithBudgetAlert(10000, func(phase string, used, limit int64) {
		fired = true
	})

	s.RecordPhaseStart("sdd-spec")
	s.Add(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 3000, CompletionTokens: 500})
	time.Sleep(50 * time.Millisecond)
	if fired {
		t.Error("alert should not fire when under budget limit")
	}
}

func TestEngramPhaseCostContent_Format(t *testing.T) {
	s := NewSessionStats("sdd-orchestrator", "test-session-5")
	s.RecordPhaseStart("sdd-spec")
	s.Add(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 3000, CompletionTokens: 800, CachedTokens: 500})
	s.RecordPhaseEnd("sdd-spec")

	content := s.EngramPhaseCostContent("my-project", "add-payment", "test-session-5")

	for _, required := range []string{
		"project: my-project",
		"change: add-payment",
		"session_id: test-session-5",
		"total_tokens: 3800",
		"sdd-spec:",
	} {
		if !strings.Contains(content, required) {
			t.Errorf("EngramPhaseCostContent missing %q\nContent:\n%s", required, content)
		}
	}
}

func TestAdd_ConcurrentSafe(t *testing.T) {
	s := NewSessionStats("agent", "concurrent-test")
	s.RecordPhaseStart("sdd-apply")

	const goroutines = 50
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.Add(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 100, CompletionTokens: 20})
		}()
	}
	wg.Wait()

	if s.TotalTokens() != int64(goroutines)*120 {
		t.Errorf("concurrent Add: got %d, want %d", s.TotalTokens(), int64(goroutines)*120)
	}
	bd := s.PhaseBreakdown()
	if bd["sdd-apply"] != int64(goroutines)*120 {
		t.Errorf("phase breakdown concurrent: got %d, want %d", bd["sdd-apply"], int64(goroutines)*120)
	}
}

func TestHookPhaseStart_NoopWhenNil(t *testing.T) {
	var h *Hook
	h.PhaseStart("sdd-verify")
	h.PhaseEnd("sdd-verify")
	bd := h.PhaseBreakdown()
	if bd != nil {
		t.Error("nil hook PhaseBreakdown should return nil")
	}
}

func TestHookPhaseStart_Works(t *testing.T) {
	h := &Hook{
		stats: NewSessionStats("sdd-orchestrator", "hook-test-1"),
	}
	h.PhaseStart("sdd-spec")
	h.Record(UsageDelta{Model: "claude-sonnet-4", PromptTokens: 2000})
	h.PhaseEnd("sdd-spec")

	bd := h.PhaseBreakdown()
	if bd["sdd-spec"] != 2000 {
		t.Errorf("hook phase breakdown: got %d, want 2000", bd["sdd-spec"])
	}
}
