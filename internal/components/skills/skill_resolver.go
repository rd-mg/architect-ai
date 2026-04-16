package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ResolvedStandards represents the set of standards and conventions
// matched for a specific task.
type ResolvedStandards struct {
	Rules           []ResolvedRule
	Conventions     []ConventionRef
	SkillResolution string // "injected", "fallback-registry", "fallback-path", "none"
}

// ResolvedRule holds the compact rule set for a matched skill.
type ResolvedRule struct {
	Skill   string
	Content string
}

// ConventionRef holds a reference to a project convention file.
type ConventionRef struct {
	Path  string
	Notes string
}

// ResolveStandardsForTask searches the project registry (local or memory)
// to find skills and conventions relevant to the given task and modified paths.
func ResolveStandardsForTask(projectRoot, task, phase string, paths []string) (ResolvedStandards, error) {
	// For now, we attempt to read from .atl/skill-registry.md
	registryPath := filepath.Join(projectRoot, ".atl", "skill-registry.md")
	data, err := os.ReadFile(registryPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ResolvedStandards{SkillResolution: "none"}, nil
		}
		return ResolvedStandards{}, fmt.Errorf("read skill registry: %w", err)
	}

	content := string(data)
	rs := ResolvedStandards{
		Rules:           make([]ResolvedRule, 0),
		Conventions:     make([]ConventionRef, 0),
		SkillResolution: "injected",
	}

	// 1. Extract Compact Rules
	rs.Rules = extractCompactRules(content, task, paths)

	// 2. Extract Project Conventions
	rs.Conventions = extractConventions(content)

	if len(rs.Rules) == 0 && len(rs.Conventions) == 0 {
		rs.SkillResolution = "none"
	}

	return rs, nil
}

// BuildResolvedStandardsBlock generates the markdown block to be injected
// into the sub-agent's system prompt.
func BuildResolvedStandardsBlock(rs ResolvedStandards) string {
	if len(rs.Rules) == 0 && len(rs.Conventions) == 0 {
		return ""
	}

	var b strings.Builder
	if len(rs.Rules) > 0 {
		b.WriteString("## Project Standards (auto-resolved)\n\n")
		for _, rule := range rs.Rules {
			b.WriteString(fmt.Sprintf("### Skill: %s\n%s\n\n", rule.Skill, rule.Content))
		}
	}

	if len(rs.Conventions) > 0 {
		b.WriteString("## Project Conventions\n\n")
		b.WriteString("Read these files for project-specific patterns:\n")
		for _, conv := range rs.Conventions {
			b.WriteString(fmt.Sprintf("- %s — %s\n", conv.Path, conv.Notes))
		}
		b.WriteString("\n")
	}

	return b.String()
}

func extractCompactRules(registryContent, task string, paths []string) []ResolvedRule {
	rules := make([]ResolvedRule, 0)
	// Simple matching logic: find the "## Compact Rules" section and extract blocks.
	// In a real implementation, we would use a proper markdown parser or regex.
	// For the rebase, we implement a robust enough version.
	
	section := extractSection(registryContent, "## Compact Rules")
	if section == "" {
		return rules
	}

	// Blocks are under "### Skill: <name>"
	lines := strings.Split(section, "\n")
	var currentSkill string
	var currentContent strings.Builder
	
	for _, line := range lines {
		if strings.HasPrefix(line, "### Skill: ") {
			if currentSkill != "" {
				if matchSkill(currentSkill, task, paths) {
					rules = append(rules, ResolvedRule{
						Skill:   currentSkill,
						Content: strings.TrimSpace(currentContent.String()),
					})
				}
				currentContent.Reset()
			}
			currentSkill = strings.TrimPrefix(line, "### Skill: ")
		} else if currentSkill != "" {
			currentContent.WriteString(line + "\n")
		}
	}
	
	// Last one
	if currentSkill != "" && matchSkill(currentSkill, task, paths) {
		rules = append(rules, ResolvedRule{
			Skill:   currentSkill,
			Content: strings.TrimSpace(currentContent.String()),
		})
	}

	return rules
}

func extractConventions(registryContent string) []ConventionRef {
	refs := make([]ConventionRef, 0)
	section := extractSection(registryContent, "## Project Conventions")
	if section == "" {
		return refs
	}

	lines := strings.Split(section, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") {
			// Format: - path — notes
			parts := strings.SplitN(strings.TrimPrefix(line, "- "), " — ", 2)
			if len(parts) == 2 {
				refs = append(refs, ConventionRef{
					Path:  parts[0],
					Notes: parts[1],
				})
			}
		}
	}
	return refs
}

func extractSection(content, title string) string {
	start := strings.Index(content, title)
	if start == -1 {
		return ""
	}
	
	// Section ends at the next header of same or higher level
	rest := content[start+len(title):]
	end := strings.Index(rest, "\n## ")
	if end == -1 {
		return rest
	}
	return rest[:end]
}

func matchSkill(skillName, task string, paths []string) bool {
	// Simple matching: if the skill name appears in the task or any path extension matches.
	// In a more advanced version, we'd use the Trigger mapping from the registry.
	skillLower := strings.ToLower(skillName)
	taskLower := strings.ToLower(task)
	
	if strings.Contains(taskLower, skillLower) {
		return true
	}
	
	for _, path := range paths {
		ext := strings.ToLower(filepath.Ext(path))
		switch skillLower {
		case "go", "golang":
			if ext == ".go" {
				return true
			}
		case "typescript", "ts":
			if ext == ".ts" || ext == ".tsx" {
				return true
			}
		case "react":
			if ext == ".tsx" || ext == ".jsx" {
				return true
			}
		}
	}
	return false
}
