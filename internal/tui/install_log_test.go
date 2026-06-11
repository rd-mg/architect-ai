package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rd-mg/architect-ai/internal/pipeline"
)

func TestWriteInstallLog_CreatesFile(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	result := pipeline.ExecutionResult{
		Apply: pipeline.StageResult{
			Steps: []pipeline.StepResult{
				{StepID: "install-deps", Status: pipeline.StepStatusSucceeded},
			},
		},
	}
	writeInstallLog(result)

	logPath := filepath.Join(homeDir, ".architect-ai", "install-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("install log not created: %v", err)
	}
	if !strings.Contains(string(data), `"status":"ok"`) {
		t.Errorf("expected ok status in log, got: %s", data)
	}
}

func TestWriteInstallLog_FailedStepsRecorded(t *testing.T) {
	homeDir := t.TempDir()
	t.Setenv("HOME", homeDir)

	result := pipeline.ExecutionResult{
		Apply: pipeline.StageResult{
			Steps: []pipeline.StepResult{
				{StepID: "install-deps", Status: pipeline.StepStatusSucceeded},
				{StepID: "configure-agents", Status: pipeline.StepStatusFailed,
					Err: &testError{"connection timeout"}},
			},
		},
	}
	writeInstallLog(result)

	logPath := filepath.Join(homeDir, ".architect-ai", "install-log.jsonl")
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("install log not created: %v", err)
	}

	var entry struct {
		Status      string `json:"status"`
		FailedSteps []struct {
			ID  string `json:"id"`
			Err string `json:"error"`
		} `json:"failed_steps"`
	}
	line := strings.TrimSpace(string(data))
	if err := json.Unmarshal([]byte(line), &entry); err != nil {
		t.Fatalf("parse log entry: %v", err)
	}
	if entry.Status != "failed" {
		t.Errorf("expected status=failed, got %q", entry.Status)
	}
	if len(entry.FailedSteps) != 1 {
		t.Errorf("expected 1 failed step, got %d", len(entry.FailedSteps))
	}
	if entry.FailedSteps[0].ID != "configure-agents" {
		t.Errorf("expected failed step id=configure-agents, got %q", entry.FailedSteps[0].ID)
	}
}

func TestRotateInstallLog_KeepsMaxLines(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "install-log.jsonl")

	var sb strings.Builder
	for i := 0; i < 600; i++ {
		sb.WriteString(`{"ts":"2026-01-01T00:00:00Z","status":"ok"}`)
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(logPath, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	rotateInstallLog(logPath, 500)

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read rotated log: %v", err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) > 500 {
		t.Errorf("rotated log has %d lines, want <= 500", len(lines))
	}
}

func TestSetProgram_StoresReference(t *testing.T) {
	m := &Model{}
	if m.Program != nil {
		t.Error("Program should be nil before SetProgram")
	}
	m.SetProgram(nil)
	if m.Program != nil {
		t.Error("Program should be nil after SetProgram(nil)")
	}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }
