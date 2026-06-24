package protoshell

import (
	"strings"
)

type BudgetLayer struct {
	ProjectStandards int
}

type Skill struct {
	Name            string
	Content         string
	TriggerKeywords []string
}

type SkillRegistry struct {
	Skills []Skill
}

type AssemblyContext struct {
	PhaseName    string
	TaskKeywords []string
	Technologies []string
	BudgetConfig BudgetLayer
}

func AssembleContext(registry SkillRegistry, ctx AssemblyContext) (string, int, error) {
	var injected []string
	var totalTokens int

	for _, skill := range registry.Skills {
		if skillMatches(skill, ctx) {
			content := skill.Content
			tokens := estimateTokens(content)
			if totalTokens+tokens > ctx.BudgetConfig.ProjectStandards {
				break
			}
			injected = append(injected, content)
			totalTokens += tokens
		}
	}
	return strings.Join(injected, "\n\n"), totalTokens, nil
}

func skillMatches(skill Skill, ctx AssemblyContext) bool {
	for _, kw := range skill.TriggerKeywords {
		if containsAny(ctx.PhaseName, ctx.TaskKeywords, ctx.Technologies, kw) {
			return true
		}
	}
	return false
}

func containsAny(phaseName string, taskKeywords []string, technologies []string, keyword string) bool {
	keyword = strings.ToLower(keyword)
	
	if strings.Contains(strings.ToLower(phaseName), keyword) {
		return true
	}
	
	for _, kw := range taskKeywords {
		if strings.Contains(strings.ToLower(kw), keyword) {
			return true
		}
	}
	
	for _, tech := range technologies {
		if strings.Contains(strings.ToLower(tech), keyword) {
			return true
		}
	}
	
	return false
}

func estimateTokens(content string) int {
	return len(content) / 4
}
