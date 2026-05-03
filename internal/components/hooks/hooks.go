package hooks

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type PreTaskHook func(ctx context.Context, task string) error
type PostTaskHook func(ctx context.Context, task string, err error) error

// PrePhaseHook fires before the orchestrator delegates an SDD phase.
// phaseName is the SDD phase name (e.g. "sdd-apply").
// changeID is the active change name.
type PrePhaseHook func(ctx context.Context, phaseName, changeID string) error

// PostPhaseHook fires after an SDD phase sub-agent returns.
// phaseErr is nil on success; non-nil if the phase returned a blocker.
type PostPhaseHook func(ctx context.Context, phaseName, changeID string, phaseErr error) error

var (
	preTaskHooks   []PreTaskHook
	postTaskHooks  []PostTaskHook
	prePhaseHooks  []PrePhaseHook
	postPhaseHooks []PostPhaseHook
	mu             sync.RWMutex
)

// RegisterPreTask adds a hook to be fired before a task starts.
func RegisterPreTask(fn PreTaskHook) {
	mu.Lock()
	defer mu.Unlock()
	preTaskHooks = append(preTaskHooks, fn)
}

// RegisterPostTask adds a hook to be fired after a task completes.
func RegisterPostTask(fn PostTaskHook) {
	mu.Lock()
	defer mu.Unlock()
	postTaskHooks = append(postTaskHooks, fn)
}

// RegisterPrePhase adds a hook fired before each SDD phase delegation.
func RegisterPrePhase(fn PrePhaseHook) {
	mu.Lock()
	defer mu.Unlock()
	prePhaseHooks = append(prePhaseHooks, fn)
}

// RegisterPostPhase adds a hook fired after each SDD phase completes.
func RegisterPostPhase(fn PostPhaseHook) {
	mu.Lock()
	defer mu.Unlock()
	postPhaseHooks = append(postPhaseHooks, fn)
}

type OutcomeStatus string

const (
	HookSuccess OutcomeStatus = "success"
	HookError   OutcomeStatus = "error"
	HookPanic   OutcomeStatus = "panic"
	HookTimeout OutcomeStatus = "timeout"
)

type Outcome struct {
	Name     string
	Stage    string
	Status   OutcomeStatus
	Duration time.Duration
	Error    string
}

// FirePreTask executes all registered pre-task hooks safely and returns outcomes.
func FirePreTask(ctx context.Context, task string) []Outcome {
	mu.RLock()
	hooks := make([]PreTaskHook, len(preTaskHooks))
	copy(hooks, preTaskHooks)
	mu.RUnlock()

	outcomes := make([]Outcome, 0, len(hooks))
	for i, fn := range hooks {
		name := fmt.Sprintf("pre-task-%d", i)
		outcomes = append(outcomes, runHook(ctx, name, "pre", func(ctx context.Context) error {
			return fn(ctx, task)
		}))
	}
	return outcomes
}

// FirePostTask executes all registered post-task hooks safely and returns outcomes.
func FirePostTask(ctx context.Context, task string, err error) []Outcome {
	mu.RLock()
	hooks := make([]PostTaskHook, len(postTaskHooks))
	copy(hooks, postTaskHooks)
	mu.RUnlock()

	outcomes := make([]Outcome, 0, len(hooks))
	for i, fn := range hooks {
		name := fmt.Sprintf("post-task-%d", i)
		outcomes = append(outcomes, runHook(ctx, name, "post", func(ctx context.Context) error {
			return fn(ctx, task, err)
		}))
	}
	return outcomes
}

// FirePrePhase executes all registered pre-phase hooks safely.
func FirePrePhase(ctx context.Context, phaseName, changeID string) []Outcome {
	mu.RLock()
	hooks := make([]PrePhaseHook, len(prePhaseHooks))
	copy(hooks, prePhaseHooks)
	mu.RUnlock()

	outcomes := make([]Outcome, 0, len(hooks))
	for i, fn := range hooks {
		fn := fn
		name := fmt.Sprintf("pre-phase-%d", i)
		outcomes = append(outcomes, runHook(ctx, name, "pre-phase", func(ctx context.Context) error {
			return fn(ctx, phaseName, changeID)
		}))
	}
	return outcomes
}

// FirePostPhase executes all registered post-phase hooks safely.
func FirePostPhase(ctx context.Context, phaseName, changeID string, phaseErr error) []Outcome {
	mu.RLock()
	hooks := make([]PostPhaseHook, len(postPhaseHooks))
	copy(hooks, postPhaseHooks)
	mu.RUnlock()

	outcomes := make([]Outcome, 0, len(hooks))
	for i, fn := range hooks {
		fn := fn
		name := fmt.Sprintf("post-phase-%d", i)
		outcomes = append(outcomes, runHook(ctx, name, "post-phase", func(ctx context.Context) error {
			return fn(ctx, phaseName, changeID, phaseErr)
		}))
	}
	return outcomes
}

func runHook(ctx context.Context, name, stage string, fn func(context.Context) error) Outcome {
	outcome := Outcome{
		Name:  name,
		Stage: stage,
	}

	timeout := 2 * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	start := time.Now()
	done := make(chan error, 1)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				done <- fmt.Errorf("panic: %v", r)
			}
		}()
		done <- fn(ctx)
	}()

	select {
	case err := <-done:
		outcome.Duration = time.Since(start)
		if err != nil {
			if strings.HasPrefix(err.Error(), "panic:") {
				outcome.Status = HookPanic
			} else {
				outcome.Status = HookError
			}
			outcome.Error = err.Error()
		} else {
			outcome.Status = HookSuccess
		}
	case <-ctx.Done():
		outcome.Duration = time.Since(start)
		outcome.Status = HookTimeout
		outcome.Error = "hook timed out"
	}

	return outcome
}

// Reset clears all registered hooks (primarily for tests).
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	preTaskHooks = nil
	postTaskHooks = nil
	prePhaseHooks = nil
	postPhaseHooks = nil
}
