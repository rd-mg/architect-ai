package pipeline

import (
	"context"
	"errors"
	"fmt"
	"time"
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
