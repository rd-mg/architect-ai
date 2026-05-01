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

	RegisterPreTask(func(ctx context.Context, task string) error {
		if task == "test-task" {
			preCalled = true
		}
		return nil
	})

	RegisterPostTask(func(ctx context.Context, task string, err error) error {
		if task == "test-task" {
			postCalled = true
			capturedErr = err
		}
		return nil
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

	RegisterPreTask(func(ctx context.Context, task string) error {
		panic("at the disco")
	})

	// This should not panic the test
	FirePreTask(context.Background(), "panic-task")
}
