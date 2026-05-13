package skills

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	Score   int
}

// ConventionRef holds a reference to a project convention file.
type ConventionRef struct {
	Path  string
	Notes string
}

// TriggerEntry holds trigger information parsed from the Quick Index table.
type TriggerEntry struct {
	Name    string
	Kind    string
	Trigger string
}

// fileExtMap maps common extensions to canonical language/skill names.
var fileExtMap = map[string][]string{
	".go":   {"go", "golang"},
	".ts":   {"typescript", "ts"},
	".tsx":  {"typescript", "ts", "react", "tsx"},
	".jsx":  {"react", "jsx"},
	".js":   {"javascript", "js"},
	".py":   {"python"},
	".rs":   {"rust"},
	".rb":   {"ruby"},
	".java": {"java"},
	".kt":   {"kotlin"},
	".swift": {"swift"},
	".c":    {"c"},
	".h":    {"c"},
	".cpp":  {"cpp", "c++"},
	".hpp":  {"cpp", "c++"},
	".scala": {"scala"},
	".sql":  {"sql"},
	".sh":   {"shell", "bash"},
	".bash": {"shell", "bash"},
	".yaml": {"yaml"},
	".yml":  {"yaml"},
	".json": {"json"},
	".md":   {"markdown"},
	".css":  {"css"},
	".html": {"html"},
	".vue":  {"vue"},
	".svelte": {"svelte"},
}

// toolConditionalSkills lists skills that should only be injected
// when their corresponding tool is available.
var toolConditionalSkills = map[string]string{
	"mcp-notebooklm-orchestrator": "notebooklm",
	"mcp-context7-skill":          "context7",
}

// IsToolConditionalSkill returns true and the required tool name if
// the skill should only be injected when its tool is available.
func IsToolConditionalSkill(skillName string) (string, bool) {
	tool, ok := toolConditionalSkills[skillName]
	return tool, ok
}

// ResolveStandardsForTask searches the project registry (local or memory)
// to find skills and conventions relevant to the given task and modified paths.
// availableTools lists tool names (e.g. "notebooklm", "context7") that are
// currently available; tool-conditional skills whose tool is not present are skipped.
func ResolveStandardsForTask(projectRoot, task, phase string, paths []string, availableTools []string) (ResolvedStandards, error) {
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

	availSet := make(map[string]bool, len(availableTools))
	for _, t := range availableTools {
		availSet[strings.ToLower(strings.TrimSpace(t))] = true
	}

	// 1. Parse trigger entries from the Quick Index table
	triggers := parseQuickIndexTable(content)

	// 2. Extract Compact Rules
	rs.Rules = extractCompactRules(content, task, paths, triggers, availSet)

	// 3. Extract Project Conventions
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

func extractCompactRules(registryContent, task string, paths []string, triggers map[string]*TriggerEntry, availTools map[string]bool) []ResolvedRule {
	rules := make([]ResolvedRule, 0)

	section := extractSection(registryContent, "## Compact Rules")
	if section == "" {
		return rules
	}

	lines := strings.Split(section, "\n")
	var currentSkill string
	var currentContent strings.Builder

	finalizeSkill := func() {
		if currentSkill == "" {
			return
		}
		content := strings.TrimSpace(currentContent.String())
		if content != "" {
			score := matchSkill(currentSkill, task, paths, triggers)
			if score > 0 {
				if reqTool, isCond := IsToolConditionalSkill(currentSkill); !isCond || availTools[reqTool] {
					rules = append(rules, ResolvedRule{
						Skill:   currentSkill,
						Content: content,
						Score:   score,
					})
				}
			}
		}
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "### ") {
			heading := strings.TrimPrefix(line, "### ")
			heading = strings.TrimPrefix(heading, "Skill: ")
			heading = strings.TrimSpace(heading)

			if _, isSkill := triggers[heading]; isSkill {
				finalizeSkill()
				currentSkill = heading
				currentContent.Reset()
			}
		} else if currentSkill != "" {
			currentContent.WriteString(line + "\n")
		}
	}

	finalizeSkill()

	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Score > rules[j].Score
	})

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

	rest := content[start+len(title):]
	end := strings.Index(rest, "\n## ")
	if end == -1 {
		return rest
	}
	return rest[:end]
}

func parseQuickIndexTable(registryContent string) map[string]*TriggerEntry {
	entries := make(map[string]*TriggerEntry)
	section := extractSection(registryContent, "## Quick Index")
	if section == "" {
		return entries
	}

	lines := strings.Split(section, "\n")
	inTable := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "|") && strings.Contains(line, "Skill") {
			inTable = true
			continue
		}
		if !inTable || !strings.HasPrefix(line, "|") {
			continue
		}
		if strings.Contains(line, "---") {
			continue
		}

		cols := splitTableRow(line)
		if len(cols) < 4 {
			continue
		}

		entry := &TriggerEntry{
			Name:    strings.TrimSpace(cols[0]),
			Kind:    strings.TrimSpace(cols[1]),
			Trigger: strings.TrimSpace(cols[2]),
		}
		entries[entry.Name] = entry
	}
	return entries
}

func splitTableRow(line string) []string {
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	result := make([]string, len(parts))
	for i, p := range parts {
		result[i] = strings.TrimSpace(p)
	}
	return result
}

func matchSkill(skillName, task string, paths []string, triggers map[string]*TriggerEntry) int {
	trigger := ""
	if entry, ok := triggers[skillName]; ok && entry.Trigger != "" {
		trigger = entry.Trigger
	}

	skillLower := strings.ToLower(skillName)
	taskLower := strings.ToLower(task)
	triggerLower := strings.ToLower(trigger)

	triggerScore := 0
	nameScore := 0
	fileScore := 0

	if triggerLower != "" && taskLower != "" {
		triggerWords := strings.Fields(triggerLower)
		for _, tw := range triggerWords {
			tw = strings.Trim(tw, `".,;:!?`)
			if len(tw) < 3 {
				continue
			}
			if strings.Contains(taskLower, tw) {
				triggerScore = 100
				break
			}
		}
	}

	if strings.Contains(taskLower, skillLower) || (triggerLower != "" && strings.Contains(triggerLower, skillLower)) {
		if triggerScore == 0 {
			nameScore = 50
		}
	}

	for _, path := range paths {
		ext := strings.ToLower(filepath.Ext(path))
		labels, ok := fileExtMap[ext]
		if !ok {
			continue
		}
		for _, label := range labels {
			if label == skillLower || strings.Contains(triggerLower, label) {
				fileScore = 30
				break
			}
		}
		if fileScore > 0 {
			break
		}
	}

	total := triggerScore + nameScore + fileScore
	if triggerScore > 0 && nameScore > 0 {
		total = triggerScore
	}

	return total
}
