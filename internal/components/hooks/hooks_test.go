package hooks

import (
	"context"
	"errors"
	"testing"
)

func TestHooks(t *testing.T) {
	Reset()
	defer Reset()

	var preCalled, postCalled bool
	var capturedErr error

	RegisterPreTask(func(ctx context.Context, task string) error {
		if task == "test-task" {
			preCalled = true
		}
		return nil
	})

	RegisterPostTask(func(ctx context.Context, task string, err error) error {
		if task == "test-task" {
			postCalled = true
			capturedErr = err
		}
		return nil
	})

	ctx := context.Background()
	testErr := errors.New("boom")

	FirePreTask(ctx, "test-task")
	FirePostTask(ctx, "test-task", testErr)

	if !preCalled {
		t.Error("PreTask hook was not called")
	}
	if !postCalled {
		t.Error("PostTask hook was not called")
	}
	if capturedErr != testErr {
		t.Errorf("expected error %v, got %v", testErr, capturedErr)
	}
}

func TestHookPanicRecovery(t *testing.T) {
	Reset()
	defer Reset()

	RegisterPreTask(func(ctx context.Context, task string) error {
		panic("at the disco")
	})

	// This should not panic the test
	FirePreTask(context.Background(), "panic-task")
}

func TestFirePrePhase_CallsAllHooks(t *testing.T) {
	mu.Lock()
	prePhaseHooks = nil
	mu.Unlock()

	var called []string
	RegisterPrePhase(func(ctx context.Context, phaseName, changeID string) error {
		called = append(called, phaseName+"/"+changeID)
		return nil
	})

	outcomes := FirePrePhase(context.Background(), "sdd-apply", "my-change")
	if len(outcomes) != 1 {
		t.Fatalf("expected 1 outcome, got %d", len(outcomes))
	}
	if outcomes[0].Status != HookSuccess {
		t.Errorf("expected success, got %s: %s", outcomes[0].Status, outcomes[0].Error)
	}
	if len(called) != 1 || called[0] != "sdd-apply/my-change" {
		t.Errorf("unexpected call args: %v", called)
	}
}

func TestFirePostPhase_PropagatesPhaseError(t *testing.T) {
	mu.Lock()
	postPhaseHooks = nil
	mu.Unlock()

	wantErr := errors.New("sdd-apply failed: compile error")
	var gotErr error
	RegisterPostPhase(func(ctx context.Context, phaseName, changeID string, phaseErr error) error {
		gotErr = phaseErr
		return nil
	})

	FirePostPhase(context.Background(), "sdd-apply", "my-change", wantErr)

	if gotErr != wantErr {
		t.Errorf("expected phaseErr to propagate, got %v", gotErr)
	}
}

func TestFirePrePhase_NoHooks_ReturnsEmpty(t *testing.T) {
	mu.Lock()
	prePhaseHooks = nil
	mu.Unlock()

	outcomes := FirePrePhase(context.Background(), "sdd-verify", "change-x")
	if len(outcomes) != 0 {
		t.Errorf("expected 0 outcomes with no hooks, got %d", len(outcomes))
	}
}

func TestFirePostPhase_PanicRecovered(t *testing.T) {
	mu.Lock()
	postPhaseHooks = nil
	mu.Unlock()

	RegisterPostPhase(func(ctx context.Context, phaseName, changeID string, phaseErr error) error {
		panic("hook panicked")
	})

	outcomes := FirePostPhase(context.Background(), "sdd-apply", "c", nil)
	if len(outcomes) != 1 {
		t.Fatalf("expected 1 outcome")
	}
	if outcomes[0].Status != HookPanic {
		t.Errorf("expected panic status, got %s", outcomes[0].Status)
	}
}
