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

	RegisterPreTask(func(ctx context.Context, task string) {
		if task == "test-task" {
			preCalled = true
		}
	})

	RegisterPostTask(func(ctx context.Context, task string, err error) {
		if task == "test-task" {
			postCalled = true
			capturedErr = err
		}
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

	RegisterPreTask(func(ctx context.Context, task string) {
		panic("at the disco")
	})

	// This should not panic the test
	FirePreTask(context.Background(), "panic-task")
}
