package agentbuilder

import (
	"fmt"
	"os"
	"strings"
)

// markerFormat is the HTML comment marker used to identify a custom-agent block.
// Example: <!-- architect-ai:custom-agent:my-skill -->
const markerFormat = "<!-- architect-ai:custom-agent:%s -->"

// InjectSDDReference appends (or replaces) a custom-agent reference block in the
// system prompt file at systemPromptPath.
//
// For SDDPhaseSupport mode the block declares that the skill supports an existing
// phase. For SDDNewPhase mode the block integrates it as a first-class new phase.
//
// The function is a no-op when agent.SDDConfig is nil or the mode is SDDStandalone.
func InjectSDDReference(agent *GeneratedAgent, systemPromptPath string) error {
	if agent == nil || agent.SDDConfig == nil || agent.SDDConfig.Mode == SDDStandalone {
		return nil
	}

	data, err := os.ReadFile(systemPromptPath)
	if err != nil {
		return fmt.Errorf("sdd inject: read %s: %w", systemPromptPath, err)
	}

	content := string(data)
	marker := fmt.Sprintf(markerFormat, agent.Name)
	block := buildSDDBlock(agent, marker)

	if strings.Contains(content, marker) {
		// Replace the existing block.
		content = replaceBlock(content, marker, block)
	} else {
		// Append the block.
		if !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "\n" + block + "\n"
	}

	if err := os.WriteFile(systemPromptPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("sdd inject: write %s: %w", systemPromptPath, err)
	}

	return nil
}

// buildSDDBlock returns the full marker+content block for the agent.
func buildSDDBlock(agent *GeneratedAgent, marker string) string {
	cfg := agent.SDDConfig
	var body string

	switch cfg.Mode {
	case SDDPhaseSupport:
		body = fmt.Sprintf(
			"## Custom Agent: %s (Phase Support)\n\n"+
				"This skill provides additional support for the `sdd-%s` phase.\n"+
				"When working on tasks related to `%s`, load the `%s` skill for enhanced guidance.\n\n"+
				"Trigger phrases: %s\n",
			agent.Title,
			cfg.TargetPhase,
			cfg.TargetPhase,
			agent.Name,
			agent.Trigger,
		)
	case SDDNewPhase:
		phaseName := cfg.PhaseName
		if phaseName == "" {
			phaseName = agent.Name
		}
		body = fmt.Sprintf(
			"## Custom Agent: %s (New SDD Phase)\n\n"+
				"This skill adds a new phase `%s` to the SDD dependency graph.\n"+
				"Load the `%s` skill when the orchestrator launches you for the `%s` phase.\n\n"+
				"Trigger phrases: %s\n",
			agent.Title,
			phaseName,
			agent.Name,
			phaseName,
			agent.Trigger,
		)
	default:
		body = fmt.Sprintf("## Custom Agent: %s\n\nTrigger: %s\n", agent.Title, agent.Trigger)
	}

	endMarker := fmt.Sprintf("<!-- /architect-ai:custom-agent:%s -->", agent.Name)
	return marker + "\n" + body + endMarker
}

// replaceBlock replaces the content between the opening marker and its matching
// closing marker with the new block string.
func replaceBlock(content, marker, newBlock string) string {
	endMarker := fmt.Sprintf("<!-- /architect-ai:custom-agent:%s -->", extractName(marker))

	start := strings.Index(content, marker)
	if start == -1 {
		return content + "\n" + newBlock
	}

	end := strings.Index(content[start:], endMarker)
	if end == -1 {
		// No closing marker: replace from start marker to end of line.
		lineEnd := strings.Index(content[start:], "\n")
		if lineEnd == -1 {
			return content[:start] + newBlock
		}
		return content[:start] + newBlock + content[start+lineEnd:]
	}

	replaceEnd := start + end + len(endMarker)
	return content[:start] + newBlock + content[replaceEnd:]
}

// extractName parses the agent name from a marker string.
// Example: "<!-- architect-ai:custom-agent:my-skill -->" → "my-skill"
func extractName(marker string) string {
	prefix := "<!-- architect-ai:custom-agent:"
	suffix := " -->"
	if strings.HasPrefix(marker, prefix) && strings.HasSuffix(marker, suffix) {
		return marker[len(prefix) : len(marker)-len(suffix)]
	}
	return ""
}

// ResolveMatchingSkills reads the markdown skill registry and returns the
// concatenated compact rules for skills that match the provided task paths.
// It implements the "Differential Context Injection" feature.
func ResolveMatchingSkills(registryPath string, taskPaths []string) (string, error) {
	data, err := os.ReadFile(registryPath)
	if err != nil {
		return "", fmt.Errorf("resolve skills: read registry: %w", err)
	}

	content := string(data)
	
	// 1. Identify matching skills based on triggers
	matches := findMatchingSkills(content, taskPaths)
	if len(matches) == 0 {
		return "", nil
	}

	// 2. Extract compact rules for matched skills
	return extractCompactRules(content, matches), nil
}

func findMatchingSkills(content string, paths []string) []string {
	var matches []string
	
	// Very simple parser for the Markdown tables in the registry.
	// We look for lines like: | trigger | skill-name | path |
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if !strings.HasPrefix(line, "|") || strings.Count(line, "|") < 3 {
			continue
		}
		
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		
		trigger := strings.TrimSpace(parts[1])
		skillName := strings.TrimSpace(parts[2])
		
		if trigger == "Trigger" || trigger == "---" {
			continue
		}

		if skillName == "" || skillName == "Skill" {
			continue
		}

		// Check if trigger matches any of the task paths
		if triggerMatches(trigger, paths) {
			matches = append(matches, skillName)
		}
	}
	
	return matches
}

func triggerMatches(trigger string, paths []string) bool {
	t := strings.ToLower(trigger)
	for _, p := range paths {
		pLow := strings.ToLower(p)

		// Match by file extension
		if (strings.Contains(t, "go ") || strings.Contains(t, "golang")) && strings.HasSuffix(pLow, ".go") {
			return true
		}
		if (strings.Contains(t, "typescript") || strings.Contains(t, "react") || strings.Contains(t, "javascript")) &&
			(strings.HasSuffix(pLow, ".ts") || strings.HasSuffix(pLow, ".tsx") || strings.HasSuffix(pLow, ".js") || strings.HasSuffix(pLow, ".jsx")) {
			return true
		}
		if strings.Contains(t, "python") && strings.HasSuffix(pLow, ".py") {
			return true
		}
		if (strings.Contains(t, "bash") || strings.Contains(t, "shell")) && strings.HasSuffix(pLow, ".sh") {
			return true
		}

		// Fallback: clean keyword containment
		tClean := strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
				return r
			}
			return ' '
		}, t)
		keywords := strings.Fields(tClean)
		for _, kw := range keywords {
			if len(kw) < 3 {
				continue
			}
			if strings.Contains(pLow, kw) {
				return true
			}
		}
	}
	return false
}

func extractCompactRules(content string, skillNames []string) string {
	const marker = "<!-- architect-ai:registry:compact-rules -->"
	startIdx := strings.Index(content, marker)
	if startIdx == -1 {
		return ""
	}

	rulesSection := content[startIdx+len(marker):]
	var sb strings.Builder

	for _, name := range skillNames {
		// Look for "### name" header
		header := "### " + name
		idx := strings.Index(rulesSection, header)
		if idx == -1 {
			continue
		}

		// Found the skill section. Extract until the next "###" or end of file.
		section := rulesSection[idx+len(header):]
		nextIdx := strings.Index(section, "###")
		if nextIdx != -1 {
			section = section[:nextIdx]
		}
		
		sb.WriteString(header + "\n")
		sb.WriteString(strings.TrimSpace(section) + "\n\n")
	}

	return strings.TrimSpace(sb.String())
}
