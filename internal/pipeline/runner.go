package pipeline

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

// Runner executes a list of steps for a given stage.
type Runner struct {
	FailurePolicy FailurePolicy
	OnProgress    ProgressFunc

	// Hardening defaults
	StepTimeout  time.Duration
	MaxRetries   int
	RetryBackoff time.Duration
}

// Run executes steps sequentially, one after another. This is the original
// behavior preserved for backward compatibility.
// Deprecated: Use RunGroups for parallel execution within groups.
func (r Runner) Run(stage Stage, steps []Step) StageResult {
	if r.StepTimeout == 0 {
		r.StepTimeout = 5 * time.Minute
	}
	if r.MaxRetries < 0 {
		r.MaxRetries = 0
	}
	if r.RetryBackoff == 0 {
		r.RetryBackoff = 2 * time.Second
	}

	result := StageResult{Stage: stage, Success: true, Steps: make([]StepResult, 0, len(steps))}
	var errs []error

	for _, step := range steps {
		r.emitProgress(ProgressEvent{
			StepID: step.ID(),
			Stage:  stage,
			Status: StepStatusRunning,
		})

		var lastErr error
		var stepResult StepResult

		// Retry loop
		for attempt := 0; attempt <= r.MaxRetries; attempt++ {
			if attempt > 0 {
				time.Sleep(r.RetryBackoff)
				r.emitProgress(ProgressEvent{
					StepID: step.ID(),
					Stage:  stage,
					Status: StepStatusRunning,
					Notes:  fmt.Sprintf("Retry attempt %d", attempt),
				})
			}

			started := time.Now().UTC()

			// Simple timeout channel pattern
			done := make(chan error, 1)
			go func() {
				done <- step.Run()
			}()

			select {
			case err := <-done:
				lastErr = err
			case <-time.After(r.StepTimeout):
				lastErr = fmt.Errorf("step timed out after %v", r.StepTimeout)
			}

			finished := time.Now().UTC()
			stepResult = StepResult{
				StepID:     step.ID(),
				StartedAt:  started,
				FinishedAt: finished,
			}

			if lastErr == nil {
				break
			}
		}

		if lastErr != nil {
			stepResult.Status = StepStatusFailed
			stepResult.Err = lastErr
			result.Steps = append(result.Steps, stepResult)

			r.emitProgress(ProgressEvent{StepID: step.ID(), Stage: stage, Status: StepStatusFailed, Err: lastErr})

			errs = append(errs, lastErr)
			result.Success = false

			if r.FailurePolicy == StopOnError {
				result.Err = lastErr
				return result
			}

			continue
		}

		stepResult.Status = StepStatusSucceeded
		result.Steps = append(result.Steps, stepResult)
		r.emitProgress(ProgressEvent{StepID: step.ID(), Stage: stage, Status: StepStatusSucceeded})
	}

	if len(errs) > 0 {
		result.Err = errors.Join(errs...)
	}

	return result
}

// RunGroups executes groups sequentially; steps within each group run in parallel.
// Groups run in order (group[0] → group[1] → ...). On StopOnError, execution
// stops after the first group that contains a failed step. On ContinueOnError,
// all groups run regardless of failures and all errors are aggregated.
func (r Runner) RunGroups(ctx context.Context, stage Stage, groups []StepGroup) StageResult {
	r.applyDefaults()

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

// RunGroup executes all steps in a StepGroup concurrently using errgroup.
// Returns when ALL steps complete (success or failure).
// Respects FailurePolicy: ContinueOnError collects all errors;
// StopOnError cancels remaining steps via context.
func (r Runner) RunGroup(ctx context.Context, stage Stage, group StepGroup) StageResult {
	r.applyDefaults()

	result := StageResult{
		Stage:   stage,
		Steps:   make([]StepResult, len(group.Steps)),
		Success: true,
	}

	var mu sync.Mutex
	var errs []error

	g, gctx := errgroup.WithContext(ctx)

	for i, step := range group.Steps {
		i, step := i, step // capture loop variables

		r.emitProgress(ProgressEvent{
			StepID:      step.ID(),
			Stage:       stage,
			Status:      StepStatusRunning,
			ParallelIdx: i,
			GroupSize:   len(group.Steps),
		})

		g.Go(func() error {
			sr := r.executeStepWithRetry(gctx, stage, step)

			mu.Lock()
			result.Steps[i] = sr
			if sr.Status == StepStatusFailed {
				errs = append(errs, sr.Err)
				result.Success = false
			}
			mu.Unlock()

			r.emitProgress(ProgressEvent{
				StepID:      step.ID(),
				Stage:       stage,
				Status:      sr.Status,
				Err:         sr.Err,
				ParallelIdx: i,
				GroupSize:   len(group.Steps),
			})

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

// executeStepWithRetry runs a single step with retry and timeout logic.
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
			return StepResult{
				StepID:     step.ID(),
				Status:     StepStatusSucceeded,
				StartedAt:  started,
				FinishedAt: finished,
			}
		}
		// Last attempt falls through to failure
		if attempt == r.MaxRetries {
			return StepResult{
				StepID:     step.ID(),
				Status:     StepStatusFailed,
				Err:        err,
				StartedAt:  started,
				FinishedAt: finished,
			}
		}

		r.emitProgress(ProgressEvent{
			StepID: step.ID(),
			Stage:  stage,
			Status: StepStatusRunning,
			Notes:  fmt.Sprintf("Retry attempt %d", attempt+1),
		})
	}
	// Unreachable but satisfies compiler
	return StepResult{StepID: step.ID(), Status: StepStatusFailed}
}

// applyDefaults fills in zero-valued fields with sensible defaults.
func (r *Runner) applyDefaults() {
	if r.StepTimeout == 0 {
		r.StepTimeout = 5 * time.Minute
	}
	if r.MaxRetries < 0 {
		r.MaxRetries = 0
	}
	if r.RetryBackoff == 0 {
		r.RetryBackoff = 2 * time.Second
	}
}

func (r Runner) emitProgress(event ProgressEvent) {
	if r.OnProgress != nil {
		r.OnProgress(event)
	}
}