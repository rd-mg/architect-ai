package process

import (
	"context"
	"testing"
	"time"
)

func TestRun_Success(t *testing.T) {
	ctx := context.Background()
	res, err := Run(ctx, "echo", []string{"hello"}, Options{Timeout: 1 * time.Second})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(res.Stdout) != "hello\n" {
		t.Errorf("expected hello\\n, got %q", string(res.Stdout))
	}
	if res.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", res.ExitCode)
	}
	if res.TimedOut {
		t.Errorf("expected TimedOut false, got true")
	}
}

func TestRun_Timeout(t *testing.T) {
	ctx := context.Background()
	res, err := Run(ctx, "sh", []string{"-c", "sleep 60"}, Options{
		Timeout:   100 * time.Millisecond,
		KillGrace: 50 * time.Millisecond,
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if !res.TimedOut {
		t.Fatal("expected TimedOut=true")
	}
	if res.ExitCode == 0 {
		t.Error("expected non-zero exit code for timeout")
	}
}

func TestRun_OutputTruncation(t *testing.T) {
	ctx := context.Background()
	res, _ := Run(ctx, "sh", []string{"-c", "echo 1234567890"}, Options{
		Timeout:        1 * time.Second,
		MaxOutputBytes: 5,
	})
	
	if string(res.Stdout) != "12345" {
		t.Errorf("expected truncated output '12345', got %q", string(res.Stdout))
	}
}

func TestOptionsFor(t *testing.T) {
	opts := OptionsFor(FastCheck)
	if opts.Timeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", opts.Timeout)
	}
	if !opts.StdinClosed {
		t.Errorf("expected stdin closed")
	}
}
