package cli

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/rd-mg/architect-ai/internal/install/architect"
)

// GuardResult is returned by the guard check subcommand as JSON.
type GuardResult struct {
	Verdict   string `json:"verdict"`             // "ok" | "blocked"
	Rule      string `json:"rule,omitempty"`      // Trigger name
	Reason    string `json:"reason,omitempty"`    // Human-readable explanation
	Violation int    `json:"violation,omitempty"` // 0 on ok, > 0 on blocked
}

// RunGuard implements the `architect-ai guard` CLI subcommand.
func RunGuard(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return errors.New("guard requires a subcommand: check")
	}

	fs := flag.NewFlagSet("guard check", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	ref := fs.Int("ref", 0, "files referenced (read)")
	write := fs.Int("write", 0, "files to write")
	tests := fs.Bool("tests", false, "involves test execution")
	build := fs.Bool("build", false, "involves build step")
	pr := fs.Bool("pr", false, "is PR creation")
	incident := fs.Bool("incident", false, "is incident response")
	toolCalls := fs.Int("calls", 0, "running tool call count")
	exploratory := fs.Int("explore", 0, "exploratory reads in session")
	edits := fs.Int("edits", 0, "non-mechanical edits in session")
	jsonOut := fs.Bool("json", false, "output as JSON")

	// args[0]="check" — flags start at args[1:]
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}

	tc := architect.TaskContext{
		FilesReferenced:    *ref,
		FilesToWrite:       *write,
		InvolvesTests:      *tests,
		InvolvesBuild:      *build,
		IsPRCreation:       *pr,
		IsIncident:         *incident,
		ToolCallCount:      *toolCalls,
		ExploratoryReads:   *exploratory,
		NonMechanicalEdits: *edits,
	}

	result := GuardResult{Verdict: "ok", Violation: 0}

	if trigger := architect.CheckMandatoryTriggers(tc); trigger.Fired {
		result = GuardResult{
			Verdict:   "blocked",
			Rule:      trigger.Rule,
			Reason:    trigger.Reason,
			Violation: 1,
		}
	}

	if result.Verdict == "blocked" {
		if *jsonOut {
			enc := json.NewEncoder(stdout)
			enc.SetIndent("", "  ")
			_ = enc.Encode(result)
		} else {
			fmt.Fprintf(stdout, "GUARD BLOCKED: %s — %s\n", result.Rule, result.Reason)
		}
		return fmt.Errorf("BLOCKED by %s: %s", result.Rule, result.Reason)
	}

	if *jsonOut {
		_ = json.NewEncoder(stdout).Encode(result)
		return nil
	}

	fmt.Fprintln(stdout, "GUARD OK — no mandatory triggers fire")
	return nil
}

// RunGuardAuto is a convenience wrapper for AI agents to self-check.
// It parses args from a single string and exits with code 0 (ok) or 1 (blocked).
// Intended for `architect-ai guard check ...` calls from agent bash.
func RunGuardAuto(args []string) int {
	if err := RunGuard(args, os.Stdout); err != nil {
		if strings.HasPrefix(err.Error(), "BLOCKED") {
			return 1
		}
		return 2
	}
	return 0
}

// ValidateAgentWriteDir checks that the agent's write operations are within
// the project root. Returns error if any path escapes.
func ValidateAgentWriteDir(projectRoot string, paths ...string) error {
	absRoot, err := filepath.Abs(projectRoot)
	if err != nil {
		return fmt.Errorf("resolve project root: %w", err)
	}
	for _, p := range paths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("resolve path %q: %w", p, err)
		}
		if !strings.HasPrefix(absPath, absRoot) {
			return fmt.Errorf("path %q escapes project root %q", p, absRoot)
		}
	}
	return nil
}
