package cli

import (
	"fmt"
	"io"

	"github.com/rd-mg/architect-ai/internal/gate"
	"github.com/rd-mg/architect-ai/internal/paths"
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

func RunGate(args []string, stdout io.Writer, stderr io.Writer) error {
	devMode := false
	var filteredArgs []string
	for _, arg := range args {
		if arg == "--dev" {
			devMode = true
		} else {
			filteredArgs = append(filteredArgs, arg)
		}
	}
	ctx := paths.New(".", devMode)

	if len(filteredArgs) < 1 {
		return fmt.Errorf("usage: architect-ai gate [check|inject|purge|l2-purge] [--dev]")
	}

	g := gate.New(GateSourceFile)

	switch filteredArgs[0] {
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
		if exitCode != 0 {
			return fmt.Errorf("gate check failed")
		}
		return nil

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
		if exitCode != 0 {
			return fmt.Errorf("gate injection failed")
		}
		return nil

	case "purge":
		results := g.Purge(GateTargets)
		for _, r := range results {
			if r.Error != nil {
				fmt.Fprintf(stderr, "FAIL  %s: %v\n", r.File, r.Error)
			} else {
				fmt.Fprintf(stdout, "OK    %s: gate removed\n", r.File)
			}
		}
		return nil

	case "l2-purge":
		results := g.PurgeL2AutoScoring(ctx.L2SkillGlob())
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
		if exitCode != 0 {
			return fmt.Errorf("l2-purge failed")
		}
		return nil

	default:
		return fmt.Errorf("unknown subcommand: %s", filteredArgs[0])
	}
}
