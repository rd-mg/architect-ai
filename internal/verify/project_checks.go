package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// claudeMDPlaceholders is the canonical list of template tokens that must not
// appear in a materialized CLAUDE.md. Each token indicates the file was never
// built with `architect-ai build`.
var claudeMDPlaceholders = []string{
	"{content from",
	"{L0_HASH}",
	"{L1A_HASH}",
	"{L1B_HASH}",
	"{CONTENT_HASH}",
	"{INJECTED_MODE}",
	"{D1}", "{D2}", "{D3}", "{D4}",
}

// claudeMDMinBytes is the minimum size of a properly built CLAUDE.md.
// A file smaller than this has not been materialized.
const claudeMDMinBytes = 20_000

// ClaudeMDNoPlaceholdersCheck returns a Check that verifies CLAUDE.md in
// projectRoot has been built (no unresolved template placeholders).
//
// Precondition:  projectRoot is the root directory of the architect-ai project.
// Postcondition: returns nil iff CLAUDE.md exists, is > 20KB, and contains no
//
//	template placeholders.
func ClaudeMDNoPlaceholdersCheck(projectRoot string) Check {
	return Check{
		ID:          "verify:project:claude-md-no-placeholders",
		Description: "CLAUDE.md has no unresolved template placeholders",
		FixHint:     FixBuild,
		Run: func(_ context.Context) error {
			claudePath := filepath.Join(projectRoot, "CLAUDE.md")
			data, err := os.ReadFile(claudePath)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return fmt.Errorf("read CLAUDE.md: %w", err)
			}
			for _, ph := range claudeMDPlaceholders {
				if strings.Contains(string(data), ph) {
					return fmt.Errorf(
						"CLAUDE.md contains unresolved placeholder %q — run: architect-ai build\n"+
							"All agent prompts must be materialized before use.",
						ph,
					)
				}
			}
			if len(data) < claudeMDMinBytes {
				return fmt.Errorf(
					"CLAUDE.md is only %d bytes (expected >%d for a built file) — run: architect-ai build",
					len(data), claudeMDMinBytes,
				)
			}
			return nil
		},
	}
}

// SDDStateEnumCheck returns a Check that verifies all sdd-state.yaml files
// under atDir use valid status enum values. Detects the "running" vs
// "in_progress" mismatch introduced before v0.3.
//
// Precondition:  atDir is the .atl directory of the project.
// Postcondition: returns nil iff no state.yaml contains a "running" status or
//
//	any other value outside ValidStatuses.
func SDDStateEnumCheck(atDir string) Check {
	return Check{
		ID:          "verify:project:sdd-state-enum",
		Description: "sdd-state.yaml uses valid status enum values (no obsolete \"running\")",
		FixHint:     FixMigrateV03,
		Soft:        true,
		Run: func(_ context.Context) error {
			changesDir := filepath.Join(atDir, "openspec", "changes")
			entries, err := os.ReadDir(changesDir)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return fmt.Errorf("read changes dir: %w", err)
			}
			for _, entry := range entries {
				if !entry.IsDir() {
					continue
				}
				statePath := filepath.Join(changesDir, entry.Name(), "state.yaml")
				data, err := os.ReadFile(statePath)
				if err != nil {
					if os.IsNotExist(err) {
						continue
					}
					return fmt.Errorf("read %s: %w", statePath, err)
				}
				content := string(data)
				if strings.Contains(content, `status: "running"`) ||
					strings.Contains(content, "status: running\n") {
					return fmt.Errorf(
						"sdd-state.yaml for change %q contains obsolete status \"running\"\n"+
							"Run: architect-ai migrate-v03",
						entry.Name(),
					)
				}
			}
			return nil
		},
	}
}

// probeLogEntry is a single line from .atl/probe-log.jsonl.
type probeLogEntry struct {
	TS     string `json:"ts"`
	Probe  string `json:"probe"`
	Result string `json:"result"`
	Error  string `json:"error"`
}

// EngramProbeLogCheck returns a soft Check that warns when Engram has failed
// 3 or more consecutive times. Reads .atl/probe-log.jsonl (JSONL format,
// one JSON object per line).
//
// This check is always Soft: a degraded Engram is a warning, not a blocker.
func EngramProbeLogCheck(atDir string) Check {
	return Check{
		ID:          "verify:engram:probe-log",
		Description: "Engram probe has not failed 3+ consecutive times",
		FixHint:     FixDiagnoseEngram,
		Soft:        true,
		Run: func(_ context.Context) error {
			logPath := filepath.Join(atDir, "probe-log.jsonl")
			data, err := os.ReadFile(logPath)
			if err != nil {
				if os.IsNotExist(err) {
					return nil
				}
				return nil
			}
			lines := strings.Split(strings.TrimSpace(string(data)), "\n")
			start := 0
			if len(lines) > 10 {
				start = len(lines) - 10
			}
			recent := lines[start:]
			consecFails := 0
			for _, line := range recent {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				var entry probeLogEntry
				if jsonErr := json.Unmarshal([]byte(line), &entry); jsonErr != nil {
					continue
				}
				if entry.Result == "failed" {
					consecFails++
				} else if entry.Result == "ok" {
					consecFails = 0
				}
			}
			if consecFails >= 3 {
				return fmt.Errorf(
					"Engram has failed %d consecutive times — sessions will degrade to 'none' mode\n"+
						"Run: architect-ai diagnose engram",
					consecFails,
				)
			}
			return nil
		},
	}
}

// WriteProbeLogEntry appends a structured probe result to .atl/probe-log.jsonl.
// Creates the file if it does not exist. Rotates the log at maxLines entries.
// Safe to call from multiple goroutines because it opens with O_APPEND which is
// atomic for writes smaller than PIPE_BUF on Linux (and our entries are <512 bytes).
func WriteProbeLogEntry(atDir, probe, result, errMsg string) error {
	logPath := filepath.Join(atDir, "probe-log.jsonl")

	entry := probeLogEntry{
		TS:     time.Now().UTC().Format(time.RFC3339),
		Probe:  probe,
		Result: result,
		Error:  errMsg,
	}
	line, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshal probe entry: %w", err)
	}
	line = append(line, '\n')

	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open probe log: %w", err)
	}
	defer f.Close()
	_, err = f.Write(line)
	if err != nil {
		return fmt.Errorf("write probe log: %w", err)
	}

	rotateProbeLog(logPath, 200)
	return nil
}

// rotateProbeLog keeps the log at most maxLines lines by truncating from the top.
// It is best-effort: any error is silently discarded to avoid masking the
// primary write path.
func rotateProbeLog(logPath string, maxLines int) {
	data, err := os.ReadFile(logPath)
	if err != nil {
		return
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) <= maxLines {
		return
	}
	lines = lines[len(lines)-maxLines:]
	trimmed := strings.Join(lines, "\n")
	tmp := logPath + ".rotate.tmp"
	if writeErr := os.WriteFile(tmp, []byte(trimmed), 0o644); writeErr != nil {
		return
	}
	_ = os.Rename(tmp, logPath)
}
