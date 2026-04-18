package hooks

import (
	"context"
	"sync"
)

type PreTaskHook func(ctx context.Context, task string)
type PostTaskHook func(ctx context.Context, task string, err error)

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

// FirePreTask executes all registered pre-task hooks safely.
func FirePreTask(ctx context.Context, task string) {
	mu.RLock()
	hooks := make([]PreTaskHook, len(preTaskHooks))
	copy(hooks, preTaskHooks)
	mu.RUnlock()

	for _, fn := range hooks {
		safeRun(func() { fn(ctx, task) })
	}
}

// FirePostTask executes all registered post-task hooks safely.
func FirePostTask(ctx context.Context, task string, err error) {
	mu.RLock()
	hooks := make([]PostTaskHook, len(postTaskHooks))
	copy(hooks, postTaskHooks)
	mu.RUnlock()

	for _, fn := range hooks {
		safeRun(func() { fn(ctx, task, err) })
	}
}

func safeRun(fn func()) {
	defer func() {
		_ = recover() // Observability hooks MUST NOT crash the main flow
	}()
	fn()
}

// Reset clears all registered hooks (primarily for tests).
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	preTaskHooks = nil
	postTaskHooks = nil
}
