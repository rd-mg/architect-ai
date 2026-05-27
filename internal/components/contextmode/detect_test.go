package contextmode_test

import (
	"context"
	"errors"
	"testing"

	contextmode "github.com/rd-mg/architect-ai/internal/components/contextmode"
	"github.com/rd-mg/architect-ai/internal/process"
)

func TestDetectNotInstalled(t *testing.T) {
	saved := process.Run
	t.Cleanup(func() { process.Run = saved })

	process.Run = func(_ context.Context, _ string, _ []string, _ process.Options) (process.Result, error) {
		return process.Result{Error: errors.New("not found")}, nil
	}

	cap, err := contextmode.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect returned unexpected error: %v", err)
	}
	if cap.Installed {
		t.Error("Expected Installed=false when process.Run fails")
	}
}

func TestDetectInstalled(t *testing.T) {
	saved := process.Run
	t.Cleanup(func() { process.Run = saved })

	process.Run = func(_ context.Context, _ string, _ []string, _ process.Options) (process.Result, error) {
		return process.Result{Stdout: []byte("context-mode v1.0.0")}, nil
	}

	cap, err := contextmode.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect returned unexpected error: %v", err)
	}
	if !cap.Installed {
		t.Error("Expected Installed=true when process.Run succeeds")
	}
	if cap.Version != "context-mode v1.0.0" {
		t.Errorf("Version = %q, want %q", cap.Version, "context-mode v1.0.0")
	}
	if !cap.DoctorAvailable {
		t.Error("Expected DoctorAvailable=true when installed")
	}
}

func TestDetectErrorSwallowed(t *testing.T) {
	saved := process.Run
	t.Cleanup(func() { process.Run = saved })

	// Even if process.Run returns a hard error, Detect should return nil error.
	process.Run = func(_ context.Context, _ string, _ []string, _ process.Options) (process.Result, error) {
		return process.Result{Error: errors.New("binary not found")}, errors.New("binary not found")
	}

	cap, err := contextmode.Detect(context.Background())
	if err != nil {
		t.Fatalf("Detect must swallow errors: got %v", err)
	}
	if cap.Installed {
		t.Error("Expected Installed=false on error")
	}
}
