# Phase 1 — Go Pipeline Parallelization

> **Cognitive Mode**: +++Systemic +++Empirical +++Pragmatic  
> **CCLD Tag**: `[PHASE-1][PIPELINE][PARALLEL][GO]`  
> **Status**: BLOCKED until Phase 0 Task A completes  
> **Estimated Duration**: 2–3 sessions  
> **Depends On**: `audit/pipeline-sequential.md`, `audit/state-management.md`

---

## 1.1 Objective

Refactor `internal/pipeline/runner.go` and `internal/pipeline/orchestrator.go` to execute **independent steps in parallel** using `golang.org/x/sync/errgroup`. Preserve the existing rollback model, failure policy, and progress callbacks. Add concurrent-safe state management.

**Target Outcome**: Install time for N independent components reduces from O(N × T) to O(max(T) + ε).

---

## 1.2 Root Cause (from Phase 0 Audit)

### 1.2.1 `Runner.Run()` — Sequential Loop (CONFIRMED BOTTLENECK)

```go
// internal/pipeline/runner.go — current
func (r Runner) Run(stage Stage, steps []Step) StageResult {
    for _, step := range steps {           // ← SEQUENTIAL. No fan-out.
        r.emitProgress(...)
        // retry loop
        done := make(chan error, 1)
        go func() { done <- step.Run() }() // ← goroutine per step but AWAITED immediately
        select {
        case err := <-done: ...
        case <-time.After(r.StepTimeout): ...
        }
    }
}
```

**Problem**: Each step is spawned as a goroutine but the outer loop blocks on the `select` before starting the next step. Net effect: strictly sequential.

### 1.2.2 `StagePlan` — No Independent-Step Metadata

```go
// internal/pipeline/stages.go
type StagePlan struct {
    Prepare []Step
    Apply   []Step
}
```

**Problem**: No way for the runner to know which Apply steps can execute in parallel vs which have file-system write conflicts. All steps treated identically.

### 1.2.3 `Orchestrator.Execute()` — No Parallel Dispatch

```go
func (o *Orchestrator) Execute(plan StagePlan) ExecutionResult {
    prepareResult := o.runner.Run(StagePrepare, plan.Prepare) // sequential
    applyResult   := o.runner.Run(StageApply, plan.Apply)     // sequential
    ...
}
```

**Problem**: Even if Prepare and Apply contain independent sub-groups, they run as one flat sequential list.

---

## 1.3 Refactoring Plan

### 1.3.1 New Type: `StepGroup`

Introduce `StepGroup` to annotate which steps are safe to run concurrently.

```go
// internal/pipeline/stages.go — AFTER

// StepGroup is a set of steps that can execute concurrently.
// All steps in a group must be free of write conflicts with each other.
// Groups themselves are executed sequentially (group[0] → group[1] → ...).
type StepGroup struct {
    Steps []Step
}

// StagePlan now contains ordered groups, not flat steps.
// Backward compatibility: builders can use SingleGroup() to wrap flat slices.
type StagePlan struct {
    Prepare []StepGroup // groups executed in order; within group, steps are parallel
    Apply   []StepGroup
}

// SingleGroup wraps a flat slice into a single StepGroup (sequential fallback).
func SingleGroup(steps ...Step) StepGroup {
    return StepGroup{Steps: steps}
}
```

**Migration**: All existing `BuildRealStagePlan` callers use `SingleGroup(existingSteps...)` initially. No behavioral change until groups are explicitly split.

---

### 1.3.2 New `Runner.RunGroup()` — Parallel Fan-Out

```go
// internal/pipeline/runner.go — NEW METHOD

import "golang.org/x/sync/errgroup"

// RunGroup executes all steps in a StepGroup concurrently.
// Returns when ALL steps complete (success or failure).
// Respects FailurePolicy: ContinueOnError collects all errors;
// StopOnError cancels remaining steps via context.
func (r Runner) RunGroup(ctx context.Context, stage Stage, group StepGroup) StageResult {
    result := StageResult{Stage: stage, Steps: make([]StepResult, len(group.Steps)), Success: true}
    
    type indexedResult struct {
        idx    int
        result StepResult
    }
    
    var mu sync.Mutex
    var errs []error
    
    g, gctx := errgroup.WithContext(ctx)
    
    for i, step := range group.Steps {
        i, step := i, step // capture
        r.emitProgress(ProgressEvent{StepID: step.ID(), Stage: stage, Status: StepStatusRunning})
        
        g.Go(func() error {
            sr := r.executeStepWithRetry(gctx, stage, step)
            
            mu.Lock()
            result.Steps[i] = sr
            if sr.Status == StepStatusFailed {
                errs = append(errs, sr.Err)
                result.Success = false
            }
            mu.Unlock()
            
            r.emitProgress(ProgressEvent{StepID: step.ID(), Stage: stage, Status: sr.Status, Err: sr.Err})
            
            if r.FailurePolicy == StopOnError && sr.Status == StepStatusFailed {
                return sr.Err // signals errgroup to cancel gctx
            }
            return nil
        })
    }
    
    _ = g.Wait()
    
    if len(errs) > 0 {
        result.Err = errors.Join(errs...)
    }
    return result
}

// executeStepWithRetry extracts the retry + timeout logic from Run().
// Accepts context for cancellation (parallel fan-out requires ctx propagation).
func (r Runner) executeStepWithRetry(ctx context.Context, stage Stage, step Step) StepResult {
    for attempt := 0; attempt <= r.MaxRetries; attempt++ {
        if attempt > 0 {
            select {
            case <-ctx.Done():
                return StepResult{StepID: step.ID(), Status: StepStatusFailed, Err: ctx.Err()}
            case <-time.After(r.RetryBackoff):
            }
        }
        
        started := time.Now().UTC()
        done := make(chan error, 1)
        go func() { done <- step.Run() }()
        
        var err error
        select {
        case err = <-done:
        case <-time.After(r.StepTimeout):
            err = fmt.Errorf("step %q timed out after %v", step.ID(), r.StepTimeout)
        case <-ctx.Done():
            err = ctx.Err()
        }
        
        finished := time.Now().UTC()
        if err == nil {
            return StepResult{StepID: step.ID(), Status: StepStatusSucceeded, StartedAt: started, FinishedAt: finished}
        }
        // last attempt falls through to return failure
        if attempt == r.MaxRetries {
            return StepResult{StepID: step.ID(), Status: StepStatusFailed, Err: err, StartedAt: started, FinishedAt: finished}
        }
    }
    // unreachable but satisfies compiler
    return StepResult{StepID: step.ID(), Status: StepStatusFailed}
}

// Run remains for backward compatibility. Wraps steps in SingleGroup.
func (r Runner) Run(stage Stage, steps []Step) StageResult {
    // deprecated: use RunGroups. Preserved for existing callers.
    return r.RunGroups(context.Background(), stage, []StepGroup{SingleGroup(steps...)})
}

// RunGroups executes groups sequentially; steps within each group in parallel.
func (r Runner) RunGroups(ctx context.Context, stage Stage, groups []StepGroup) StageResult {
    combined := StageResult{Stage: stage, Success: true}
    for _, group := range groups {
        gr := r.RunGroup(ctx, stage, group)
        combined.Steps = append(combined.Steps, gr.Steps...)
        if !gr.Success {
            combined.Success = false
            combined.Err = gr.Err
            if r.FailurePolicy == StopOnError {
                return combined
            }
        }
    }
    return combined
}
```

---

### 1.3.3 Updated `Orchestrator.Execute()` — Group-Aware

```go
// internal/pipeline/orchestrator.go — AFTER

func (o *Orchestrator) Execute(plan StagePlan) ExecutionResult {
    ctx := context.Background()
    
    o.indexGroupSteps(plan.Prepare)
    o.indexGroupSteps(plan.Apply)
    
    prepareResult := o.runner.RunGroups(ctx, StagePrepare, plan.Prepare)
    if !prepareResult.Success {
        return ExecutionResult{Prepare: prepareResult, Err: prepareResult.Err}
    }
    
    applyResult := o.runner.RunGroups(ctx, StageApply, plan.Apply)
    result := ExecutionResult{Prepare: prepareResult, Apply: applyResult}
    if applyResult.Success {
        return result
    }
    
    result.Err = applyResult.Err
    if o.policy.ShouldRollback(StageApply, applyResult.Err) {
        result.Rollback = ExecuteRollback(applyResult.Steps, o.stepByID)
        if !result.Rollback.Success {
            result.Err = result.Rollback.Err
        }
    }
    return result
}

func (o *Orchestrator) indexGroupSteps(groups []StepGroup) {
    for _, group := range groups {
        for _, step := range group.Steps {
            o.stepByID[step.ID()] = step
        }
    }
}
```

---

### 1.3.4 `BuildRealStagePlan` — Explicit Parallelism Groups

Identify which components have write conflicts:

**Independent (can run in parallel)**:
- `ComponentEngram` — writes to `~/.gemini/` (or agent-specific paths)
- `ComponentPersona` — writes base system prompt file (agent-specific path)
- `ComponentContext7` — writes MCP config entry
- `ComponentPermission` — writes permissions block
- `ComponentGGA` — writes GGA config

**Sequential Dependencies**:
- `ComponentSDD` must follow `ComponentEngram` (SDD appends to Engram-written files)
- `ComponentSkills` must follow `ComponentSDD`

**Proposed Group Layout** (post soft-ordering: Persona first):

```go
// Group 1: No-dependency components (parallel)
group1 := StepGroup{Steps: []Step{
    personaStep,     // must be first group due to soft-ordering
}}

// Group 2: Engram + independent components (parallel after Persona writes base)
group2 := StepGroup{Steps: []Step{
    engramStep,
    context7Step,
    permissionStep,
    ggaStep,
    themeStep,
}}

// Group 3: SDD (depends on Engram)
group3 := StepGroup{Steps: []Step{sddStep}}

// Group 4: Skills (depends on SDD)
group4 := StepGroup{Steps: []Step{skillsStep}}
```

This reduces 6-step install from T_engram + T_sdd + T_skills + T_context7 + T_persona + T_perms to max(T_persona) + max(T_engram, T_context7, T_perms, T_gga) + T_sdd + T_skills.

Empirical estimate: 4 serial steps → 2 serial + 1 parallel batch = ~40–60% latency reduction.

---

### 1.3.5 Progress Callback — Concurrent-Safe

The existing `OnProgress ProgressFunc` is called from within goroutines. Callers (TUI) must be ready for concurrent calls.

```go
// internal/pipeline/stages.go — ADD
// ProgressEvent now includes ParallelBatch metadata
type ProgressEvent struct {
    StepID      string
    Stage       Stage
    Status      StepStatus
    Err         error
    Notes       string
    ParallelIdx int // index within concurrent group (0 = sequential)
    GroupSize   int // total steps in this group
}
```

TUI model must serialize progress events via its existing `tea.Program` message channel — no direct state mutation from pipeline goroutines.

---

## 1.4 Rollback Compatibility

The rollback system iterates `steps []StepResult` in reverse and calls `Rollback()` on succeeded steps. This is compatible with parallel execution because:
1. `StepResult` slice is pre-allocated by index (`result.Steps[i] = sr`).
2. Rollback order is by result index, not execution order — safe.
3. Steps that are rolled back concurrently would require a parallel rollback runner (out of scope for Phase 1; sequential rollback is safe and conservative).

**Decision**: Keep rollback sequential. Add TODO comment for Phase 1.5 parallel rollback.

---

## 1.5 Testing Requirements

### New Tests

| Test | Location | Verifies |
|---|---|---|
| `TestRunnerRunGroupParallelism` | `pipeline/runner_test.go` | N steps complete faster than N×T when parallel |
| `TestRunnerRunGroupStopOnError` | `pipeline/runner_test.go` | StopOnError cancels remaining goroutines |
| `TestRunnerRunGroupContinueOnError` | `pipeline/runner_test.go` | ContinueOnError collects all errors |
| `TestRunnerRunGroupTimeoutPerStep` | `pipeline/runner_test.go` | Step timeout fires per-step, not per-group |
| `TestOrchestratorExecuteGroups` | `pipeline/orchestrator_test.go` | Groups execute in order; within group, parallel |
| `TestRunnerConcurrentProgressEvents` | `pipeline/runner_test.go` | Progress callback fires from multiple goroutines without data race |

### Existing Test Preservation

All existing `runner_test.go` tests must pass unchanged — `Run()` backward-compat wrapper ensures this.

### Race Detector

```bash
go test -race ./internal/pipeline/...
go test -race ./internal/app/...
```

Must show zero races before merge.

---

## 1.6 Files to Create / Modify

| File | Action | Notes |
|---|---|---|
| `internal/pipeline/stages.go` | MODIFY | Add `StepGroup`, `SingleGroup`, updated `StagePlan` |
| `internal/pipeline/runner.go` | MODIFY | Add `RunGroup`, `RunGroups`, `executeStepWithRetry`; update `Run` as compat wrapper |
| `internal/pipeline/orchestrator.go` | MODIFY | Add `indexGroupSteps`; update `Execute` to use `RunGroups` |
| `internal/pipeline/runner_test.go` | CREATE | New parallel tests |
| `internal/pipeline/orchestrator_test.go` | MODIFY | Group-aware tests |
| `internal/cli/install.go` | MODIFY | `BuildRealStagePlan` returns grouped plan |
| `go.mod` | MODIFY | Add `golang.org/x/sync` if not already present |

---

## 1.7 Acceptance Criteria

- [ ] `Runner.Run()` is backward-compatible (all existing tests pass)
- [ ] `RunGroup()` executes N independent steps in O(max(T)) not O(N×T) — proven by timing test
- [ ] `-race` detector shows zero races on pipeline package
- [ ] TUI progress reporting handles concurrent events without panic
- [ ] Rollback on partial parallel failure works correctly (all succeeded steps rolled back in reverse order)
- [ ] `BuildRealStagePlan` produces correct group topology for MVP component set

---

## 1.8 Sub-Agent Delegation

```
[PHASE-1 ORCHESTRATOR]
    │
    ├── [1A] go-writer-agent     → stages.go StepGroup types
    ├── [1B] go-writer-agent     → runner.go RunGroup/RunGroups/executeStepWithRetry
    ├── [1C] go-writer-agent     → orchestrator.go Execute group-aware (depends 1A+1B)
    ├── [1D] go-writer-agent     → BuildRealStagePlan groups (depends 1C)
    └── [1E] go-tester-agent     → all new tests (depends 1A-1D)
```

1A, 1B can launch in parallel.  
1C launches after 1A+1B.  
1D launches after 1C.  
1E launches after all writers.
