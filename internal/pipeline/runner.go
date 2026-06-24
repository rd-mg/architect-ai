package pipeline

import (
	"context"
	"errors"
	"fmt"
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

	Nonce string // Session-unique UUID for State-Synchronized DAG
}

func (r Runner) Run(ctx context.Context, stage Stage, steps []Step) StageResult {
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
		r.emitProgress(ProgressEvent{StepID: step.ID(), Stage: stage, Status: StepStatusRunning})

		var lastErr error
		var stepResult StepResult
		var status StepStatus = StepStatusFailed

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
				done <- step.Run(ctx)
			}()

			select {
			case err := <-done:
				lastErr = err
				status = StepStatusFailed
				if err == nil {
					status = StepStatusSucceeded
				}
			case <-ctx.Done():
				lastErr = ctx.Err()
				status = StepStatusInterrupted
			case <-time.After(r.StepTimeout):
				lastErr = fmt.Errorf("step timed out after %v", r.StepTimeout)
				status = StepStatusTerminated
			}

			finished := time.Now().UTC()
			stepResult = StepResult{
				StepID:     step.ID(),
				StartedAt:  started,
				FinishedAt: finished,
				Status:     status,
				Err:        lastErr,
			}

			if lastErr == nil {
				break
			}
		}

		if lastErr != nil {
			result.Steps = append(result.Steps, stepResult)

			r.emitProgress(ProgressEvent{StepID: step.ID(), Stage: stage, Status: status, Err: lastErr})

			errs = append(errs, lastErr)
			result.Success = false

			if r.FailurePolicy == StopOnError {
				result.Err = lastErr
				return result
			}

			continue
		}

		result.Steps = append(result.Steps, stepResult)
		r.emitProgress(ProgressEvent{StepID: step.ID(), Stage: stage, Status: StepStatusSucceeded})
	}

	if len(errs) > 0 {
		result.Err = errors.Join(errs...)
	}

	return result
}

func (r Runner) emitProgress(event ProgressEvent) {
	if r.OnProgress != nil {
		event.Nonce = r.Nonce
		r.OnProgress(event)
	}
}

// GroupMode controls whether steps in a StepGroup run sequentially or in parallel.
type GroupMode int

const (
	Sequential GroupMode = iota
	Parallel
)

// StepGroup bundles steps with an execution mode.
type StepGroup struct {
	Steps []Step
	Mode  GroupMode
}

// RunGroup executes steps sequentially or in parallel depending on Mode.
// Parallel mode uses errgroup to cancel remaining goroutines on first error.
// Call this AFTER the B6 mutex fix — parallel runners write shared state.
func (r Runner) RunGroup(ctx context.Context, stage Stage, g StepGroup) StageResult {
	if g.Mode == Sequential {
		return r.Run(ctx, stage, g.Steps)
	}

	if r.StepTimeout == 0 {
		r.StepTimeout = 5 * time.Minute
	}

	result := StageResult{Stage: stage, Success: true, Steps: make([]StepResult, 0, len(g.Steps))}

	eg, egCtx := errgroup.WithContext(ctx)

	for _, step := range g.Steps {
		step := step // capture loop variable
		eg.Go(func() error {
			r.emitProgress(ProgressEvent{StepID: step.ID(), Stage: stage, Status: StepStatusRunning})

			started := time.Now().UTC()
			var lastErr error
			var status StepStatus = StepStatusFailed

			done := make(chan error, 1)
			go func() {
				done <- step.Run(egCtx)
			}()

			select {
			case err := <-done:
				lastErr = err
				if err == nil {
					status = StepStatusSucceeded
				}
			case <-egCtx.Done():
				lastErr = egCtx.Err()
				status = StepStatusInterrupted
			case <-time.After(r.StepTimeout):
				lastErr = fmt.Errorf("step timed out after %v", r.StepTimeout)
				status = StepStatusTerminated
			}

			finished := time.Now().UTC()
			stepResult := StepResult{
				StepID:     step.ID(),
				StartedAt:  started,
				FinishedAt: finished,
				Status:     status,
				Err:        lastErr,
			}

			result.Steps = append(result.Steps, stepResult)
			r.emitProgress(ProgressEvent{StepID: step.ID(), Stage: stage, Status: status, Err: lastErr})

			return lastErr
		})
	}

	if err := eg.Wait(); err != nil {
		result.Success = false
		result.Err = err
	}

	return result
}
