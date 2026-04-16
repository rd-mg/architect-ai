package pipeline

import (
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
}

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
		r.emitProgress(ProgressEvent{StepID: step.ID(), Stage: stage, Status: StepStatusRunning})

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

func (r Runner) emitProgress(event ProgressEvent) {
	if r.OnProgress != nil {
		r.OnProgress(event)
	}
}
