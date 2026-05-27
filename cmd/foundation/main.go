package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"

	"github.com/rd-mg/architect-ai/internal/foundation"
)

// Tier1Skills are the hardcoded set of skills whose compact rules are
// always injected into foundation.md.
var tier1Skills = []foundation.SkillRef{
	{Name: "adaptive-reasoning", CompactLines: 100},
	{Name: "context-guardian", CompactLines: 100},
	{Name: "architecture-guardrails", CompactLines: 100},
	{Name: "work-unit-commits", CompactLines: 150},
	{Name: "branch-pr", CompactLines: 80},
	{Name: "bash-expert", CompactLines: 80},
	{Name: "ripgrep", CompactLines: 40},
}

func main() {
	os.MkdirAll(".atl/_generated", 0755)

	// Extract compact rules from /internal/assets/skills/
	extractor := foundation.NewExtractor("internal/assets/skills")
	results := extractor.ExtractAll(tier1Skills)

	// Check for errors
	var warnings []string
	for _, r := range results {
		if r.Err != nil {
			warnings = append(warnings, fmt.Sprintf("WARN: %s — %v", r.Name, r.Err))
		}
		if r.Content == "" {
			warnings = append(warnings, fmt.Sprintf("WARN: skill '%s' has empty compact rules — SKIPPED", r.Name))
		}
	}

	// Generate foundation block
	gen := foundation.NewGenerator()
	block := gen.Generate(results)

	// Write foundation.md
	if err := block.WriteToFile(".atl/_generated/foundation.md"); err != nil {
		fmt.Fprintf(os.Stderr, "FOUNDATION_ERROR: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  OK   .atl/_generated/foundation.md (%d bytes)\n", len(block.Content))

	// Compute and write skills-lock.json
	lock := computeLock(tier1Skills)
	lockData, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "FOUNDATION_ERROR: cannot marshal lock: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(".atl/_generated/skills-lock.json", lockData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "FOUNDATION_ERROR: cannot write lock: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("  OK   .atl/_generated/skills-lock.json (%d bytes)\n", len(lockData))

	for _, w := range warnings {
		fmt.Fprintln(os.Stderr, w)
	}
}

type skillsLock struct {
	Version   int               `json:"version"`
	Skills    map[string]string `json:"skills"`
	Generated string            `json:"generated_at"`
}

func computeLock(skills []foundation.SkillRef) skillsLock {
	lock := skillsLock{
		Version:   1,
		Skills:    make(map[string]string),
		Generated: "auto",
	}
	for _, s := range skills {
		data, err := os.ReadFile(fmt.Sprintf("internal/assets/skills/%s/SKILL.md", s.Name))
		if err != nil {
			lock.Skills[s.Name] = "err"
			continue
		}
		sum := sha256.Sum256(data)
		lock.Skills[s.Name] = fmt.Sprintf("%x", sum[:8])
	}
	return lock
}
