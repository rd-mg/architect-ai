package firewall

import (
	"fmt"
	"os"
	"strings"

	"github.com/rd-mg/architect-ai/internal/paths"
)

const (
	FirewallSource  = "internal/assets/_shared/caveman-firewall.md"
	FirewallStart   = "<!-- caveman-firewall:v1:START -->"
	FirewallEnd     = "<!-- caveman-firewall:v1:END -->"

	SddApplyProtocol = "internal/assets/opencode/sdd-phase-protocols/sdd-apply.md"
)

// FirewallTarget represents a file where the caveman firewall should be present.
type FirewallTarget struct {
	File         string
	CheckPattern string
	InjectMode   string // "include" | "reference" | "patch"
	InjectBefore string
}

// GetTargets returns the firewall targets according to the path context.
func GetTargets(ctx paths.Context) []FirewallTarget {
	targets := []FirewallTarget{
		{
			File:         SddApplyProtocol,
			CheckPattern: "caveman-firewall",
			InjectMode:   "include",
			InjectBefore: "## Atomic Commit Protocol",
		},
		{
			File:         ctx.SddApplySkillPath(),
			CheckPattern: "caveman-firewall",
			InjectMode:   "reference",
			InjectBefore: "## What You Receive",
		},
	}
	
	if reg := ctx.RegistryPath(); reg != "" {
		targets = append(targets, FirewallTarget{
			File:         reg,
			CheckPattern: "caveman_firewall_active",
			InjectMode:   "patch",
			InjectBefore: "protected_facts MUST always include",
		})
	}
	return targets
}

// CheckResult holds the result of checking a single target file.
type CheckResult struct {
	File    string
	Present bool
	Pattern string
}

// InjectResult holds the result of injecting into a single target file.
type InjectResult struct {
	File  string
	OK    bool
	Error string
}

// Check verifies that the firewall source exists and all targets contain their patterns.
func Check(targets []FirewallTarget) ([]CheckResult, bool) {
	results := []CheckResult{}

	// First check the source file exists
	_, err := os.Stat(FirewallSource)
	results = append(results, CheckResult{
		File:    FirewallSource,
		Present: err == nil,
		Pattern: "source file exists",
	})

	allOK := err == nil
	for _, t := range targets {
		data, err := os.ReadFile(t.File)
		if err != nil {
			results = append(results, CheckResult{
				File:    t.File,
				Present: false,
				Pattern: t.CheckPattern,
			})
			allOK = false
			continue
		}
		present := strings.Contains(string(data), t.CheckPattern)
		if !present {
			allOK = false
		}
		results = append(results, CheckResult{
			File:    t.File,
			Present: present,
			Pattern: t.CheckPattern,
		})
	}
	return results, allOK
}

// Inject injects firewall markers into targets that don't have them.
func Inject(targets []FirewallTarget) []InjectResult {
	results := []InjectResult{}

	for _, t := range targets {
		data, err := os.ReadFile(t.File)
		if err != nil {
			results = append(results, InjectResult{
				File:  t.File,
				OK:    false,
				Error: fmt.Sprintf("file not found: %v", err),
			})
			continue
		}
		content := string(data)

		if strings.Contains(content, t.CheckPattern) {
			results = append(results, InjectResult{
				File: t.File,
				OK:   true,
				Error: "already present",
			})
			continue
		}

		var injection string
		switch t.InjectMode {
		case "include":
			injection = "\n{{ include \"_shared/caveman-firewall.md\" }}\n\n"
		case "reference":
			injection = "\n## Register Firewall (MANDATORY)\n\nSee `internal/assets/_shared/caveman-firewall.md` for full rules.\nKey rule: NORMAL register for ALL code, comments, commits. Cannot be disabled.\n\n"
		case "patch":
			injection = "\n- key: caveman_firewall_active\n  value: true\n  note: \"NORMAL register mandatory for ALL code artifacts. Cannot be disabled.\"\n"
		}

		if t.InjectBefore != "" {
			idx := strings.Index(content, t.InjectBefore)
			if idx < 0 {
				content = content + injection
			} else {
				content = content[:idx] + injection + content[idx:]
			}
		} else {
			content = content + injection
		}

		if err := atomicWrite(t.File, content); err != nil {
			results = append(results, InjectResult{
				File:  t.File,
				OK:    false,
				Error: err.Error(),
			})
			continue
		}
		results = append(results, InjectResult{
			File: t.File,
			OK:   true,
			Error: "injected",
		})
	}
	return results
}

func atomicWrite(path, content string) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
