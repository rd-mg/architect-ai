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

var (
	preTaskHooks  []PreTaskHook
	postTaskHooks []PostTaskHook
	mu            sync.RWMutex
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
}
