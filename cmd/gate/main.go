package main

import (
	"fmt"
	"io"
	"os"

	"github.com/rd-mg/architect-ai/internal/gate"
)

// Targets where the gate MUST exist
var GateTargets = []gate.Target{
	{
		File:         "internal/assets/opencode/sdd-orchestrator.md",
		InsertBefore: "## Delegation Rules",
		Required:     true,
	},
	{
		File:         "internal/assets/opencode/general-orchestrator.md",
		InsertBefore: "## Delegation Rules",
		Required:     true,
		ReplaceV1:    true, // Has v1 gate — replace with v2
	},
}

// Template targets — gate is included via {{ include }} in the templates
// After architect-ai build, the gate will be in the deployed configs automatically
var TemplateTargets = []string{
	"internal/assets/templates/CLAUDE.md.tmpl",
	"internal/assets/templates/GEMINI.md.tmpl",
	"internal/assets/templates/antigravity-agent.md.tmpl",
	"internal/assets/templates/copilot-instructions.md.tmpl",
}

var GateSourceFile = "internal/assets/_shared/adaptive-reasoning-gate-v2.md"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	if len(args) < 1 {
		fmt.Fprintln(stderr, "Usage: architect-ai gate [check|inject|purge|l2-purge]")
		return 1
	}

	g := gate.New(GateSourceFile)

	switch args[0] {
	case "check":
		results := g.Check(GateTargets)
		exitCode := 0
		for _, r := range results {
			status := "PRESENT"
			if !r.Present {
				status = "MISSING"
				exitCode = 1
			}
			version := r.Version
			if version == "" {
				version = "none"
			}
			fmt.Fprintf(stdout, "GATE %-60s %s (version: %s)\n", r.File, status, version)
		}
		return exitCode

	case "inject":
		results := g.Inject(GateTargets)
		exitCode := 0
		for _, r := range results {
			if r.Error != nil {
				fmt.Fprintf(stderr, "FAIL  %s: %v\n", r.File, r.Error)
				exitCode = 1
			} else {
				action := "injected"
				if r.AlreadyPresent {
					action = "skipped (already present)"
				}
				if r.Replaced {
					action = "replaced v1 with v2"
				}
				fmt.Fprintf(stdout, "OK    %s: %s\n", r.File, action)
			}
		}
		// Also ensure {{ include }} is in template targets
		g.EnsureIncludeInTemplates(TemplateTargets)
		return exitCode

	case "purge":
		results := g.Purge(GateTargets)
		for _, r := range results {
			if r.Error != nil {
				fmt.Fprintf(stderr, "FAIL  %s: %v\n", r.File, r.Error)
			} else {
				fmt.Fprintf(stdout, "OK    %s: gate removed\n", r.File)
			}
		}
		return 0

	case "l2-purge":
		results := g.PurgeL2AutoScoring()
		exitCode := 0
		for _, r := range results {
			if r.Error != nil {
				fmt.Fprintf(stderr, "FAIL  %s: %v\n", r.File, r.Error)
				exitCode = 1
			} else {
				if r.Modified {
					fmt.Fprintf(stdout, "PATCHED %s: auto-scoring removed\n", r.File)
				} else {
					fmt.Fprintf(stdout, "SKIP    %s: no auto-scoring pattern found\n", r.File)
				}
			}
		}
		return exitCode

	default:
		fmt.Fprintf(stderr, "Unknown subcommand: %s\n", args[0])
		return 1
	}
}
