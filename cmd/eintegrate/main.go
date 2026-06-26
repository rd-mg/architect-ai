package main

import (
	"fmt"
	"os"
	"regexp"
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

	// Phase 3: Orchestrator redesign checks (E-17 through E-23)
	// E-17: L0 must NOT contain sequential_thinking Intention Gate
	if checkFile("internal/assets/generic/thinking-agent.md", "sequential_thinking") {
		errs = append(errs, "E-17: thinking-agent.md still contains sequential_thinking Intention Gate — remove it")
	}
	// E-18: L1a must NOT contain Router Gate
	if checkFile("internal/assets/generic/general-orchestrator.md", "ROUTER GATE") {
		errs = append(errs, "E-18: general-orchestrator.md still contains ROUTER GATE — remove it (L0 owns routing)")
	}
	// E-19: L1b must have Step 0 D1-D4 SDD classification
	if !checkFile("internal/assets/generic/sdd-orchestrator.md", "D1-D4 Classification (SDD-Specific)") {
		errs = append(errs, "E-19: sdd-orchestrator.md missing Step 0 D1-D4 SDD Classification")
	}
	// E-20: L0 must have Intent Router section
	if !checkFile("internal/assets/generic/thinking-agent.md", "Intent Router") {
		errs = append(errs, "E-20: thinking-agent.md missing Intent Router section (L0 proxy gate)")
	}
	// E-21: L0 Architecture Constitution must be compact (< 200 words)
	if constitutionWordCount("internal/assets/generic/thinking-agent.md") > 200 {
		errs = append(errs, "E-21: L0 Architecture Constitution exceeds 200 words — compress it")
	}
	// E-22: L1b must contain Forwarded State check in Session-Setup
	if !checkFile("internal/assets/generic/sdd-orchestrator.md", "session_state") {
		errs = append(errs, "E-22: sdd-orchestrator.md missing session_state forwarding check")
	}
	// E-23: L1a must contain Session State Reader
	if !checkFile("internal/assets/generic/general-orchestrator.md", "Session State Reader") {
		errs = append(errs, "E-23: general-orchestrator.md missing Session State Reader section")
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

func constitutionWordCount(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	content := string(data)
	start := strings.Index(content, "## Architecture Constitution")
	if start < 0 {
		return 0
	}
	end := strings.Index(content[start+len("## Architecture Constitution"):], "\n## ")
	if end < 0 {
		end = len(content) - start
	}
	section := content[start : start+end+len("## Architecture Constitution")]
	words := regexp.MustCompile(`\S+`).FindAllString(section, -1)
	return len(words)
}
