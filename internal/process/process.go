package process

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"syscall"
	"time"
)

type Category string

const (
	FastCheck     Category = "fast-check"
	Install       Category = "install"
	AgentGenerate Category = "agent-generate"
	MCPHealth     Category = "mcp-health"
	HookCommand   Category = "hook"
)

type Options struct {
	Timeout        time.Duration
	MaxOutputBytes int64
	KillGrace      time.Duration
	StdinClosed    bool
}

type Result struct {
	Stdout   []byte
	Stderr   []byte
	ExitCode int
	TimedOut bool
	Error    error
}

func OptionsFor(category Category) Options {
	switch category {
	case FastCheck:
		return Options{Timeout: 5 * time.Second, MaxOutputBytes: 64 * 1024, KillGrace: 1 * time.Second, StdinClosed: true}
	case Install:
		return Options{Timeout: 10 * time.Minute, MaxOutputBytes: 4 << 20, KillGrace: 5 * time.Second, StdinClosed: true}
	case AgentGenerate:
		return Options{Timeout: 2 * time.Minute, MaxOutputBytes: 1 << 20, KillGrace: 2 * time.Second, StdinClosed: true}
	case MCPHealth:
		return Options{Timeout: 15 * time.Second, MaxOutputBytes: 256 << 10, KillGrace: 2 * time.Second, StdinClosed: true}
	case HookCommand:
		return Options{Timeout: 2 * time.Second, MaxOutputBytes: 64 << 10, KillGrace: 1 * time.Second, StdinClosed: true}
	default:
		return Options{Timeout: 30 * time.Second, MaxOutputBytes: 1 << 20, KillGrace: 2 * time.Second, StdinClosed: true}
	}
}

func Run(ctx context.Context, cmdName string, args []string, opts Options) (Result, error) {
	if opts.Timeout == 0 {
		opts.Timeout = 30 * time.Second
	}
	if opts.MaxOutputBytes == 0 {
		opts.MaxOutputBytes = 1 << 20
	}
	if opts.KillGrace == 0 {
		opts.KillGrace = 2 * time.Second
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, opts.Timeout)
	defer cancel()

	cmd := exec.CommandContext(timeoutCtx, cmdName, args...)
	
	// Create a new process group so we can kill the entire tree
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if opts.StdinClosed {
		// Close stdin by not providing it (os.DevNull equivalent or empty pipe)
		cmd.Stdin = nil
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout = &stdoutBuf
	cmd.Stderr = &stderrBuf

	err := cmd.Start()
	if err != nil {
		return Result{Error: err}, fmt.Errorf("failed to start process: %w", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	var waitErr error
	var timedOut bool

	select {
	case <-timeoutCtx.Done():
		timedOut = true
		// Kill the process group
		if cmd.Process != nil {
			syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		}
		waitErr = errors.New("process timed out")
		// Wait for goroutine to exit
		<-done
	case err := <-done:
		waitErr = err
	}

	exitCode := 0
	var exitError *exec.ExitError
	if errors.As(waitErr, &exitError) {
		exitCode = exitError.ExitCode()
	} else if waitErr != nil {
		// Not an exit error or timed out
		exitCode = -1
	}

	// Truncate output if needed
	stdout := stdoutBuf.Bytes()
	if int64(len(stdout)) > opts.MaxOutputBytes {
		stdout = stdout[:opts.MaxOutputBytes]
	}
	
	stderr := stderrBuf.Bytes()
	if int64(len(stderr)) > opts.MaxOutputBytes {
		stderr = stderr[:opts.MaxOutputBytes]
	}

	return Result{
		Stdout:   stdout,
		Stderr:   stderr,
		ExitCode: exitCode,
		TimedOut: timedOut,
		Error:    waitErr,
	}, waitErr
}
