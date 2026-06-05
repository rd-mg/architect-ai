package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	var errs []string

	if !checkFile(".atl/skill-registry.md", "sdd-propose") {
		errs = append(errs, "E-03: sdd-propose missing in skill-registry.md")
	}
	if !checkFile(".atl/skill-registry.md", "PHASE 1 — Gather") {
		errs = append(errs, "E-04: context-guardian protocol missing in skill-registry.md")
	}
	if !checkFile("internal/assets/skills/sdd-verify/SKILL.md", "THEY ARE PROBABLY LYING") {
		errs = append(errs, "E-05: Adversarial Verification Stance missing in sdd-verify")
	}
	if !checkFile(".atl/skill-registry.md", "ctx_fetch_and_index") {
		errs = append(errs, "E-06: odoo ctx_fetch_and_index missing in skill-registry.md")
	}
	if !checkFile(".atl/skill-registry.md", "LspFindReferences") {
		errs = append(errs, "E-07: sdd-explore codegraph missing in skill-registry.md")
	}
	if !checkFile("internal/assets/templates/GEMINI.md.tmpl", "KEYWORD ROUTING") {
		errs = append(errs, "E-08: KEYWORD ROUTING missing in GEMINI.md.tmpl")
	}
	if !checkFile("internal/assets/templates/antigravity-agent.md.tmpl", "KEYWORD ROUTING") {
		errs = append(errs, "E-08: KEYWORD ROUTING missing in antigravity-agent.md.tmpl")
	}
	if !checkFile("internal/assets/workflows/arch-materialize.md", "MATERIALIZE_COMPLETE") {
		errs = append(errs, "E-09: arch-materialize.md missing or incomplete")
	}
	if !checkFile("internal/assets/workflows/arch-hardening.md", "HARDENING_COMPLETE") {
		errs = append(errs, "E-10: arch-hardening.md missing HARDENING_COMPLETE")
	}

	if len(errs) > 0 {
		fmt.Fprintln(os.Stderr, "Phase E Integration checks failed:")
		for _, e := range errs {
			fmt.Fprintln(os.Stderr, " - "+e)
		}
		os.Exit(1)
	}

	fmt.Println("Phase E Integration Complete")
}

func checkFile(path, substr string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return strings.Contains(string(data), substr)
}
