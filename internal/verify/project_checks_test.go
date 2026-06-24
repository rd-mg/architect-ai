package verify

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestClaudeMDNoPlaceholdersCheck_Pass(t *testing.T) {
	dir := t.TempDir()
	content := strings.Repeat("x", claudeMDMinBytes+1)
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	check := ClaudeMDNoPlaceholdersCheck(dir)
	if err := check.Run(context.Background()); err != nil {
		t.Errorf("expected no error for clean CLAUDE.md, got: %v", err)
	}
}

func TestClaudeMDNoPlaceholdersCheck_FailOnPlaceholder(t *testing.T) {
	cases := []string{
		"{content from .atl/agents/architect.md}",
		"{L0_HASH}",
		"{L1A_HASH}",
		"{INJECTED_MODE}",
		"{D4}",
	}
	for _, ph := range cases {
		t.Run(ph, func(t *testing.T) {
			dir := t.TempDir()
			content := strings.Repeat("a", claudeMDMinBytes) + ph
			if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte(content), 0o644); err != nil {
				t.Fatalf("setup: %v", err)
			}
			check := ClaudeMDNoPlaceholdersCheck(dir)
			if err := check.Run(context.Background()); err == nil {
				t.Errorf("expected error for placeholder %q, got nil", ph)
			}
		})
	}
}

func TestClaudeMDNoPlaceholdersCheck_FailOnSmallFile(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "CLAUDE.md"), []byte("too small"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	check := ClaudeMDNoPlaceholdersCheck(dir)
	if err := check.Run(context.Background()); err == nil {
		t.Error("expected error for small CLAUDE.md, got nil")
	}
}

func TestClaudeMDNoPlaceholdersCheck_MissingFile(t *testing.T) {
	dir := t.TempDir()
	check := ClaudeMDNoPlaceholdersCheck(dir)
	if err := check.Run(context.Background()); err != nil {
		t.Errorf("missing CLAUDE.md should not error (project may not use it): %v", err)
	}
}

func TestSDDStateEnumCheck_Pass(t *testing.T) {
	dir := t.TempDir()
	changesDir := filepath.Join(dir, "openspec", "changes", "my-change")
	if err := os.MkdirAll(changesDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	content := `version: "3.0"
change_name: "my-change"
phases:
  sdd-explore:
    status: "in_progress"
`
	if err := os.WriteFile(filepath.Join(changesDir, "state.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	check := SDDStateEnumCheck(dir)
	if err := check.Run(context.Background()); err != nil {
		t.Errorf("expected no error for in_progress status: %v", err)
	}
}

func TestSDDStateEnumCheck_FailOnRunning(t *testing.T) {
	dir := t.TempDir()
	changesDir := filepath.Join(dir, "openspec", "changes", "bad-change")
	if err := os.MkdirAll(changesDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	content := `version: "3.0"
change_name: "bad-change"
phases:
  sdd-apply:
    status: "running"
`
	if err := os.WriteFile(filepath.Join(changesDir, "state.yaml"), []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	check := SDDStateEnumCheck(dir)
	if err := check.Run(context.Background()); err == nil {
		t.Error("expected error for obsolete 'running' status, got nil")
	}
}

func TestEngramProbeLogCheck_NoLog(t *testing.T) {
	dir := t.TempDir()
	check := EngramProbeLogCheck(dir)
	if err := check.Run(context.Background()); err != nil {
		t.Errorf("missing probe log should not error: %v", err)
	}
}

func TestEngramProbeLogCheck_ThreeConsecutiveFails(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "probe-log.jsonl")

	entries := []probeLogEntry{
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "ok"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed", Error: "timeout"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed", Error: "timeout"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed", Error: "connection refused"},
	}
	var sb strings.Builder
	for _, e := range entries {
		line, _ := json.Marshal(e)
		sb.Write(line)
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(logPath, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	check := EngramProbeLogCheck(dir)
	if err := check.Run(context.Background()); err == nil {
		t.Error("expected error for 3 consecutive failures, got nil")
	}
}

func TestEngramProbeLogCheck_RecoveredAfterFails(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "probe-log.jsonl")

	entries := []probeLogEntry{
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "failed"},
		{TS: time.Now().UTC().Format(time.RFC3339), Probe: "engram", Result: "ok"},
	}
	var sb strings.Builder
	for _, e := range entries {
		line, _ := json.Marshal(e)
		sb.Write(line)
		sb.WriteByte('\n')
	}
	if err := os.WriteFile(logPath, []byte(sb.String()), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	check := EngramProbeLogCheck(dir)
	if err := check.Run(context.Background()); err != nil {
		t.Errorf("after recovery 'ok' entry, check should pass: %v", err)
	}
}

func TestWriteProbeLogEntry_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	if err := WriteProbeLogEntry(dir, "engram", "ok", ""); err != nil {
		t.Fatalf("WriteProbeLogEntry: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(dir, "probe-log.jsonl"))
	if err != nil {
		t.Fatalf("read log: %v", err)
	}
	if !strings.Contains(string(data), `"result":"ok"`) {
		t.Errorf("expected ok entry in log, got: %s", data)
	}
}
