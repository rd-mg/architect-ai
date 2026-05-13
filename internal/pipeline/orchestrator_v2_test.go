package pipeline

import (
	"context"
	"testing"
	"time"
)

func TestRunnerNoncePropagation(t *testing.T) {
	events := []ProgressEvent{}
	nonce := "session-uuid-123"
	runner := Runner{
		Nonce: nonce,
		OnProgress: func(e ProgressEvent) {
			events = append(events, e)
		},
	}

	steps := []Step{newTestStep("step-1", &[]string{})}
	runner.Run(context.Background(), StageApply, steps)

	if len(events) == 0 {
		t.Fatal("expected events to be emitted")
	}

	for _, e := range events {
		if e.Nonce != nonce {
			t.Errorf("event.Nonce = %q, want %q", e.Nonce, nonce)
		}
	}
}

func TestRunnerWatchdogTimeout(t *testing.T) {
	events := []ProgressEvent{}
	runner := Runner{
		StepTimeout: 50 * time.Millisecond,
		OnProgress: func(e ProgressEvent) {
			events = append(events, e)
		},
	}

	// Step that takes longer than StepTimeout
	slowStep := &timeoutTestStep{id: "slow-step", duration: 200 * time.Millisecond}
	result := runner.Run(context.Background(), StageApply, []Step{slowStep})

	if result.Success {
		t.Error("expected result.Success to be false")
	}

	lastStepResult := result.Steps[0]
	if lastStepResult.Status != StepStatusTerminated {
		t.Errorf("step status = %q, want %q", lastStepResult.Status, StepStatusTerminated)
	}

	// Verify terminal event was also marked as terminated
	lastEvent := events[len(events)-1]
	if lastEvent.Status != StepStatusTerminated {
		t.Errorf("last event status = %q, want %q", lastEvent.Status, StepStatusTerminated)
	}
}

func TestOrchestratorRecursiveReasoningGate(t *testing.T) {
	orchestrator := NewOrchestrator(DefaultRollbackPolicy())
	
	// Implementation fails quality check (e.g. too many defects)
	orchestrator.QualityCheck = func(res StageResult) bool {
		return false // Quality check failed
	}

	result := orchestrator.Execute(context.Background(), StagePlan{
		Apply: []Step{newTestStep("apply-1", &[]string{})},
	})

	if result.NextRecommendedStage != "design" {
		t.Errorf("NextRecommendedStage = %q, want %q", result.NextRecommendedStage, "design")
	}

	if result.Err == nil {
		t.Error("expected error due to quality check failure")
	}
}

type timeoutTestStep struct {
	id       string
	duration time.Duration
}

func (s *timeoutTestStep) ID() string { return s.id }
func (s *timeoutTestStep) Run(ctx context.Context) error {
	select {
	case <-time.After(s.duration):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
