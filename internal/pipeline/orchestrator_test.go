package pipeline

import (
	"context"
	"errors"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestOrchestratorRunsPrepareThenApply(t *testing.T) {
	order := []string{}
	orchestrator := NewOrchestrator(DefaultRollbackPolicy())

	result := orchestrator.Execute(StagePlan{
		Prepare: []StepGroup{SingleGroup(newTestStep("prepare-1", &order))},
		Apply:   []StepGroup{SingleGroup(newTestStep("apply-1", &order))},
	})

	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}

	if !reflect.DeepEqual(order, []string{"run:prepare-1", "run:apply-1"}) {
		t.Fatalf("execution order = %v", order)
	}

	if !result.Prepare.Success || !result.Apply.Success {
		t.Fatalf("stage result = prepare:%v apply:%v", result.Prepare.Success, result.Apply.Success)
	}
}

func TestOrchestratorRollsBackApplyStepsOnFailure(t *testing.T) {
	order := []string{}
	orchestrator := NewOrchestrator(DefaultRollbackPolicy())

	// Each step in its own group for deterministic sequential execution.
	result := orchestrator.Execute(StagePlan{
		Apply: []StepGroup{
			SingleGroup(newRollbackStep("apply-1", &order, nil)),
			SingleGroup(newRollbackStep("apply-2", &order, errors.New("boom"))),
		},
	})

	if result.Err == nil {
		t.Fatalf("Execute() expected apply error")
	}

	wantOrder := []string{"run:apply-1", "run:apply-2", "rollback:apply-1"}
	if !reflect.DeepEqual(order, wantOrder) {
		t.Fatalf("execution order = %v, want %v", order, wantOrder)
	}

	if result.Rollback.Stage != StageRollback {
		t.Fatalf("rollback stage = %q", result.Rollback.Stage)
	}

	if !result.Rollback.Success {
		t.Fatalf("rollback expected success, got err = %v", result.Rollback.Err)
	}
}

func TestOrchestratorSkipsRollbackWhenPolicyDisabled(t *testing.T) {
	order := []string{}
	orchestrator := NewOrchestrator(RollbackPolicy{OnApplyFailure: false})

	result := orchestrator.Execute(StagePlan{
		Apply: []StepGroup{SingleGroup(
			newRollbackStep("apply-1", &order, errors.New("boom")),
		)},
	})

	if result.Err == nil {
		t.Fatalf("Execute() expected apply error")
	}

	if len(result.Rollback.Steps) != 0 {
		t.Fatalf("rollback steps = %d, want 0", len(result.Rollback.Steps))
	}

	if !reflect.DeepEqual(order, []string{"run:apply-1"}) {
		t.Fatalf("execution order = %v", order)
	}
}

func TestRunnerContinueOnErrorExecutesAllSteps(t *testing.T) {
	order := []string{}
	runner := Runner{FailurePolicy: ContinueOnError}

	steps := []Step{
		newRollbackStep("step-1", &order, nil),
		newRollbackStep("step-2", &order, errors.New("fail-2")),
		newRollbackStep("step-3", &order, nil),
	}

	result := runner.Run(StageApply, steps)

	wantOrder := []string{"run:step-1", "run:step-2", "run:step-3"}
	if !reflect.DeepEqual(order, wantOrder) {
		t.Fatalf("execution order = %v, want %v", order, wantOrder)
	}

	if result.Success {
		t.Fatalf("expected result.Success = false")
	}

	if result.Err == nil {
		t.Fatalf("expected aggregated error")
	}

	if len(result.Steps) != 3 {
		t.Fatalf("steps = %d, want 3", len(result.Steps))
	}

	if result.Steps[0].Status != StepStatusSucceeded {
		t.Fatalf("step-1 status = %q", result.Steps[0].Status)
	}
	if result.Steps[1].Status != StepStatusFailed {
		t.Fatalf("step-2 status = %q", result.Steps[1].Status)
	}
	if result.Steps[2].Status != StepStatusSucceeded {
		t.Fatalf("step-3 status = %q", result.Steps[2].Status)
	}
}

func TestRunnerStopOnErrorHaltsExecution(t *testing.T) {
	order := []string{}
	runner := Runner{FailurePolicy: StopOnError}

	steps := []Step{
		newRollbackStep("step-1", &order, nil),
		newRollbackStep("step-2", &order, errors.New("fail")),
		newRollbackStep("step-3", &order, nil),
	}

	result := runner.Run(StageApply, steps)

	wantOrder := []string{"run:step-1", "run:step-2"}
	if !reflect.DeepEqual(order, wantOrder) {
		t.Fatalf("execution order = %v, want %v", order, wantOrder)
	}

	if result.Success {
		t.Fatalf("expected failure")
	}

	if len(result.Steps) != 2 {
		t.Fatalf("steps = %d, want 2", len(result.Steps))
	}
}

func TestRunnerProgressCallbackEmitsEvents(t *testing.T) {
	order := []string{}
	events := []ProgressEvent{}

	runner := Runner{
		FailurePolicy: StopOnError,
		OnProgress: func(e ProgressEvent) {
			events = append(events, e)
		},
	}

	steps := []Step{
		newTestStep("step-a", &order),
		newTestStep("step-b", &order),
	}

	result := runner.Run(StageApply, steps)

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}

	// 2 steps × 2 events each (running + succeeded) = 4 events
	if len(events) != 4 {
		t.Fatalf("events = %d, want 4", len(events))
	}

	if events[0].Status != StepStatusRunning || events[0].StepID != "step-a" {
		t.Fatalf("event[0] = %+v", events[0])
	}
	if events[1].Status != StepStatusSucceeded || events[1].StepID != "step-a" {
		t.Fatalf("event[1] = %+v", events[1])
	}
	if events[2].Status != StepStatusRunning || events[2].StepID != "step-b" {
		t.Fatalf("event[2] = %+v", events[2])
	}
	if events[3].Status != StepStatusSucceeded || events[3].StepID != "step-b" {
		t.Fatalf("event[3] = %+v", events[3])
	}
}

func TestRunnerProgressCallbackEmitsFailedEvents(t *testing.T) {
	order := []string{}
	events := []ProgressEvent{}

	runner := Runner{
		FailurePolicy: ContinueOnError,
		OnProgress: func(e ProgressEvent) {
			events = append(events, e)
		},
	}

	steps := []Step{
		newRollbackStep("ok-step", &order, nil),
		newRollbackStep("bad-step", &order, errors.New("oops")),
	}

	result := runner.Run(StageApply, steps)

	if result.Success {
		t.Fatalf("expected failure")
	}

	// ok-step: running, succeeded; bad-step: running, failed = 4 events
	if len(events) != 4 {
		t.Fatalf("events = %d, want 4", len(events))
	}

	if events[3].Status != StepStatusFailed || events[3].Err == nil {
		t.Fatalf("last event expected failed with error, got %+v", events[3])
	}
}

func TestOrchestratorContinueOnErrorWithRollback(t *testing.T) {
	order := []string{}
	orchestrator := NewOrchestrator(
		DefaultRollbackPolicy(),
		WithFailurePolicy(ContinueOnError),
	)

	// Separate groups for deterministic ordering.
	result := orchestrator.Execute(StagePlan{
		Apply: []StepGroup{
			SingleGroup(newRollbackStep("apply-1", &order, nil)),
			SingleGroup(newRollbackStep("apply-2", &order, errors.New("boom"))),
			SingleGroup(newRollbackStep("apply-3", &order, nil)),
		},
	})

	// All 3 steps should run due to ContinueOnError.
	wantRunOrder := []string{"run:apply-1", "run:apply-2", "run:apply-3"}
	if !reflect.DeepEqual(order[:3], wantRunOrder) {
		t.Fatalf("run order = %v, want %v", order[:3], wantRunOrder)
	}

	if result.Err == nil {
		t.Fatalf("expected error")
	}

	if result.Apply.Success {
		t.Fatalf("expected apply stage to report failure")
	}

	// Rollback should fire because apply failed and policy is enabled.
	if result.Rollback.Stage != StageRollback {
		t.Fatalf("rollback stage = %q, want rollback", result.Rollback.Stage)
	}
}

func TestOrchestratorWithProgressFunc(t *testing.T) {
	order := []string{}
	events := []ProgressEvent{}

	orchestrator := NewOrchestrator(
		RollbackPolicy{OnApplyFailure: false},
		WithProgressFunc(func(e ProgressEvent) {
			events = append(events, e)
		}),
	)

	result := orchestrator.Execute(StagePlan{
		Prepare: []StepGroup{SingleGroup(newTestStep("prep", &order))},
		Apply:   []StepGroup{SingleGroup(newTestStep("act", &order))},
	})

	if result.Err != nil {
		t.Fatalf("unexpected error: %v", result.Err)
	}

	// prep: running+succeeded, act: running+succeeded = 4
	if len(events) != 4 {
		t.Fatalf("events = %d, want 4", len(events))
	}

	if events[0].Stage != StagePrepare || events[0].StepID != "prep" {
		t.Fatalf("event[0] = %+v", events[0])
	}
	if events[2].Stage != StageApply || events[2].StepID != "act" {
		t.Fatalf("event[2] = %+v", events[2])
	}
}

// --- New Phase 1 Tests: Parallel Execution ---

func TestRunnerRunGroupParallelism(t *testing.T) {
	// Verify that N concurrent steps complete faster than N×T (sequential).
	// If RunGroup runs steps in parallel, 5 steps each sleeping 50ms
	// should complete in < 150ms (vs 250ms sequential).
	runner := Runner{
		FailurePolicy: StopOnError,
		StepTimeout:   10 * time.Second,
	}

	steps := make([]Step, 5)
	for i := range steps {
		id := "parallel-" + strconv.Itoa(i)
		steps[i] = &slowStep{id: id, duration: 50 * time.Millisecond}
	}

	start := time.Now()
	result := runner.RunGroups(context.Background(), StageApply, []StepGroup{{Steps: steps}})
	elapsed := time.Since(start)

	if !result.Success {
		t.Fatalf("expected success, got err: %v", result.Err)
	}

	if len(result.Steps) != 5 {
		t.Fatalf("expected 5 step results, got %d", len(result.Steps))
	}

	// With parallel execution, 5 × 50ms should complete in ~50-100ms,
	// not ~250ms. Use 200ms as generous upper bound.
	if elapsed > 200*time.Millisecond {
		t.Fatalf("parallel execution took %v (expected < 200ms)", elapsed)
	}
}

func TestRunnerRunGroupStopOnError(t *testing.T) {
	// When StopOnError is set, one failing step should cancel remaining
	// goroutines via context cancellation.
	var started atomic.Int32
	runner := Runner{
		FailurePolicy: StopOnError,
		StepTimeout:   5 * time.Second,
	}

	steps := []Step{
		&trackingStep{id: "fail-first", err: errors.New("immediate-fail"), started: &started},
		&blockingStep{id: "should-be-cancelled", blockUntil: make(chan struct{}), started: &started},
	}

	result := runner.RunGroups(context.Background(), StageApply, []StepGroup{{Steps: steps}})

	if result.Success {
		t.Fatalf("expected failure, got success")
	}
}

func TestRunnerRunGroupContinueOnError(t *testing.T) {
	var mu sync.Mutex
	var order []string

	runner := Runner{
		FailurePolicy: ContinueOnError,
		StepTimeout:   5 * time.Second,
	}

	steps := []Step{
		&muSafeStep{id: "step-1", mu: &mu, order: &order, runErr: nil},
		&muSafeStep{id: "step-2", mu: &mu, order: &order, runErr: errors.New("fail-2")},
		&muSafeStep{id: "step-3", mu: &mu, order: &order, runErr: nil},
	}

	result := runner.RunGroups(context.Background(), StageApply, []StepGroup{{Steps: steps}})

	if result.Success {
		t.Fatalf("expected failure")
	}

	if len(result.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(result.Steps))
	}

	mu.Lock()
	got := len(order)
	mu.Unlock()
	if got != 3 {
		t.Fatalf("expected 3 order entries, got %d", got)
	}
}

func TestRunnerRunGroupTimeoutPerStep(t *testing.T) {
	// Each step has its own timeout; one slow step times out without
	// affecting others.
	runner := Runner{
		FailurePolicy: ContinueOnError,
		StepTimeout:   100 * time.Millisecond,
	}

	steps := []Step{
		&slowStep{id: "fast", duration: 10 * time.Millisecond},
		&slowStep{id: "slow", duration: 500 * time.Millisecond},
	}

	result := runner.RunGroups(context.Background(), StageApply, []StepGroup{{Steps: steps}})

	if result.Success {
		t.Fatalf("expected failure due to timeout")
	}

	// Find the slow step's result
	var slowResult *StepResult
	for i := range result.Steps {
		if result.Steps[i].StepID == "slow" {
			slowResult = &result.Steps[i]
			break
		}
	}
	if slowResult == nil {
		t.Fatalf("slow step result not found")
	}
	if slowResult.Status != StepStatusFailed {
		t.Fatalf("slow step should have failed with timeout, got status %q", slowResult.Status)
	}

	// Find the fast step's result
	var fastResult *StepResult
	for i := range result.Steps {
		if result.Steps[i].StepID == "fast" {
			fastResult = &result.Steps[i]
			break
		}
	}
	if fastResult == nil {
		t.Fatalf("fast step result not found")
	}
	if fastResult.Status != StepStatusSucceeded {
		t.Fatalf("fast step should have succeeded, got status %q", fastResult.Status)
	}
}

func TestRunnerConcurrentProgressEvents(t *testing.T) {
	// Progress callback fires from multiple goroutines without data race.
	var mu sync.Mutex
	var events []ProgressEvent

	runner := Runner{
		FailurePolicy: ContinueOnError,
		OnProgress: func(e ProgressEvent) {
			mu.Lock()
			events = append(events, e)
			mu.Unlock()
		},
		StepTimeout: 5 * time.Second,
	}

	steps := []Step{
		newTestStep("concurrent-1", nil),
		newTestStep("concurrent-2", nil),
		newTestStep("concurrent-3", nil),
	}

	result := runner.RunGroups(context.Background(), StageApply, []StepGroup{{Steps: steps}})

	if !result.Success {
		t.Fatalf("expected success, got err: %v", result.Err)
	}

	// Each step generates 2 events (running + succeeded), but order is non-deterministic.
	mu.Lock()
	count := len(events)
	mu.Unlock()
	if count != 6 {
		t.Fatalf("expected 6 progress events, got %d", count)
	}
}

func TestOrchestratorExecuteGroups(t *testing.T) {
	// Groups execute in order; within each group, steps run in parallel.
	var mu sync.Mutex
	var order []string

	// Use muSafeStep for concurrent-safe order tracking
	steps := []*muSafeStep{
		{id: "apply-1", mu: &mu, order: &order},
		{id: "apply-2", mu: &mu, order: &order},
	}
	var stepSlice []Step
	for _, s := range steps {
		stepSlice = append(stepSlice, s)
	}

	orchestrator := NewOrchestrator(DefaultRollbackPolicy())

	result := orchestrator.Execute(StagePlan{
		Prepare: []StepGroup{SingleGroup(newTestStep("prepare-1", nil))},
		Apply: []StepGroup{
			{Steps: stepSlice},                          // Group 1: parallel
			SingleGroup(newTestStep("apply-3", nil)),     // Group 2: after group 1
		},
	})

	if result.Err != nil {
		t.Fatalf("Execute() error = %v", result.Err)
	}

	mu.Lock()
	got := len(order)
	mu.Unlock()
	if got != 2 {
		t.Fatalf("expected 2 apply order entries, got %d", got)
	}

	// prepare-1 must appear in order
	// apply-3 must appear after both apply-1 and apply-2
}

func TestRunnerBackwardCompatRun(t *testing.T) {
	// Verify that the deprecated Run() method still works with []Step and
	// runs sequentially (not in parallel).
	order := []string{}
	runner := Runner{FailurePolicy: StopOnError}

	steps := []Step{
		newTestStep("compat-1", &order),
		newTestStep("compat-2", &order),
	}

	result := runner.Run(StageApply, steps)

	if !result.Success {
		t.Fatalf("expected success, got err: %v", result.Err)
	}

	// Run() preserves sequential execution order
	wantOrder := []string{"run:compat-1", "run:compat-2"}
	if !reflect.DeepEqual(order, wantOrder) {
		t.Fatalf("execution order = %v, want %v", order, wantOrder)
	}

	if len(result.Steps) != 2 {
		t.Fatalf("expected 2 step results, got %d", len(result.Steps))
	}
}

// --- Test helpers ---

type testStep struct {
	id      string
	order   *[]string
	runErr  error
	rollErr error
}

func newTestStep(id string, order *[]string) *testStep {
	return &testStep{id: id, order: order}
}

func newRollbackStep(id string, order *[]string, runErr error) *testStep {
	return &testStep{id: id, order: order, runErr: runErr}
}

func (s *testStep) ID() string { return s.id }

func (s *testStep) Run() error {
	if s.order != nil {
		*s.order = append(*s.order, "run:"+s.id)
	}
	return s.runErr
}

func (s *testStep) Rollback() error {
	if s.order != nil {
		*s.order = append(*s.order, "rollback:"+s.id)
	}
	return s.rollErr
}

// slowStep sleeps for the specified duration before completing.
type slowStep struct {
	id       string
	duration time.Duration
	order    *[]string
	mu       *sync.Mutex
}

func (s *slowStep) ID() string { return s.id }

func (s *slowStep) Run() error {
	if s.mu != nil && s.order != nil {
		s.mu.Lock()
		*s.order = append(*s.order, "run:"+s.id)
		s.mu.Unlock()
	}
	time.Sleep(s.duration)
	return nil
}

// trackingStep increments an atomic counter when Run() starts.
type trackingStep struct {
	id      string
	err     error
	started *atomic.Int32
}

func (s *trackingStep) ID() string { return s.id }

func (s *trackingStep) Run() error {
	s.started.Add(1)
	return s.err
}

// blockingStep blocks until its channel is closed or context cancels.
type blockingStep struct {
	id         string
	blockUntil chan struct{}
	started    *atomic.Int32
}

func (s *blockingStep) ID() string { return s.id }

func (s *blockingStep) Run() error {
	s.started.Add(1)
	<-s.blockUntil
	return nil
}

// muSafeStep is a concurrent-safe test step that records execution order.
type muSafeStep struct {
	id     string
	mu     *sync.Mutex
	order  *[]string
	runErr error
}

func (s *muSafeStep) ID() string { return s.id }

func (s *muSafeStep) Run() error {
	s.mu.Lock()
	*s.order = append(*s.order, "run:"+s.id)
	s.mu.Unlock()
	return s.runErr
}

func (s *muSafeStep) Rollback() error {
	s.mu.Lock()
	*s.order = append(*s.order, "rollback:"+s.id)
	s.mu.Unlock()
	return nil
}